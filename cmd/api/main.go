package main

import (
	"booking-system/configs"
	httpBooking "booking-system/internal/booking/delivery/http"
	"booking-system/internal/booking/delivery/worker"
	"booking-system/internal/booking/delivery/ws"
	"booking-system/internal/booking/event"
	bookingService "booking-system/internal/booking/service"
	confirmbookingUC "booking-system/internal/booking/usecase/confirm_booking"
	expirebookingUC "booking-system/internal/booking/usecase/expire_booking"
	handlepayment "booking-system/internal/booking/usecase/handle_payment"
	lockseatUC "booking-system/internal/booking/usecase/lock_seat"
	httpBusDelivery "booking-system/internal/bus/delivery/http"
	httpBusHandler "booking-system/internal/bus/delivery/http/handler"
	"booking-system/internal/bus/usecase"
	"booking-system/internal/provider"
	"booking-system/pkg/database"
	"booking-system/pkg/postgres"
	postgresRepository "booking-system/pkg/postgres/repository"
	"booking-system/pkg/redis"
	"booking-system/pkg/shared/middleware"
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
	db, err := postgres.NewPostgres(&cfg.Database)
	if err != nil {
		panic(err)
	}

	if cfg.Mode == "development" {
		postgres.AutoMigrate(db)
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	txManager := database.NewTransaction(db)

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

	busRepo := postgresRepository.NewBusRepository(db)
	seatRepo := postgresRepository.NewSeatRepository(db)

	grpcConn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer grpcConn.Close()

	busGrpcClient := buspb.NewBusServiceClient(grpcConn)

	busProvider := provider.NewBusProvider(busGrpcClient)
	seatLockRepo := redis.NewLockSeatRepository(redisClient)
	bookingLockRepo := redis.NewLockBookingRepository(redisClient)

	lockSeatUsecase := lockseatUC.New(
		busProvider,
		seatLockRepo,
		eventPublisher,
	)

	bookingRepo := postgresRepository.NewBookingRepository(db)
	bookingSeatRepo := postgresRepository.NewBookingSeatRepository(db)
	userRepo := postgresRepository.NewUserRepository(db)

	bookingRepoAdapter := &provider.BookingRepoAdapter{Repo: bookingRepo}
	userPortAdapter := &provider.UserPortAdapter{Repo: userRepo}
	seatPortAdapter := &provider.SeatPortAdapter{Repo: seatRepo}
	busPortAdapter := &provider.BusPortAdapter{Repo: busRepo}
	pricingService := bookingService.NewPricingService()
	paymentBus := event.NewPaymentEventBus()
	paymentProcessor := bookingService.NewFakePaymentProcessor(paymentBus)

	handlePaymentUsecase := handlepayment.New(
		bookingRepoAdapter,
		seatPortAdapter,
		busPortAdapter,
		bookingLockRepo,
		eventPublisher,
		txManager,
	)
	paymentPublisher := event.NewPaymentPublisher(paymentBus)
	paymentWorker := worker.NewPaymentWorker(paymentBus, paymentProcessor, handlePaymentUsecase)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := paymentWorker.Start(workerCtx); err != nil && err != context.Canceled {
			log.Printf("Payment worker stopped with error: %v", err)
		}
	}()

	// Start DLQWorker as well to handle failed payment tasks and prevent blocking on DLQ channel
	dlqWorker := worker.NewDLQWorker(paymentBus)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := dlqWorker.Start(workerCtx); err != nil && err != context.Canceled {
			log.Printf("DLQ worker stopped with error: %v", err)
		}
	}()

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

	expireBookingUsecase := expirebookingUC.New(
		bookingRepoAdapter,
		seatPortAdapter,
		txManager,
	)

	lockExpirationWorker := worker.NewLockExpirationWorker(redisClient, eventPublisher, expireBookingUsecase)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := lockExpirationWorker.Start(workerCtx); err != nil && err != context.Canceled {
			log.Printf("Lock expiration worker stopped with error: %v", err)
		}
	}()

	seatUsecase := usecase.NewSeatUsecase(seatRepo, txManager)
	busUsecase := usecase.NewBusUsecase(busRepo, seatUsecase, seatLockRepo, txManager)
	busHandler := httpBusHandler.NewBusHandler(busUsecase)

	r := gin.Default()

	api := r.Group("/api")

	r.Static("/ui", "./web")

	api.Use(middleware.RequestIdMiddleware())

	api.Use(middleware.ErrorHandler())

	httpBusDelivery.RegiserBusRouter(api, busHandler)
	httpBooking.RegisterRoutes(api, httpBooking.NewBookingHandler(lockSeatUsecase, confirmBookingUsecase))

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

	// 3. Close database and Redis connections
	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing DB connection: %v", err)
		} else {
			log.Println("DB connection closed successfully")
		}
	}

	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing Redis client: %v", err)
	} else {
		log.Println("Redis client closed successfully")
	}

	log.Println("Server exiting")
}
