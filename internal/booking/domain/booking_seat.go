package domain

import "github.com/google/uuid"

type BookingSeat struct {
	BookingID uuid.UUID
	SeatID    uuid.UUID
}
