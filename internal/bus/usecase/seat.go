package usecase

import (
	"booking-system/internal/bus/domain"
	"booking-system/internal/bus/repository"
	"booking-system/pkg/shared"
	"context"

	"github.com/google/uuid"
)

type SeatUsecase struct {
	repo repository.SeatRepository
	tx   shared.Transaction
}

func NewSeatUsecase(repo repository.SeatRepository, tx shared.Transaction) *SeatUsecase {
	return &SeatUsecase{repo: repo, tx: tx}
}

func (u *SeatUsecase) Create(ctx context.Context, Seat *domain.Seat) (*domain.Seat, error) {
	return u.repo.Create(ctx, Seat)
}

func (u *SeatUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Seat, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *SeatUsecase) Update(ctx context.Context, Seat *domain.Seat) (*domain.Seat, error) {
	Seat, err := u.repo.GetByID(ctx, Seat.ID)
	if err != nil {
		return nil, err
	}

	return u.repo.Update(ctx, Seat)
}

func (u *SeatUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	Seat, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if Seat == nil {
		return nil
	}
	return u.repo.Delete(ctx, id)
}

func (u *SeatUsecase) List(ctx context.Context) ([]*domain.Seat, error) {
	return u.repo.List(ctx)
}
