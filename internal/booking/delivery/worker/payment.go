package worker

import (
	handlepayment "booking-system/internal/booking/usecase/handle_payment"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	gkafka "github.com/segmentio/kafka-go"
)

type PaymentWorker struct {
	reader    *gkafka.Reader
	dlqWriter *gkafka.Writer
	handler   *handlepayment.HandlePaymentUsecase
}

func NewPaymentWorker(reader *gkafka.Reader, dlqWriter *gkafka.Writer, handler *handlepayment.HandlePaymentUsecase) *PaymentWorker {
	return &PaymentWorker{
		reader:    reader,
		dlqWriter: dlqWriter,
		handler:   handler,
	}
}

func (w *PaymentWorker) Start(ctx context.Context) error {
	log.Println("Payment worker started")

	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		msg, err := w.reader.FetchMessage(ctx) // manual commit offset
		if err != nil {
			if ctx.Err() != nil {
				log.Println("context cancelled")
				return ctx.Err()
			}
			log.Printf("Payment worker: Error fetching message: %v", err)
			continue
		}
		if msg.Key == nil {
			log.Println("Payment worker: message has no key, skipping")
			_ = w.reader.CommitMessages(ctx, msg) // Next to new message
			continue
		}

		var event events.PaymentProcessedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Payment worker: error unmarshalling message: %v", err)
			_ = w.reader.CommitMessages(ctx, msg) // Next to new message
			continue
		}

		log.Printf("Payment worker: processing payment for booking %s: Status %s", event.BookingID, event.Status)
		backoff := constants.DelayRetry
		var processErr error

		// In memorry retry with delay and exponential backoff
		for attempt := 1; attempt <= constants.MaxKafkaEventRetry; attempt++ {
			if event.Status == "SUCCESS" {
				processErr = w.handler.Success(ctx, event.BookingID)
			} else {
				processErr = w.handler.Failed(ctx, event.BookingID)
			}
			if processErr == nil {
				break
			}
			if attempt < constants.MaxKafkaEventRetry {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
					backoff *= 2
					continue
				}
			}
		}

		if processErr != nil {
			log.Printf("Payment worker: error handling payment after %d attempts: %v", constants.MaxKafkaEventRetry, processErr)
			dlqMsg := gkafka.Message{
				Key:   msg.Key,
				Value: msg.Value,
				Headers: []gkafka.Header{{Key: "error_reason", Value: []byte(processErr.Error())},
					{Key: "retry_attempts", Value: []byte(strconv.Itoa(constants.MaxKafkaEventRetry))},
					{Key: "failed_at", Value: []byte(time.Now().Format(time.RFC3339))},
				},
			}

			if err := w.dlqWriter.WriteMessages(ctx, dlqMsg); err != nil {
				log.Printf("Payment worker: error writing to DLQ: %v", err)
				continue
			}
		}

		if err := w.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("Payment worker: error committing message: %v", err)
			continue
		}
	}
}
