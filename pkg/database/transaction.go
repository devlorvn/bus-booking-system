package database

import (
	"context"

	"gorm.io/gorm"
)

type txContextKey struct{}

type Transaction struct {
	db *gorm.DB
}

func NewTransaction(db *gorm.DB) *Transaction {
	return &Transaction{db: db}
}

func (t *Transaction) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txContext := context.WithValue(ctx, txContextKey{}, tx)
		return fn(txContext)
	})
}

func GetTx(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txContextKey{}).(*gorm.DB)
	return tx, ok
}
