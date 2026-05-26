package bus

import (
	"booking-system/pkg/shared"
	"context"
)

type Usecase struct {
	repo Repository
	tx   shared.Transaction
}

func NewUsecase(repo Repository, tx shared.Transaction) *Usecase {
	return &Usecase{repo: repo, tx: tx}
}

func (u *Usecase) Create(ctx context.Context, bus *Entity) (*Entity, error) {
	return u.repo.Create(ctx, bus)
}

func (u *Usecase) GetByID(ctx context.Context, id string) (*Entity, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *Usecase) Update(ctx context.Context, bus *Entity) (*Entity, error) {
	bus, err := u.repo.GetByID(ctx, bus.ID)
	if err != nil {
		return nil, err
	}

	return u.repo.Update(ctx, bus)
}

func (u *Usecase) Delete(ctx context.Context, id string) error {
	bus, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if bus == nil {
		return nil
	}
	return u.repo.Delete(ctx, id)
}

func (u *Usecase) List(ctx context.Context) ([]*Entity, error) {
	return u.repo.List(ctx)
}
