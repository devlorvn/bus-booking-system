package lockseat

import (
	"context"

	busDomain "booking-system/internal/bus/domain"

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
}

type SeatLockPort interface {
	AcquireSeatLocks(
		ctx context.Context,
		busID uuid.UUID,
		seatCodes []string,
		tempUserID string,
	) error

	ReleaseSeatLocks(
		ctx context.Context,
		busID uuid.UUID,
		seatCodes []string,
		tempUserID string,
	) error
}

type EventPublisher interface {
	PublishSeatLocked(
		ctx context.Context,
		busID uuid.UUID,
		seatCodes []string,
	) error
}
