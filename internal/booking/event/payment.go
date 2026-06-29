package event

import "github.com/google/uuid"

type BookingCreatedEvent struct {
	BookingID uuid.UUID `json:"booking_id"`
}

type PaymentSuccessEvent struct {
	BookingID uuid.UUID `json:"booking_id"`
}

type PaymentFailedEvent struct {
	BookingID uuid.UUID `json:"booking_id"`
	Reason    string    `json:"reason"`
}
