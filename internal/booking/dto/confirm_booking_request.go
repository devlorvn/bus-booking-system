package dto

import "github.com/google/uuid"

type ConfirmBookingRequest struct {
	TempUserID string
	BusID      uuid.UUID
	SeatCodes  []string

	Name        string
	PhoneNumber string
	Email       string
}
