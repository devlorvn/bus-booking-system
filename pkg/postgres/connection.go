package postgres

import (
	"booking-system/configs"
	"booking-system/pkg/postgres/model"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(cfg *configs.Database) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SslMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to Postgres successfully")
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&model.Bus{},
		&model.Seat{},
		&model.Booking{},
		&model.BookingSeat{},
	)
	if err != nil {
		return err
	}
	fmt.Println("Database migration completed successfully")
	return nil
}
