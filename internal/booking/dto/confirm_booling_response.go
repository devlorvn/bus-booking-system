package dto

import "github.com/google/uuid"

type ConfirmBookingResponse struct {
	BookingID   uuid.UUID `json:"booking_id"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
}
