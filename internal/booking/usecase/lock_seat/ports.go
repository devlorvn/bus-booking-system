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
	PublishSeatLocked(busID string, seatID string, seatCode string, tempUserID string) error
	PublishSeatReleased(busID string, seatID string, seatCode string) error
}
