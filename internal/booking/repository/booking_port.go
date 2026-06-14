package repository

import (
	"booking-system/internal/booking/domain"
	"context"
)

type BookingPort interface {
	Create(ctx context.Context, booking *domain.Booking) (*domain.Booking, error)
}
