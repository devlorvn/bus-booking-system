package seed

import (
	postgredModel "booking-system/pkg/postgres/model"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedBusWithSeats(db *gorm.DB) error {
	ctx := context.Background()

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// =========================
		// CREATE BUS
		// =========================

		busID := uuid.New()

		bus := postgredModel.Bus{
			ID:             busID,
			LicensePlate:   "51A-99999",
			FromLocation:   "Ho Chi Minh",
			ToLocation:     "Da Lat",
			DepartureTime:  time.Now().Add(24 * time.Hour),
			Price:          250000,
			TotalSeats:     24,
			AvailableSeats: 24,
			Status:         "OPEN",
		}

		if err := tx.Create(&bus).Error; err != nil {
			return err
		}

		// =========================
		// GENERATE SEATS
		// =========================

		rows := []string{"A", "B", "C", "D"}

		seatsPerRow := bus.TotalSeats / len(rows)

		seats := make([]postgredModel.Seat, 0)

		for _, row := range rows {
			for i := 1; i <= seatsPerRow; i++ {

				seat := postgredModel.Seat{
					ID:        uuid.New(),
					BusID:     busID,
					SeatCode:  fmt.Sprintf("%s%d", row, i),
					Status:    "AVAILABLE",
					CreatedAt: time.Now(),
				}

				seats = append(seats, seat)
			}
		}

		// =========================
		// BULK INSERT SEATS
		// =========================

		if err := tx.Create(&seats).Error; err != nil {
			return err
		}

		return nil
	})
}
