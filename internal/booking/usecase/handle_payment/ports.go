package handlepayment

import (
	"context"

	busDomain "booking-system/internal/bus/domain"

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
	GetSeatByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*busDomain.Seat, error)
	ReleaseSeatsByBookingID(ctx context.Context, bookingID uuid.UUID) error
}
