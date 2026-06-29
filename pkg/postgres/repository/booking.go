package repository

import (
	"booking-system/internal/booking/domain"
	"booking-system/pkg/database"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) dbFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *BookingRepository) Create(ctx context.Context, booking *domain.Booking) (*domain.Booking, error) {
	if err := r.dbFromContext(ctx).Create(booking).Error; err != nil {
		return nil, err
	}
	return booking, nil
}

func (r *BookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	var booking domain.Booking
	if err := r.dbFromContext(ctx).First(&booking, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) Update(ctx context.Context, booking *domain.Booking) (*domain.Booking, error) {
	if err := r.dbFromContext(ctx).Table("bookings").Where("id = ?", booking.ID).Updates(map[string]interface{}{
		"bus_id":         booking.BusID,
		"user_id":        booking.UserID,
		"status":         booking.Status,
		"payment_status": booking.PaymentStatus,
		"total_amount":   booking.TotalAmount,
		"total_seats":    booking.TotalSeats,
	}).Error; err != nil {
		return nil, err
	}
	return booking, nil
}

func (r *BookingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.dbFromContext(ctx).Delete(&domain.Booking{}, "id = ?", id).Error
}

func (r *BookingRepository) List(ctx context.Context) ([]*domain.Booking, error) {
	var bookings []*domain.Booking
	if err := r.dbFromContext(ctx).Find(&bookings).Error; err != nil {
		return nil, err
	}
	return bookings, nil
}
