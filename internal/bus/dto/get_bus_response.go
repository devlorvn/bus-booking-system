package dto

import "booking-system/internal/bus/domain"

type GetBusResponse struct {
	domain.Bus
	Seats []domain.Seat `json:"seats"`
}
