package model

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	BusID         uuid.UUID `gorm:"type:uuid"`
	UserID        uuid.UUID `gorm:"type:uuid"`
	Status        string    `gorm:"type:varchar(50)"`
	PaymentStatus string    `gorm:"type:varchar(50)"`
	TotalAmount   float64   `gorm:"type:decimal(10,2)"`
	TotalSeats    int       `gorm:"type:int"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
