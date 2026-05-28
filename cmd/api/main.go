package main

import (
	"booking-system/configs"
	"booking-system/internal/bus/usecase"
	httpDelivery "booking-system/internal/delivery/http"
	httpHandler "booking-system/internal/delivery/http/handler"
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
	busHandler := httpHandler.NewBusHandler(busUsecase)

	handleGroup := &httpDelivery.HandlerHttpGroup{
		BusHandler: busHandler,
	}

	r := httpDelivery.NewRouter(handleGroup)

	r.Run(":" + cfg.Port)
}
