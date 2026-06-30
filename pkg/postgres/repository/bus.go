package repository

import (
	"booking-system/internal/bus/domain"
	"booking-system/pkg/database"
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BusRepository struct {
	db *gorm.DB
}

func NewBusRepository(db *gorm.DB) *BusRepository {
	return &BusRepository{db: db}
}

func (r *BusRepository) dbFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *BusRepository) Create(ctx context.Context, bus *domain.Bus) (*domain.Bus, error) {
	if err := r.dbFromContext(ctx).Create(bus).Error; err != nil {
		return nil, err
	}
	return bus, nil
}

func (r *BusRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bus, error) {
	var bus domain.Bus
	if err := r.dbFromContext(ctx).First(&bus, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &bus, nil
}

func (r *BusRepository) Update(ctx context.Context, bus *domain.Bus) (*domain.Bus, error) {
	if err := r.dbFromContext(ctx).Table("buses").Where("id = ?", bus.ID).Updates(map[string]interface{}{
		"license_plate":   bus.LicensePlate,
		"from_location":   bus.FromLocation,
		"to_location":     bus.ToLocation,
		"seats":           bus.Seats,
		"departure_time":  bus.DepartureTime,
		"total_seats":     bus.TotalSeats,
		"available_seats": bus.AvailableSeats,
		"price":           bus.Price,
		"status":          bus.Status,
	}).Error; err != nil {
		return nil, err
	}
	return bus, nil
}

func (r *BusRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.dbFromContext(ctx).Delete(&domain.Bus{}, "id = ?", id).Error
}

func (r *BusRepository) List(ctx context.Context) ([]*domain.Bus, error) {
	var buses []*domain.Bus
	if err := r.dbFromContext(ctx).Find(&buses).Error; err != nil {
		return nil, err
	}
	return buses, nil
}

func (r *BusRepository) DecrementAvailableSeats(ctx context.Context, busID uuid.UUID, count int) error {
	result := r.dbFromContext(ctx).Model(&domain.Bus{}).
		Where("id = ? AND available_seats >= ?", busID, count).
		Update("available_seats", gorm.Expr("available_seats - ?", count))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("NOT_ENOUGH_SEATS_AVAILABLE")
	}
	return nil
}
