package usecase

import (
	"booking-system/internal/bus/domain"
	"booking-system/internal/bus/repository"
	"booking-system/pkg/shared"
	"context"

	"github.com/google/uuid"
)

type BusUsecase struct {
	repo repository.BusRepository
	tx   shared.Transaction
}

func NewBusUsecase(repo repository.BusRepository, tx shared.Transaction) *BusUsecase {
	return &BusUsecase{repo: repo, tx: tx}
}

func (u *BusUsecase) Create(ctx context.Context, bus *domain.Bus) (*domain.Bus, error) {
	return u.repo.Create(ctx, bus)
}

func (u *BusUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bus, error) {
	return u.repo.GetByID(ctx, id)
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
