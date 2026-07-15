package main

import (
	"booking-system/configs"
	usergrpc "booking-system/internal/user/delivery/grpc"
	"booking-system/internal/user/usecase"
	"booking-system/pkg/database"
	"booking-system/pkg/postgres"
	postgresRepository "booking-system/pkg/postgres/repository"
	userpb "booking-system/proto/user/v1"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg := configs.LoadConfig()

	db, err := postgres.NewPostgres(&cfg.Database)
	if err != nil {
		slog.Error("failed to connect database: ", slog.String("error", err.Error()))
	}
	txManager := database.NewTransaction(db)

	userRepo := postgresRepository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, txManager)

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		slog.Error("failed to listen: ", slog.String("error", err.Error()))
	}
	grpcServer := grpc.NewServer()

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	// SetServingStatus is not thread-safe. SetServingStatusBlocking should be used for initialization.
	healthServer.SetServingStatus("user.v1.UserService", grpc_health_v1.HealthCheckResponse_SERVING)

	userGrpcServer := usergrpc.NewUserGRPCServer(userUsecase)
	userpb.RegisterUserServiceServer(grpcServer, userGrpcServer)

	reflection.Register(grpcServer)

	go func() {
		slog.Info("gRPC server started on port %s", slog.String("port", lis.Addr().String()))
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("failed to serve: ", slog.String("error", err.Error()))
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	grpcServer.GracefulStop()
	slog.Info("shutting down server")

}
