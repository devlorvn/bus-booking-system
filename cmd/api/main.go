package main

import (
	"booking-system/configs"
	httpBooking "booking-system/internal/booking/delivery/http"
	"booking-system/internal/booking/delivery/ws"
	lockseatUC "booking-system/internal/booking/usecase/lock_seat"
	httpBusDelivery "booking-system/internal/bus/delivery/http"
	httpBusHandler "booking-system/internal/bus/delivery/http/handler"
	"booking-system/internal/provider"
	"booking-system/pkg/kafka"
	"booking-system/pkg/redis"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/middleware"
	bookingpb "booking-system/proto/booking/v1"
	buspb "booking-system/proto/bus/v1"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := configs.LoadConfig()

	if cfg.Mode == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	redisClient := redis.NewClient(&cfg.Redis)

	// Web socket hub
	hub := ws.NewHub()

	// Create worker context
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		hub.Run(workerCtx)
	}()

	h := ws.NewHandler(hub)

	eventPublisher := provider.NewWSEventPublisher(hub)

	busConn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer busConn.Close()

	busGrpcClient := buspb.NewBusServiceClient(busConn)

	busProvider := provider.NewBusProvider(busGrpcClient)
	seatLockRepo := redis.NewLockSeatRepository(redisClient)

	lockSeatUsecase := lockseatUC.New(
		busProvider,
		seatLockRepo,
		eventPublisher,
	)

	bookingConn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer bookingConn.Close()

	bookingGrpcClient := bookingpb.NewBookingServiceClient(bookingConn)

	// Initial reader
	kafkaReader := kafka.NewReader(
		cfg.Kafka.Brokers,
		constants.NotificationTopic,
		constants.APIGatewayWebSocketPollGroup,
	)
	defer kafkaReader.Close()

	wsConsumer := ws.NewWsConsumer(
		kafkaReader,
		hub,
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := wsConsumer.Consume(workerCtx); err != nil && err != context.Canceled {
			log.Printf("WS consumer: stopped with error: %v", err)
		}
	}()

	r := gin.Default()

	api := r.Group("/api")

	r.Static("/ui", "./web")

	api.Use(middleware.RequestIdMiddleware())

	api.Use(middleware.ErrorHandler())

	httpBusDelivery.RegiserBusRouter(api, httpBusHandler.NewBusHandler(busGrpcClient))
	httpBooking.RegisterRoutes(api, httpBooking.NewBookingHandler(lockSeatUsecase, bookingGrpcClient))

	r.GET("/ws/buses/:id", h.Handle)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	// 1. Shutdown HTTP server first, allowing 5 seconds to finish active requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server Shutdown failed: %v", err)
	} else {
		log.Println("HTTP server shut down successfully")
	}

	// 2. Stop workers
	cancelWorkers()

	// Wait for all background workers to stop (with a timeout of 5 seconds)
	workersDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(workersDone)
	}()

	select {
	case <-workersDone:
		log.Println("All background workers stopped successfully")
	case <-time.After(5 * time.Second):
		log.Println("Timeout waiting for background workers to stop")
	}

	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing Redis client: %v", err)
	} else {
		log.Println("Redis client closed successfully")
	}

	log.Println("Server exiting")
}
