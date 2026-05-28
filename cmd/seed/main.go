package main

import (
	"log"

	"booking-system/configs"
	"booking-system/pkg/database/seed"
	"booking-system/pkg/postgres"
)

func main() {
	cfg := configs.LoadConfig()
	db, err := postgres.NewPostgres(&cfg.Database)
	if err != nil {
		log.Fatal(err)
	}

	err = seed.SeedBusWithSeats(db)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("seed completed")
}
