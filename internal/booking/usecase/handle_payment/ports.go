package handlepayment

import (
	"context"

	"github.com/google/uuid"
)

type BusRepository interface {
	DecrementAvailableSeats(ctx context.Context, busID uuid.UUID, count int) error
}

type SeatRepository interface {
	MarkBookedByBookingID(
		ctx context.Context,
		bookingID uuid.UUID,
	) error
}
