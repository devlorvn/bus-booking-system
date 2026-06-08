package domain

import (
	"time"

	"github.com/google/uuid"
)

type Seat struct {
	ID       uuid.UUID `json:"id"`
	BusID    uuid.UUID `json:"bus_id"`
	SeatCode string    `json:"seat_code"`
	Status   string    `json:"status"`

	CreatedAt time.Time `json:"created_at"`
}
