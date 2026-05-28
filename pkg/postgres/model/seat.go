package model

import (
	"time"

	"github.com/google/uuid"
)

type Seat struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	BusID    uuid.UUID `gorm:"type:uuid"`
	SeatCode string    `json:"seat_code"`
	Status   string    `json:"status"`

	Bus       Bus       `gorm:"foreignKey:BusID;constraint:OnDelete:CASCADE" json:"bus"`
	CreatedAt time.Time `json:"created_at"`
}
