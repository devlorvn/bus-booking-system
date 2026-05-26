package repository

import (
	"booking-system/internal/bus/domain"
	"context"
)

type BusRepository interface {
	Create(ctx context.Context, bus *domain.Bus) (*domain.Bus, error)
	GetByID(ctx context.Context, id string) (*domain.Bus, error)
	Update(ctx context.Context, bus *domain.Bus) (*domain.Bus, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*domain.Bus, error)
}
