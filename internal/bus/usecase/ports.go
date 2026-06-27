package usecase

import (
	"context"

	"github.com/google/uuid"
)

type SeatPort interface {
	IsSeatLocked(ctx context.Context, busID uuid.UUID, seatCode string) (bool, error)
	GetLockedSeats(ctx context.Context, busID uuid.UUID, seatCodes []string) (map[string]bool, error)
}
