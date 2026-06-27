package repository

import (
	"booking-system/internal/bus/domain"
	"context"

	"github.com/google/uuid"
)

type SeatRepository interface {
	Create(ctx context.Context, seat *domain.Seat) (*domain.Seat, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Seat, error)
	Update(ctx context.Context, seat *domain.Seat) (*domain.Seat, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByBusID(ctx context.Context, busID uuid.UUID) ([]*domain.Seat, error)
	GetByBusAndCodes(
		ctx context.Context,
		busID uuid.UUID,
		codes []string,
	) ([]*domain.Seat, error)
	BookSeats(
		ctx context.Context,
		busID uuid.UUID,
		codes []string,
	) error
}
