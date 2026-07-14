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
	"booking-system/internal/provider"
	"booking-system/pkg/database"
	"booking-system/pkg/kafka"
	"booking-system/pkg/postgres"
	postgresRepository "booking-system/pkg/postgres/repository"
	"booking-system/pkg/redis"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	bookingpb "booking-system/proto/booking/v1"
	buspb "booking-system/proto/bus/v1"
	userpb "booking-system/proto/user/v1"
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	// Initial repository and adapters
	bookingRepo := postgresRepository.NewBookingRepository(db)
	bookingSeatRepo := postgresRepository.NewBookingSeatRepository(db)

	bookingRepoAdapter := &provider.BookingRepoAdapter{Repo: bookingRepo}
	bookingLockRepo := redis.NewLockBookingRepository(redisClient)
	seatLockRepo := redis.NewLockSeatRepository(redisClient)
	pricingService := bookingService.NewPricingService()

	// Connecting gRPC Bus Service (port 50051)
	busConn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to connect to bus service: ", err)
	}
	defer busConn.Close()

	busGrpcClient := buspb.NewBusServiceClient(busConn)
	busProvider := provider.NewBusProvider(busGrpcClient) // connect to Bus Service via gRPC

	// Connecting gRPC User service (port 50053)
	userConn, err := grpc.NewClient("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to connect to user service: ", err)
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

	paymentPublisher := events.NewKafkaPaymentPublisher(kafkaWriter)

	// ws event publisher
	kafkaWsWriter := kafka.NewWriter(config.Kafka.Brokers, constants.NotificationTopic)
	defer kafkaWsWriter.Close()
	wsPublisher := events.NewKafkaPublisher(kafkaWsWriter)

	// Initial ConfirmBookingUsecase
	confirmBookingUsecase := confirmbookingUC.New(
		bookingRepoAdapter,
		bookingSeatRepo,
		userProvider,
		seatLockRepo,
		busProvider,
		pricingService,
		txManager,
		bookingLockRepo,
		paymentPublisher,
	)

	// Initial handle payment usecase
	handlerPaymentUsecase := handlepaymentUC.New(
		bookingRepoAdapter,
		busProvider,
		bookingLockRepo,
		wsPublisher,
		txManager,
	)

	// Initial expire booking usecase
	expireBookingUsecase := expirebookinguc.New(
		bookingRepoAdapter,
		busProvider,
		wsPublisher,
		txManager,
	)

	// Initial lock expiration worker
	lockExpirationWorker := worker.NewLockExpirationWorker(
		redisClient,
		wsPublisher,
		expireBookingUsecase,
	)

	lockSeatUsecase := lockseatUC.New(
		busProvider,
		seatLockRepo,
		wsPublisher,
	)

	// Initial payment consumer worker
	kafkaReader := kafka.NewReader(config.Kafka.Brokers, constants.PaymentTopic, constants.BookingServicePollGroup)
	defer kafkaReader.Close()

	paymentWorker := worker.NewPaymentWorker(kafkaReader, kafkaDLQWriter, handlerPaymentUsecase)
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	// Running payment consumer worker
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := paymentWorker.Start(workerCtx); err != nil {
			log.Fatal("payment worker failed: ", err)
		}
	}()

	//Running lock expiration worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := lockExpirationWorker.Start(workerCtx); err != nil {
			log.Fatal("lock expiration worker failed: ", err)
		}
	}()

	// Running Booking gRPC server on port 50052
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}
	grpcServer := grpc.NewServer()
	bookingGrpcServer := bookinggrpc.NewBookingGRPCServer(confirmBookingUsecase, lockSeatUsecase)
	bookingpb.RegisterBookingServiceServer(grpcServer, bookingGrpcServer)

	// Register reflection for debuging
	reflection.Register(grpcServer)
	go func() {
		log.Printf("Booking gRPC server started on port %s", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve: ", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Booking gRPC server...")
	grpcServer.GracefulStop()

	cancelWorker()
	workerDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(workerDone)
	}()

	select {
	case <-workerDone:
		log.Println("Payment worker exited")
		log.Println("Lock expiration worker exited")
	case <-time.After(5 * time.Second):
		log.Println("Payment worker timeout")
		log.Println("Lock expiration worker timeout")
	}

	log.Println("Booking gRPC server exited")
}
