package model

import (
	"time"

	"github.com/google/uuid"
)

type Outbox struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	AggregateType string    `gorm:"type:varchar(255);not null"`
	AggregateID   string    `gorm:"type:varchar(255);not null"`
	EventType     string    `gorm:"type:varchar(255);not null"`
	Payload       []byte    `gorm:"type:jsonb;not null"`
	Status        string    `gorm:"type:varchar(50);not null;default:'PENDING'"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}
