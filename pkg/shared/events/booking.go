package events

import (
	"encoding/json"

	"github.com/google/uuid"
)

type BookingCreatedEvent struct {
	BookingID uuid.UUID `json:"booking_id"`
	BusID     uuid.UUID `json:"bus_id"`
	SeatCodes []string  `json:"seat_codes"`
}

type BookingPendingPaymentEvent struct {
	BookingID uuid.UUID `json:"booking_id"`
}

type PaymentProcessedEvent struct {
	BookingID uuid.UUID `json:"booking_id"`
	Status    string    `json:"status"` // "SUCCESS" hoặc "FAILED"
	Reason    string    `json:"reason,omitempty"`
}

type KafkaWsEvent struct {
	Event string          `json:"event"`
	BusID string          `json:"bus_id"`
	Data  json.RawMessage `json:"data"`
}

type BookingCancelledEvent struct {
	BookingID uuid.UUID `json:"booking_id"`
	BusID     uuid.UUID `json:"bus_id"`
	SeatCodes []string  `json:"seat_codes"`
}

type SeatsReservedEvent struct {
	BookingID uuid.UUID `json:"booking_id"`
	BusID     uuid.UUID `json:"bus_id"`
	SeatCodes []string  `json:"seat_codes"`
}

type SeatsReservationFailedEvent struct {
	BookingID uuid.UUID `json:"booking_id"`
	BusID     uuid.UUID `json:"bus_id"`
	SeatCodes []string  `json:"seat_codes"`
	Reason    string    `json:"reason"`
}

type BookingConfirmedEvent struct {
	BookingID uuid.UUID `json:"booking_id"`
	BusID     uuid.UUID `json:"bus_id"`
}
