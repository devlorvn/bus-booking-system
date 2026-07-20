package postgres

import (
	"booking-system/configs"
	"booking-system/pkg/postgres/model"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(cfg *configs.Database) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SslMode)
	
	var db *gorm.DB
	var err error
	maxRetries := 15
	for i := 1; i <= maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		fmt.Printf("Failed to connect to Postgres (attempt %d/%d): %v. Retrying in 2s...\n", i, maxRetries, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres after %d attempts: %w", maxRetries, err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("Error connecting to Postgres: %v", err)
	}

	sqlDb.SetMaxIdleConns(10)           // Keep max 10 connections alive in pool
	sqlDb.SetMaxOpenConns(100)          // Max open connections is 100
	sqlDb.SetConnMaxLifetime(time.Hour) // Max lifetime of a connection is 1 hour

	fmt.Println("Connected to Postgres successfully")

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&model.User{},
		&model.Bus{},
		&model.Seat{},
		&model.Booking{},
		&model.BookingSeat{},
		&model.Outbox{},
	)
	if err != nil {
		return err
	}
	fmt.Println("Database migration completed successfully")
	return nil
}
