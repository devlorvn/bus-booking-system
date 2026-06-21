package dto

import "github.com/google/uuid"

type ConfirmBookingRequest struct {
	TempUserID string    `json:"temp_user_id"`
	BusID      uuid.UUID `json:"bus_id"`
	SeatCodes  []string  `json:"seat_codes"`

	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}
