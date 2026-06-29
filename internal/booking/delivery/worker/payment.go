package worker

import (
	"booking-system/internal/booking/event"
	"context"
	"log"
)

type PaymentWorker struct {
	bus *event.PaymentEventBus
}

func NewPaymentWorker(bus *event.PaymentEventBus) *PaymentWorker {
	return &PaymentWorker{
		bus: bus,
	}
}

func (w *PaymentWorker) Start(ctx context.Context) error {
	log.Println("Payment worker started")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case bookingEvent := <-w.bus.BookingCreated:
			log.Println("Payment requested for booking:", bookingEvent.BookingID)
		}
	}
}
