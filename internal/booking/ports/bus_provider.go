package ports

import (
	"context"

	busDomain "booking-system/internal/bus/domain"
	"booking-system/pkg/shared/events"

	"github.com/google/uuid"
)

type BusProvider interface {
	GetBus(
		ctx context.Context,
		busID uuid.UUID,
	) (*busDomain.Bus, error)

	GetSeatsByCodes(
		ctx context.Context,
		busID uuid.UUID,
		codes []string,
	) ([]*busDomain.Seat, error)

	BookSeats(
		ctx context.Context,
		busID uuid.UUID,
		codes []string,
	) error

	MarkBookedByBookingID(
		ctx context.Context,
		bookingID uuid.UUID,
		busID uuid.UUID,
		seatCount int,
	) error

	GetSeatByBookingID(
		ctx context.Context,
		bookingID uuid.UUID,
	) ([]*busDomain.Seat, error)
}

type EventPublisher interface {
	PublishSeatLocked(busID string, seatID string, seatCode string, tempUserID string) error
	PublishSeatReleased(busID string, seatCode string) error
	PublishBookingCancelled(ctx context.Context, event events.BookingCancelledEvent) error
}
