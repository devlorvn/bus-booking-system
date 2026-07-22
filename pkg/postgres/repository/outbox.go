package repository

import (
	"booking-system/pkg/database"
	"booking-system/pkg/postgres/model"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *OutboxRepository) GetPending(ctx context.Context, limit int, maxRetries int) ([]*model.Outbox, error) {
	var events []*model.Outbox
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("status = ? AND retry_count < ?", "PENDING", maxRetries).
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *OutboxRepository) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Outbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": "PROCESSED",
		}).Error
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, reason string) error {
	return r.db.WithContext(ctx).
		Model(&model.Outbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     "FAILED",
			"last_error": reason,
		}).Error
}

func (r *OutboxRepository) RecordFailure(ctx context.Context, id uuid.UUID, errMsg string, maxRetries int) error {
	return r.db.WithContext(ctx).
		Model(&model.Outbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"retry_count": gorm.Expr("retry_count + 1"),
			"last_error":  errMsg,
			"status":      gorm.Expr("CASE WHEN retry_count + 1 >= ? THEN 'FAILED' ELSE 'PENDING' END", maxRetries),
		}).Error
}

func (r *OutboxRepository) DeleteProcessed(ctx context.Context, processedOlderThan time.Time, failedOlderThan time.Time) error {
	return r.db.WithContext(ctx).
		Where("(status = ? AND created_at < ?) OR (status = ? AND created_at < ?)",
			"PROCESSED", processedOlderThan,
			"FAILED", failedOlderThan,
		).
		Delete(&model.Outbox{}).Error
}
