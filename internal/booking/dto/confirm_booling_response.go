package dto

import "github.com/google/uuid"

type ConfirmBookingResponse struct {
	BookingID   uuid.UUID
	TotalAmount float64
	Status      string
}
