package main

import (
	"booking-system/configs"
	usergrpc "booking-system/internal/user/delivery/grpc"
	"booking-system/internal/user/usecase"
	"booking-system/pkg/database"
	"booking-system/pkg/postgres"
	postgresRepository "booking-system/pkg/postgres/repository"
	userpb "booking-system/proto/user/v1"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := configs.LoadConfig()

	db, err := postgres.NewPostgres(&cfg.Database)
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}
	txManager := database.NewTransaction(db)

	userRepo := postgresRepository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, txManager)

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}
	grpcServer := grpc.NewServer()
	userGrpcServer := usergrpc.NewUserGRPCServer(userUsecase)
	userpb.RegisterUserServiceServer(grpcServer, userGrpcServer)

	reflection.Register(grpcServer)

	go func() {
		log.Printf("gRPC server started on port %s", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve: ", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	grpcServer.GracefulStop()
	log.Println("shutting down server")

}
