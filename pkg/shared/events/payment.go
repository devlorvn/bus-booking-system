package events

import "github.com/google/uuid"

type BookingCreatedEvent struct {
	BookingID uuid.UUID `json:"booking_id"`
}
type PaymentProcessedEvent struct {
	BookingID uuid.UUID `json:"booking_id"`
	Status    string    `json:"status"` // "SUCCESS" hoặc "FAILED"
	Reason    string    `json:"reason,omitempty"`
}
