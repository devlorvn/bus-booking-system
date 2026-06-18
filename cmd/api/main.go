package main

import (
	"booking-system/configs"
	lockseat "booking-system/internal/booking/usecase/lock_seat"
	httpBusDelivery "booking-system/internal/bus/delivery/http"
	httpBusHandler "booking-system/internal/bus/delivery/http/handler"
	"booking-system/internal/bus/usecase"
	"booking-system/internal/provider"
	"booking-system/pkg/database"
	"booking-system/pkg/postgres"
	postgresRepository "booking-system/pkg/postgres/repository"
	"booking-system/pkg/redis"
)

func main() {
	cfg := configs.LoadConfig()
	db, err := postgres.NewPostgres(&cfg.Database)
	if err != nil {
		panic(err)
	}
	txManager := database.NewTransaction(db)

	_ = redis.NewClient(&cfg.Redis)
	busRepo := postgresRepository.NewBusRepository(db)
	seatRepo := postgresRepository.NewSeatRepository(db)
	seatUsecase := usecase.NewSeatUsecase(seatRepo, txManager)
	busUsecase := usecase.NewBusUsecase(busRepo, seatUsecase, txManager)
	busHandler := httpBusHandler.NewBusHandler(busUsecase)

	busProvider := provider.NewBusProvider(busRepo, seatRepo)
	seatLockRepo := redis.NewLockSeatRepository(redis.NewClient(&cfg.Redis))
	// publisher := ws.NewNoopPublisher()
	_ = lockseat.NewLockSeatUsecase(
		busProvider,
		seatLockRepo,
		// publisher,
	)

	handleGroup := &httpBusDelivery.HandlerHttpGroup{
		BusHandler: busHandler,
	}

	r := httpBusDelivery.NewRouter(handleGroup)

	r.Run(":" + cfg.Port)
}
