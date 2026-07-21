package main

import (
	"booking-system/configs"
	bookinggrpc "booking-system/internal/booking/delivery/grpc"
	"booking-system/internal/booking/delivery/worker"
	bookingService "booking-system/internal/booking/service"
	confirmbookingUC "booking-system/internal/booking/usecase/confirm_booking"
	expirebookinguc "booking-system/internal/booking/usecase/expire_booking"
	handlepaymentUC "booking-system/internal/booking/usecase/handle_payment"
	lockseatUC "booking-system/internal/booking/usecase/lock_seat"
	processsagaresultUC "booking-system/internal/booking/usecase/process_saga_result"
	"booking-system/internal/provider"
	"booking-system/pkg/database"
	"booking-system/pkg/kafka"
	"booking-system/pkg/postgres"
	postgresRepository "booking-system/pkg/postgres/repository"
	"booking-system/pkg/redis"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"booking-system/pkg/shared/metrics"
	bookingpb "booking-system/proto/booking/v1"
	buspb "booking-system/proto/bus/v1"
	userpb "booking-system/proto/user/v1"
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	slog.Info("Mode", slog.String("mode", config.Mode))
	if config.Mode != "production" {
		slog.Info("Migrating database...")
		err = postgres.AutoMigrate(db)
		if err != nil {
			slog.Error("failed to migrate database: ", slog.String("error", err.Error()))
		}
	}

	txManager := database.NewTransaction(db)

	// initial redis
	redisClient := redis.NewClient(&config.Redis)
	defer redisClient.Close()

	// Initial repository and adapters
	bookingRepo := postgresRepository.NewBookingRepository(db)
	bookingSeatRepo := postgresRepository.NewBookingSeatRepository(db)
	outboxRepo := postgresRepository.NewOutboxRepository(db)

	bookingLockRepo := redis.NewLockBookingRepository(redisClient)
	seatLockRepo := redis.NewLockSeatRepository(redisClient)
	pricingService := bookingService.NewPricingService()

	// Connecting gRPC Bus Service (port 50051)
	busConn, err := grpc.NewClient(config.BusServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to bus service: ", slog.String("error", err.Error()))
	}
	defer busConn.Close()

	busGrpcClient := buspb.NewBusServiceClient(busConn)
	busProvider := provider.NewBusProvider(busGrpcClient) // connect to Bus Service via gRPC

	// Connecting gRPC User service (port 50053)
	userConn, err := grpc.NewClient(config.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to user service: ", slog.String("error", err.Error()))
	}
	defer userConn.Close()

	userGrpcClient := userpb.NewUserServiceClient(userConn)
	userProvider := provider.NewUserProvider(userGrpcClient) // connect to User Service via gRPC

	// setting kafka publishers and consumers
	kafka.CreateTopicIfNotExist(config.Kafka.Brokers, constants.BookingTopic, 1, 1)
	kafka.CreateTopicIfNotExist(config.Kafka.Brokers, constants.PaymentTopic, 1, 1)
	kafka.CreateTopicIfNotExist(config.Kafka.Brokers, constants.NotificationTopic, 1, 1)
	kafka.CreateTopicIfNotExist(config.Kafka.Brokers, constants.PaymentDLQTopic, 1, 1)

	// Booking created publisher
	kafkaWriter := kafka.NewWriter(config.Kafka.Brokers, constants.BookingTopic)
	defer kafkaWriter.Close()

	kafkaDLQWriter := kafka.NewWriter(config.Kafka.Brokers, constants.PaymentDLQTopic)
	defer kafkaDLQWriter.Close()

	bookingPublisher := events.NewKafkaBookingPublisher(kafkaWriter)

	kafkaWsWriter := kafka.NewWriter(config.Kafka.Brokers, constants.NotificationTopic)
	defer kafkaWsWriter.Close()
	seatPublisher := events.NewKafkaSeatPublisher(kafkaWsWriter)

	// Initial ConfirmBookingUsecase
	confirmBookingUsecase := confirmbookingUC.New(
		bookingRepo,
		bookingSeatRepo,
		userProvider,
		seatLockRepo,
		busProvider,
		pricingService,
		txManager,
		bookingLockRepo,
		outboxRepo,
	)

	// Initial handle payment usecase
	handlerPaymentUsecase := handlepaymentUC.New(
		bookingRepo,
		bookingSeatRepo,
		busProvider,
		bookingLockRepo,
		outboxRepo,
		txManager,
	)

	// Initial expire booking usecase
	expireBookingUsecase := expirebookinguc.New(
		bookingRepo,
		bookingSeatRepo,
		busProvider,
		outboxRepo,
		txManager,
	)

	lockSeatUsecase := lockseatUC.New(
		busProvider,
		seatLockRepo,
		seatPublisher,
	)

	// Initial process saga result usecase
	processSagaResultUsecase := processsagaresultUC.New(
		bookingRepo,
		bookingLockRepo,
		outboxRepo,
		txManager,
	)

	// Initial lock expiration worker
	lockExpirationWorker := worker.NewLockExpirationWorker(
		redisClient,
		seatPublisher,
		expireBookingUsecase,
	)

	// Initial payment consumer worker
	paymentReader := kafka.NewReader(config.Kafka.Brokers, constants.PaymentTopic, constants.BookingServicePollGroup)
	defer paymentReader.Close()

	paymentWorker := worker.NewPaymentWorker(paymentReader, kafkaDLQWriter, handlerPaymentUsecase)

	sagaResultReader := kafka.NewReader(config.Kafka.Brokers, constants.BookingTopic, "booking-service-saga-result-group")
	defer sagaResultReader.Close()

	sagaResultWorker := worker.NewSagaResultWorker(sagaResultReader, processSagaResultUsecase)

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	// Initial outbox worker
	outboxWorker := worker.NewOutboxWorker(
		outboxRepo,
		bookingPublisher,
		500*time.Millisecond,
	)

	// Running payment consumer worker
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := paymentWorker.Start(workerCtx); err != nil {
			slog.Error("payment worker failed: ", slog.String("error", err.Error()))
		}
	}()

	// Running outbox worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := outboxWorker.Start(workerCtx); err != nil {
			if err != context.Canceled {
				slog.Error("outbox worker failed: ", slog.String("error", err.Error()))
			}
		}
	}()

	//Running lock expiration worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := lockExpirationWorker.Start(workerCtx); err != nil {
			slog.Error("lock expiration worker failed: ", slog.String("error", err.Error()))
		}
	}()

	// Running saga result worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := sagaResultWorker.Start(workerCtx); err != nil {
			slog.Error("saga result worker failed: ", slog.String("error", err.Error()))
		}
	}()

	// Start metrics server
	metricsPort := config.MetricsPort
	if metricsPort == "9090" {
		metricsPort = "9092"
	}
	metrics.StartMetricsServer(metricsPort)

	// Running Booking gRPC server on port 50052
	lis, err := net.Listen("tcp", ":50052")
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
	healthServer.SetServingStatus("booking.v1.BookingService", grpc_health_v1.HealthCheckResponse_SERVING)

	bookingGrpcServer := bookinggrpc.NewBookingGRPCServer(confirmBookingUsecase, lockSeatUsecase)
	bookingpb.RegisterBookingServiceServer(grpcServer, bookingGrpcServer)

	// Register reflection for debuging
	reflection.Register(grpcServer)
	go func() {
		slog.Info("Booking gRPC server started on port %s", slog.String("port", lis.Addr().String()))
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("failed to serve: ", slog.String("error", err.Error()))
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down Booking gRPC server...")
	grpcServer.GracefulStop()

	cancelWorker()
	workerDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(workerDone)
	}()

	select {
	case <-workerDone:
		slog.Info("Payment worker exited")
		slog.Info("Lock expiration worker exited")
		slog.Info("Outbox worker exited")
		slog.Info("Saga result worker exited")
	case <-time.After(5 * time.Second):
		slog.Error("Workers timeout during exit")
	}

	slog.Info("Booking gRPC server exited")
}
