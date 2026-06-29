package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	PhoneNumber   string     `json:"phone_number"`
	LastBookingID *uuid.UUID `json:"last_booking_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
