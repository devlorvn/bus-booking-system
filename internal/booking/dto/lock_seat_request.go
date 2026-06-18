package dto

import (
	"github.com/google/uuid"
)

type LockSeatRequest struct {
	BusID      uuid.UUID `json:"bus_id" validate:"required"`
	TempUserID uuid.UUID `json:"temp_user_id" validate:"required"`
	SeatCodes  []string  `json:"seat_codes" validate:"required"`
}
