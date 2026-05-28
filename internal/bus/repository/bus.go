package repository

import (
	"booking-system/internal/bus/domain"
	"context"

	"github.com/google/uuid"
)

type BusRepository interface {
	Create(ctx context.Context, bus *domain.Bus) (*domain.Bus, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Bus, error)
	Update(ctx context.Context, bus *domain.Bus) (*domain.Bus, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*domain.Bus, error)
}
