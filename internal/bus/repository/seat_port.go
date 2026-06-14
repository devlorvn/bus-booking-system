package repository

import (
	"booking-system/internal/bus/domain"
	"context"

	"github.com/google/uuid"
)

type SeatPort interface {
	GetByCodes(
		ctx context.Context,
		busID uuid.UUID,
		codes []string,
	) ([]*domain.Seat, error)

	MarkBooked(
		ctx context.Context,
		seatIDs []uuid.UUID,
	) error
}
