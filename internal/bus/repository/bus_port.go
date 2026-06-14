package repository

import (
	"booking-system/internal/bus/domain"
	"context"

	"github.com/google/uuid"
)

type BusPort interface {
	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (*domain.Bus, error)

	DecreaseAvailableSeats(
		ctx context.Context,
		busID uuid.UUID,
		count int,
	) error
}
