package model

import (
	"time"

	"github.com/google/uuid"
)

type Bus struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	LicensePlate   string    `json:"license_plate"`
	FromLocation   string    `json:"from_location"`
	ToLocation     string    `json:"to_location"`
	DepartureTime  time.Time `json:"departure_time"`
	Price          float64   `json:"price"`
	TotalSeats     int       `json:"total_seats"`
	AvailableSeats int       `json:"available_seats"`
	Status         string    `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
