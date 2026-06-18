package main

import (
	"booking-system/configs"
	httpBooking "booking-system/internal/booking/delivery/http"
	lockseat "booking-system/internal/booking/usecase/lock_seat"
	httpBusDelivery "booking-system/internal/bus/delivery/http"
	httpBusHandler "booking-system/internal/bus/delivery/http/handler"
	"booking-system/internal/bus/usecase"
	"booking-system/internal/provider"
	"booking-system/pkg/database"
	"booking-system/pkg/postgres"
	postgresRepository "booking-system/pkg/postgres/repository"
	"booking-system/pkg/redis"
	"booking-system/pkg/shared/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := configs.LoadConfig()
	db, err := postgres.NewPostgres(&cfg.Database)
	if err != nil {
		panic(err)
	}
	txManager := database.NewTransaction(db)

	redisClient := redis.NewClient(&cfg.Redis)
	busRepo := postgresRepository.NewBusRepository(db)
	seatRepo := postgresRepository.NewSeatRepository(db)
	seatUsecase := usecase.NewSeatUsecase(seatRepo, txManager)
	busUsecase := usecase.NewBusUsecase(busRepo, seatUsecase, txManager)
	busHandler := httpBusHandler.NewBusHandler(busUsecase)

	busProvider := provider.NewBusProvider(busRepo, seatRepo)
	seatLockRepo := redis.NewLockSeatRepository(redisClient)
	// publisher := ws.NewNoopPublisher()
	lockSeatUsecase := lockseat.NewLockSeatUsecase(
		busProvider,
		seatLockRepo,
		// publisher,
	)

	r := gin.Default()

	api := r.Group("/api")

	api.Use(middleware.ErrorHandler())

	httpBusDelivery.RegiserBusRouter(api, busHandler)
	httpBooking.RegisterRoutes(api, httpBooking.NewBookingHandler(lockSeatUsecase))

	r.Run(":" + cfg.Port)
}
