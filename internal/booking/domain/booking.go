package domain

import "github.com/google/uuid"

type Booking struct {
	ID            uuid.UUID `json:"id"`
	BusID         uuid.UUID `json:"bus_id"`
	UserID        uuid.UUID `json:"user_id"`
	Status        string    `json:"status"`
	PaymentStatus string    `json:"payment_status"`
	TotalAmount   float64   `json:"total_amount"`
	TotalSeats    int       `json:"total_seats"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
