package repository

import (
	"booking-system/internal/booking/domain"
	"context"

	"github.com/google/uuid"
)

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) (*domain.Booking, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error)
	Update(ctx context.Context, booking *domain.Booking) (*domain.Booking, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*domain.Booking, error)
}
