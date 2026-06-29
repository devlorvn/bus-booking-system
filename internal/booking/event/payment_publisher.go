package event

import (
	"context"

	"github.com/google/uuid"
)

type PaymentPublisher struct {
	bus *PaymentEventBus
}

func NewPaymentPublisher(bus *PaymentEventBus) *PaymentPublisher {
	return &PaymentPublisher{
		bus: bus,
	}
}

func (p *PaymentPublisher) PublishBookingCreated(ctx context.Context, bookingID uuid.UUID) error {
	event := BookingCreatedEvent{
		BookingID: bookingID.String(),
	}
	p.bus.BookingCreated <- event
	return nil
}
