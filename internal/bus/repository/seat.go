package repository

import (
	"booking-system/internal/bus/domain"
	"context"
)

type SeatRepository interface {
	Create(ctx context.Context, seat *domain.Seat) (*domain.Seat, error)
	GetByID(ctx context.Context, id string) (*domain.Seat, error)
	Update(ctx context.Context, seat *domain.Seat) (*domain.Seat, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*domain.Seat, error)
}
