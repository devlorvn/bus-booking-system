package worker

import (
	"booking-system/internal/booking/event"
	"booking-system/internal/booking/ports"
	"context"
	"log"

	"github.com/google/uuid"
)

type PaymentWorker struct {
	bus       *event.PaymentEventBus
	processor ports.PaymentProcessor
}

func NewPaymentWorker(bus *event.PaymentEventBus, processor ports.PaymentProcessor) *PaymentWorker {
	return &PaymentWorker{
		bus:       bus,
		processor: processor,
	}
}

func (w *PaymentWorker) Start(ctx context.Context) error {
	log.Println("Payment worker started")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case bookingEvent := <-w.bus.BookingCreated:
			go func(bookingID uuid.UUID) {
				err := w.processor.Process(
					ctx,
					bookingID,
				)
				if err != nil {
					log.Println("payment error:", err)
				}
			}(bookingEvent.BookingID)
		}
	}
}
