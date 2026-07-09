package main

import (
	"booking-system/configs"
	bookinggrpc "booking-system/internal/booking/delivery/grpc"
	"booking-system/internal/booking/event"
	bookingService "booking-system/internal/booking/service"
	confirmbookingUC "booking-system/internal/booking/usecase/confirm_booking"
	"booking-system/internal/provider"
	"booking-system/pkg/database"
	"booking-system/pkg/postgres"
	postgresRepository "booking-system/pkg/postgres/repository"
	"booking-system/pkg/redis"
	bookingpb "booking-system/proto/booking/v1"
	buspb "booking-system/proto/bus/v1"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

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
	userRepo := postgresRepository.NewUserRepository(db)
	bookingRepoAdapter := &provider.BookingRepoAdapter{Repo: bookingRepo}
	userPortAdapter := &provider.UserPortAdapter{Repo: userRepo}
	bookingLockRepo := redis.NewLockBookingRepository(redisClient)
	seatLockRepo := redis.NewLockSeatRepository(redisClient)
	// 1. Kết nối gRPC sang Bus Service (cổng 50051)
	busConn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to connect to bus service: ", err)
	}
	defer busConn.Close()
	busGrpcClient := buspb.NewBusServiceClient(busConn)
	busProvider := provider.NewBusProvider(busGrpcClient) // Giao tiếp với Bus Service qua gRPC
	pricingService := bookingService.NewPricingService()
	paymentBus := event.NewPaymentEventBus()
	paymentPublisher := event.NewPaymentPublisher(paymentBus)
	// 2. Khởi tạo ConfirmBookingUsecase
	confirmBookingUsecase := confirmbookingUC.New(
		bookingRepoAdapter,
		bookingSeatRepo,
		userPortAdapter,
		seatLockRepo,
		busProvider,
		pricingService,
		txManager,
		bookingLockRepo,
		paymentPublisher,
	)
	// 3. Khởi chạy gRPC Server của Booking trên cổng 50052
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}
	grpcServer := grpc.NewServer()
	bookingGrpcServer := bookinggrpc.NewBookingGRPCServer(confirmBookingUsecase)
	bookingpb.RegisterBookingServiceServer(grpcServer, bookingGrpcServer)
	// Đăng ký reflection để debug gRPC tiện lợi
	reflection.Register(grpcServer)
	go func() {
		log.Printf("Booking gRPC server started on port %s", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve: ", err)
		}
	}()
	// Xử lý Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Booking gRPC server...")
	grpcServer.GracefulStop()
	log.Println("Booking gRPC server exited")
}
