package domain

import (
	"github.com/google/uuid"
)

type Bus struct {
	ID             uuid.UUID `json:"id"`
	LicensePlate   string    `json:"license_plate"`
	FromLocation   string    `json:"from_location"`
	ToLocation     string    `json:"to_location"`
	DepartureTime  string    `json:"departure_time"`
	TotalSeats     int       `json:"total_seats"`
	AvailableSeats int       `json:"available_seats"`
	Price          int       `json:"price"`
	Status         string    `json:"status"`

	Seats []Seat

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
