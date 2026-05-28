package usecase

import (
	"booking-system/internal/bus/domain"
	"booking-system/internal/bus/dto"
	"booking-system/internal/bus/repository"
	"booking-system/pkg/shared"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type BusUsecase struct {
	repo        repository.BusRepository
	seatUsecase *SeatUsecase
	tx          shared.Transaction
}

func NewBusUsecase(repo repository.BusRepository, seatUsecase *SeatUsecase, tx shared.Transaction) *BusUsecase {
	return &BusUsecase{repo: repo, seatUsecase: seatUsecase, tx: tx}
}

func (u *BusUsecase) Create(ctx context.Context, req dto.CreateBusRequest) (*domain.Bus, error) {
	bus := &domain.Bus{
		ID:             uuid.New(),
		LicensePlate:   req.LicensePlate,
		FromLocation:   req.FromLocation,
		ToLocation:     req.ToLocation,
		DepartureTime:  req.DepartureTime,
		Price:          req.Price,
		TotalSeats:     req.TotalSeats,
		AvailableSeats: req.TotalSeats,
		Status:         "OPEN",
	}

	u.tx.Execute(ctx, func(txCtx context.Context) error {
		u.repo.Create(txCtx, bus)

		for row := range req.RowName {
			for i := 1; i <= req.SeatsPerRow; i++ {
				seat := &domain.Seat{
					ID:       uuid.New(),
					BusID:    bus.ID,
					SeatCode: fmt.Sprintf("%s%d", req.RowName[row], i),
					Status:   "AVAILABLE",
				}
				u.seatUsecase.Create(txCtx, seat)
			}
		}

		return nil
	})
	return bus, nil
}

func (u *BusUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bus, error) {
	bus, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	seats, err := u.seatUsecase.ListByBusID(ctx, bus.ID)
	if err != nil {
		return nil, err
	}
	busSeats := make([]domain.Seat, 0, len(seats))
	for _, seat := range seats {
		if seat == nil {
			continue
		}
		busSeats = append(busSeats, *seat)
	}

	bus.Seats = busSeats
	return bus, nil
}

func (u *BusUsecase) Update(ctx context.Context, bus *domain.Bus) (*domain.Bus, error) {
	bus, err := u.repo.GetByID(ctx, bus.ID)
	if err != nil {
		return nil, err
	}

	return u.repo.Update(ctx, bus)
}

func (u *BusUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	bus, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if bus == nil {
		return nil
	}
	return u.repo.Delete(ctx, id)
}

func (u *BusUsecase) List(ctx context.Context) ([]*domain.Bus, error) {
	return u.repo.List(ctx)
}
