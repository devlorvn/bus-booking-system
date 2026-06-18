package repository

import (
	"booking-system/internal/bus/domain"
	"booking-system/pkg/database"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SeatRepository struct {
	db *gorm.DB
}

func NewSeatRepository(db *gorm.DB) *SeatRepository {
	return &SeatRepository{db: db}
}

func (r *SeatRepository) dbFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *SeatRepository) Create(ctx context.Context, seat *domain.Seat) (*domain.Seat, error) {
	if err := r.dbFromContext(ctx).Create(seat).Error; err != nil {
		return nil, err
	}
	return seat, nil
}

func (r *SeatRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Seat, error) {
	var seat domain.Seat
	if err := r.dbFromContext(ctx).First(&seat, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &seat, nil
}

func (r *SeatRepository) Update(ctx context.Context, seat *domain.Seat) (*domain.Seat, error) {
	if err := r.dbFromContext(ctx).Where("id = ?", seat.ID).Updates(map[string]interface{}{
		"bus_id":    seat.BusID,
		"seat_code": seat.SeatCode,
		"status":    seat.Status,
	}).Error; err != nil {
		return nil, err
	}
	return seat, nil
}

func (r *SeatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.dbFromContext(ctx).Delete(&domain.Seat{}, "id = ?", id).Error
}

func (r *SeatRepository) ListByBusID(ctx context.Context, busID uuid.UUID) ([]*domain.Seat, error) {
	var seats []*domain.Seat
	if err := r.dbFromContext(ctx).Where("bus_id = ?", busID).Find(&seats).Error; err != nil {
		return nil, err
	}
	return seats, nil
}

func (r *SeatRepository) GetByBusAndCodes(
	ctx context.Context,
	busID uuid.UUID,
	codes []string,
) ([]*domain.Seat, error) {
	var seats []*domain.Seat
	if err := r.dbFromContext(ctx).Where("bus_id = ? AND seat_code IN ?", busID, codes).Find(&seats).Error; err != nil {
		return nil, err
	}
	return seats, nil
}
