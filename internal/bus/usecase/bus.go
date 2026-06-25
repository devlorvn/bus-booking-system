package usecase

import (
	"booking-system/internal/bus/domain"
	"booking-system/internal/bus/dto"
	"booking-system/internal/bus/repository"
	"booking-system/pkg/shared"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type BusUsecase struct {
	repo        repository.BusRepository
	seatUsecase *SeatUsecase
	seatPort    SeatPort
	tx          shared.Transaction
}

func NewBusUsecase(repo repository.BusRepository, seatUsecase *SeatUsecase, seatPort SeatPort, tx shared.Transaction) *BusUsecase {
	return &BusUsecase{repo: repo, seatUsecase: seatUsecase, seatPort: seatPort, tx: tx}
}

func (u *BusUsecase) Create(ctx context.Context, req dto.CreateBusRequest) (*domain.Bus, error) {
	totalSeats := len(req.RowName) * req.SeatsPerRow
	bus := &domain.Bus{
		ID:             uuid.New(),
		LicensePlate:   req.LicensePlate,
		FromLocation:   req.FromLocation,
		ToLocation:     req.ToLocation,
		DepartureTime:  req.DepartureTime,
		Price:          req.Price,
		TotalSeats:     totalSeats,
		AvailableSeats: totalSeats,
		Status:         "OPEN",
	}

	err := u.tx.Execute(ctx, func(txCtx context.Context) error {
		_, err := u.repo.Create(txCtx, bus)
		if err != nil {
			return err
		}

		for row := range req.RowName {
			for i := 1; i <= req.SeatsPerRow; i++ {
				seat := &domain.Seat{
					ID:       uuid.New(),
					BusID:    bus.ID,
					SeatCode: fmt.Sprintf("%s%d", req.RowName[row], i),
					Status:   "AVAILABLE",
				}
				_, err := u.seatUsecase.Create(txCtx, seat)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return bus, nil
}

func (u *BusUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bus, error) {
	bus, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if bus == nil {
		return nil, errors.New("Bus not found")
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

func (u *BusUsecase) GetSeats(ctx context.Context, busID uuid.UUID) ([]*domain.Seat, error) {
	bus, err := u.repo.GetByID(ctx, busID)
	if err != nil {
		return nil, err
	}
	if bus == nil {
		return nil, errors.New("Bus not found")
	}
	seats, err := u.seatUsecase.ListByBusID(ctx, busID)
	if err != nil {
		return nil, err
	}

	for _, seat := range seats {
		isLocked, err := u.seatPort.IsSeatLocked(ctx, busID, seat.SeatCode)
		if err != nil {
			return nil, err
		}
		if isLocked {
			seat.Status = "LOCKED"
		}
	}

	return seats, nil
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
