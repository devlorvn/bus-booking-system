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
	"context"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
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

	// Web socket
	hub := ws.NewHub()
	go hub.Run(ctx)

	h := ws.NewHandler(hub)

	eventPublisher := provider.NewWSEventPublisher(hub)

	busRepo := postgresRepository.NewBusRepository(db)
	seatRepo := postgresRepository.NewSeatRepository(db)

	busProvider := provider.NewBusProvider(busRepo, seatRepo)
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

	go func() {
		if err := paymentWorker.Start(ctx); err != nil {
			log.Fatal(err)
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
	go func() {
		if err := lockExpirationWorker.Start(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	seatUsecase := usecase.NewSeatUsecase(seatRepo, txManager)
	busUsecase := usecase.NewBusUsecase(busRepo, seatUsecase, seatLockRepo, txManager)
	busHandler := httpBusHandler.NewBusHandler(busUsecase)

	r := gin.Default()

	api := r.Group("/api")

	r.Static("/ui", "./web")

	api.Use(middleware.ErrorHandler())

	httpBusDelivery.RegiserBusRouter(api, busHandler)
	httpBooking.RegisterRoutes(api, httpBooking.NewBookingHandler(lockSeatUsecase, confirmBookingUsecase))

	r.GET("/ws/buses/:id", h.Handle)

	r.Run(":" + cfg.Port)
}
