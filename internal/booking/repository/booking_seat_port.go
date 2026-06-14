package repository

import (
	"booking-system/internal/booking/domain"
	"context"
)

type BookingSeatPort interface {
	BulkCreate(
		ctx context.Context,
		items []domain.BookingSeat,
	) error
}
