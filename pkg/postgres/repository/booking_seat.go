package repository

import (
	"booking-system/internal/booking/domain"
	"booking-system/pkg/database"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookingSeatRepository struct {
	db *gorm.DB
}

func NewBookingSeatRepository(db *gorm.DB) *BookingSeatRepository {
	return &BookingSeatRepository{db: db}
}

func (r *BookingSeatRepository) dbFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *BookingSeatRepository) BulkCreate(ctx context.Context, items []*domain.BookingSeat) error {
	return r.dbFromContext(ctx).Create(&items).Error
}

func (r *BookingSeatRepository) GetByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*domain.BookingSeat, error) {
	var items []*domain.BookingSeat
	err := r.dbFromContext(ctx).Where("booking_id = ?", bookingID).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
