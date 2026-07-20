package main

import (
	"booking-system/configs"
	busgrpc "booking-system/internal/bus/delivery/grpc"
	"booking-system/internal/bus/delivery/worker"
	"booking-system/internal/bus/usecase"
	"booking-system/pkg/database"
	"booking-system/pkg/kafka"
	"booking-system/pkg/postgres"
	postgresRepository "booking-system/pkg/postgres/repository"
	"booking-system/pkg/redis"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/metrics"
	buspb "booking-system/proto/bus/v1"
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	config := configs.LoadConfig()
	// initial database
	db, err := postgres.NewPostgres(&config.Database)
	if err != nil {
		slog.Error("failed to connect database: ", slog.String("error", err.Error()))
		os.Exit(1)
	}
	txManager := database.NewTransaction(db)

	// initial redis
	redisClient := redis.NewClient(&config.Redis)
	defer redisClient.Close()

	kafka.CreateTopicIfNotExist(config.Kafka.Brokers, constants.BookingTopic, 1, 1)

	// initial repository
	busRepo := postgresRepository.NewBusRepository(db)
	seatRepo := postgresRepository.NewSeatRepository(db)
	seatLockRepo := redis.NewLockSeatRepository(redisClient)

	// initial usecase
	seatUsecase := usecase.NewSeatUsecase(seatRepo, txManager)
	busUsecase := usecase.NewBusUsecase(busRepo, seatUsecase, seatLockRepo, txManager)

	// initial kafka publishers/consumers
	kafkaWriter := kafka.NewWriter(config.Kafka.Brokers, constants.BookingTopic)
	defer kafkaWriter.Close()

	// initial booking events worker
	bookingReader := kafka.NewReader(
		config.Kafka.Brokers,
		constants.BookingTopic,
		constants.BusServicePollGroup,
	)
	defer bookingReader.Close()

	bookingWorker := worker.NewBookingWorker(bookingReader, kafkaWriter, seatUsecase)

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := bookingWorker.Start(workerCtx); err != nil {
			slog.Error("booking worker failed: ", slog.String("error", err.Error()))
		}
	}()

	// Start metrics server
	metrics.StartMetricsServer(config.MetricsPort)

	// initial grpc server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("failed to listen: ", slog.String("error", err.Error()))
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(metrics.UnaryServerInterceptor()),
	)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	// SetServingStatus is not thread-safe. SetServingStatusBlocking should be used for initialization.
	healthServer.SetServingStatus("bus.v1.BusService", grpc_health_v1.HealthCheckResponse_SERVING)

	busGrpcServer := busgrpc.NewBusGRPCServer(busUsecase, seatUsecase)
	buspb.RegisterBusServiceServer(grpcServer, busGrpcServer)

	// register reflection
	reflection.Register(grpcServer)

	go func() {
		slog.Info("grpc server started on port %s", slog.String("port", lis.Addr().String()))

		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("failed to serve: ", slog.String("error", err.Error()))
		}
	}()

	// handle signal
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// wait for signal
	<-quit

	slog.Info("Shutting down server...")

	// close grpc server
	grpcServer.GracefulStop()

	cancelWorker()

	// wait for worker to finish
	wg.Wait()
	workerDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(workerDone)

	}()

	select {
	case <-workerDone:
		slog.Info("Booking worker exited")
	case <-time.After(5 * time.Second):
		slog.Error("Worker timeout during exit")
	}
	slog.Info("Server exited")
}
