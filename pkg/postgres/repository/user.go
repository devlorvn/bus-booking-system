package repository

import (
	"booking-system/internal/user/domain"
	"booking-system/pkg/database"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) dbFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	if err := r.dbFromContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := r.dbFromContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	if err := r.dbFromContext(ctx).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"name":         user.Name,
		"email":        user.Email,
		"phone_number": user.PhoneNumber,
	}).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.dbFromContext(ctx).Delete(&domain.User{}, "id = ?", id).Error
}

func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	var users []*domain.User
	if err := r.dbFromContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
