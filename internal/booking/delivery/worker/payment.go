package worker

import (
	handlepayment "booking-system/internal/booking/usecase/handle_payment"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"log"
	"sync"

	gkafka "github.com/segmentio/kafka-go"
)

type PaymentWorker struct {
	reader  *gkafka.Reader
	handler *handlepayment.HandlePaymentSuccessUsecase
}

func NewPaymentWorker(reader *gkafka.Reader, handler *handlepayment.HandlePaymentSuccessUsecase) *PaymentWorker {
	return &PaymentWorker{
		reader:  reader,
		handler: handler,
	}
}

func (w *PaymentWorker) Start(ctx context.Context) error {
	log.Println("Payment worker started")

	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		msg, err := w.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("context cancelled")
				return ctx.Err()
			}
			log.Printf("Payment worker: Error reading message: %v", err)
			continue
		}
		if msg.Key == nil {
			log.Println("Payment worker: message has no key, skipping")
			continue
		}

		var event events.PaymentProcessedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Payment worker: error unmarshalling message: %v", err)
			continue
		}

		log.Printf("Payment worker: processing payment for booking %s: Status %s", event.BookingID, event.Status)

		if event.Status == "SUCCESS" {
			err = w.handler.Success(ctx, event.BookingID)
		} else {
			err = w.handler.Failed(ctx, event.BookingID)
		}
		if err != nil {
			log.Printf("Payment worker: error handling payment: %v", err)
			continue
		}
	}
}
