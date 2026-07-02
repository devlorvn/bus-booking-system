package event

import (
	"time"

	"github.com/google/uuid"
)

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

type DeadLetterEvent struct {
	EventType string
	EntityId  string
	Error     string
	FailedAt  time.Time
}
