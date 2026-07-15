package main

import (
	"log/slog"
	"os"

	"booking-system/configs"
	"booking-system/pkg/database/seed"
	"booking-system/pkg/postgres"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg := configs.LoadConfig()
	db, err := postgres.NewPostgres(&cfg.Database)
	if err != nil {
		slog.Error(err.Error())
	}

	err = seed.SeedBusWithSeats(db)
	if err != nil {
		slog.Error(err.Error())
	}

	slog.Info("seed completed")
}
