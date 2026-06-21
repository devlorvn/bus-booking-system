package lockseat

import (
	"context"

	"github.com/google/uuid"
)

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
