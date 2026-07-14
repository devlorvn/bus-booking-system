package repository

import (
	"booking-system/pkg/database"
	"booking-system/pkg/postgres/model"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) dbFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *OutboxRepository) Save(ctx context.Context, outbox *model.Outbox) error {
	return r.dbFromContext(ctx).Create(outbox).Error
}

func (r *OutboxRepository) GetPending(ctx context.Context, limit int) ([]*model.Outbox, error) {
	var events []*model.Outbox
	err := r.db.WithContext(ctx).
		Where("status = ?", "PENDING").
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *OutboxRepository) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Outbox{}).
		Where("id = ?", id).
		Update("status", "PROCESSED").Error
}
