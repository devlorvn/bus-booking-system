package dto

import "github.com/google/uuid"

type MarkBookedRequest struct {
	BookingID uuid.UUID `json:"booking_id" binding:"required"`
	BusID     uuid.UUID `json:"bus_id" binding:"required"`
	SeatCount int       `json:"seat_count" binding:"required"`
}
