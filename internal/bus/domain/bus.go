package domain

import (
	"time"

	"github.com/google/uuid"
)

type Bus struct {
	ID             uuid.UUID `json:"id"`
	LicensePlate   string    `json:"license_plate"`
	FromLocation   string    `json:"from_location"`
	ToLocation     string    `json:"to_location"`
	DepartureTime  time.Time `json:"departure_time"`
	TotalSeats     int       `json:"total_seats"`
	AvailableSeats int       `json:"available_seats"`
	Price          float64   `json:"price"`
	Status         string    `json:"status"`

	Seats []Seat `json:"seats"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
