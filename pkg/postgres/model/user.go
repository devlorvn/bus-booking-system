package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name          string     `gorm:"type:varchar(100)"`
	Email         string     `gorm:"type:varchar(100)"`
	PhoneNumber   string     `gorm:"type:varchar(20);uniqueIndex"`
	LastBookingID *uuid.UUID `gorm:"type:uuid"`

	LastBooking *Booking `gorm:"foreignKey:LastBookingID;constraint:OnDelete:SET NULL;" json:"last_booking"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
