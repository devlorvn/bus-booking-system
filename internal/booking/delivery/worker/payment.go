package worker

import (
	"booking-system/internal/booking/event"
	"booking-system/internal/booking/ports"
	handlepayment "booking-system/internal/booking/usecase/handle_payment"
	"booking-system/pkg/shared/helpers"
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PaymentWorker struct {
	bus       *event.PaymentEventBus
	processor ports.PaymentProcessor
	handler   *handlepayment.HandlePaymentSuccessUsecase
}

func NewPaymentWorker(bus *event.PaymentEventBus, processor ports.PaymentProcessor, handler *handlepayment.HandlePaymentSuccessUsecase) *PaymentWorker {
	return &PaymentWorker{
		bus:       bus,
		processor: processor,
		handler:   handler,
	}
}

func (w *PaymentWorker) Start(ctx context.Context) error {
	log.Println("Payment worker started")

	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case bookingEvent := <-w.bus.BookingCreated:
			wg.Add(1)
			go func(bookingID uuid.UUID) {
				defer wg.Done()
				err := helpers.Retry(
					ctx,
					3,
					func() error {
						return w.processor.Process(
							ctx,
							bookingID,
						)
					},
				)
				if err != nil {
					log.Println("payment permanently failed:", err)
					select {
					case w.bus.DLQ <- event.DeadLetterEvent{
						EventType: "booking_created",
						EntityId:  bookingID.String(),
						Error:     err.Error(),
						FailedAt:  time.Now(),
					}:
					case <-ctx.Done():
						log.Println("payment DLQ write cancelled due to context cancellation")
					}
				}
			}(bookingEvent.BookingID)
		case event := <-w.bus.PaymentSuccess:
			err := w.handler.Success(ctx, event.BookingID)
			if err != nil {
				log.Println("payment success error:", err)
			}

		case event := <-w.bus.PaymentFailed:
			err := w.handler.Failed(ctx, event.BookingID)
			if err != nil {
				log.Println("payment failed error:", err)
			}
			log.Println(
				"payment failed:",
				event.BookingID,
				event.Reason,
			)
		}
	}
}
