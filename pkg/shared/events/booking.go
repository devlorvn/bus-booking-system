package events

import (
	"github.com/google/uuid"
)

type BookingCancelledEvent struct {
	BookingID uuid.UUID `json:"booking_id"`
	BusID     uuid.UUID `json:"bus_id"`
	SeatCodes []string  `json:"seat_codes"`
}
