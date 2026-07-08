package main

import (
	"booking-system/configs"
	busgrpc "booking-system/internal/bus/delivery/grpc"
	"booking-system/internal/bus/usecase"
	"booking-system/pkg/database"
	"booking-system/pkg/postgres"
	postgresRepository "booking-system/pkg/postgres/repository"
	"booking-system/pkg/redis"
	buspb "booking-system/proto/bus/v1"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

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

	// initial repository
	busRepo := postgresRepository.NewBusRepository(db)
	seatRepo := postgresRepository.NewSeatRepository(db)
	seatLockRepo := redis.NewLockSeatRepository(redisClient)

	// initial usecase
	seatUsecase := usecase.NewSeatUsecase(seatRepo, txManager)
	busUsecase := usecase.NewBusUsecase(busRepo, seatUsecase, seatLockRepo, txManager)

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
	log.Println("Server exited")
}
