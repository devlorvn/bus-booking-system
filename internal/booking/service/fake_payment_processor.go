package service

import (
	"booking-system/internal/booking/event"
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type FakePaymentProcessor struct {
	bus *event.PaymentEventBus
}

func NewFakePaymentProcessor(bus *event.PaymentEventBus) *FakePaymentProcessor {
	return &FakePaymentProcessor{
		bus: bus,
	}
}

func (f *FakePaymentProcessor) Process(ctx context.Context, bookingID uuid.UUID) error {
	log.Println("processing payment:", bookingID)

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-time.After(5 * time.Second):
	}

	success := rand.Intn(100) >= 20 // 80% success

	if success {
		f.bus.PaymentSuccess <- event.PaymentSuccessEvent{
			BookingID: bookingID,
		}
	} else {
		f.bus.PaymentFailed <- event.PaymentFailedEvent{
			BookingID: bookingID,
		}
	}

	return nil
}
