package usecase

import (
	"booking-system/internal/user/domain"
	"booking-system/internal/user/repository"
	"booking-system/pkg/shared"
	"context"

	"github.com/google/uuid"
)

type UserUsecase struct {
	repo repository.UserRepository
	tx   shared.Transaction
}

func NewUserUsecase(repo repository.UserRepository, tx shared.Transaction) *UserUsecase {
	return &UserUsecase{repo: repo, tx: tx}
}

func (u *UserUsecase) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	return u.repo.Create(ctx, user)
}

func (u *UserUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *UserUsecase) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	user, err := u.repo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return u.repo.Update(ctx, user)
}

func (u *UserUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if user == nil {
		return nil
	}

	return u.repo.Delete(ctx, id)
}

func (u *UserUsecase) List(ctx context.Context) ([]*domain.User, error) {
	return u.repo.List(ctx)
}
