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
	buspb "booking-system/proto/bus/v1"
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config := configs.LoadConfig()
	// initial database
	db, err := postgres.NewPostgres(&config.Database)
	if err != nil {
		log.Fatal("failed to connect database: ", err)
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

	// initial booking failed worker
	kafkaReader := kafka.NewReader(
		config.Brokers,
		constants.BookingTopic,
		constants.BookingServicePollGroup,
	)
	defer kafkaReader.Close()

	bookingFailedWorker := worker.NewBookingCancelledWorker(kafkaReader, seatUsecase)
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := bookingFailedWorker.Start(workerCtx); err != nil {
			log.Fatal("booking cancelled worker failed: ", err)
		}
	}()

	// initial grpc server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}

	grpcServer := grpc.NewServer()
	busGrpcServer := busgrpc.NewBusGRPCServer(busUsecase, seatUsecase)
	buspb.RegisterBusServiceServer(grpcServer, busGrpcServer)

	// register reflection
	reflection.Register(grpcServer)

	go func() {
		log.Printf("grpc server started on port %s", lis.Addr().String())

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve: ", err)
		}
	}()

	// handle signal
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// wait for signal
	<-quit

	log.Println("Shutting down server...")

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
		log.Println("Booking cancelled worker exited")
	case <-time.After(5 * time.Second):
		log.Println("Booking cancelled worker timeout")
	}
	log.Println("Server exited")
}
