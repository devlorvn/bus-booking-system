package domain

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID            uuid.UUID `json:"id"`
	BusID         uuid.UUID `json:"bus_id"`
	UserID        uuid.UUID `json:"user_id"`
	Status        string    `json:"status"` // PAYMENT_PEDING, PAID, FAILED, CANCELLED, EXPIRED
	PaymentStatus string    `json:"payment_status"`
	TotalAmount   float64   `json:"total_amount"`
	TotalSeats    int       `json:"total_seats"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
