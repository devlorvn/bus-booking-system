package worker

import (
	handlepayment "booking-system/internal/booking/usecase/handle_payment"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"log/slog"
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
	slog.Info("Payment worker started")

	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		msg, err := w.reader.FetchMessage(ctx) // manual commit offset
		if err != nil {
			if ctx.Err() != nil {
				slog.Error("context cancelled")
				return ctx.Err()
			}
			slog.Error("Payment worker: Error fetching message:", slog.String("error", err.Error()))
			continue
		}
		if msg.Key == nil {
			slog.Error("Payment worker: message has no key, skipping")
			_ = w.reader.CommitMessages(ctx, msg) // Next to new message
			continue
		}

		var event events.PaymentProcessedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			slog.Error("Payment worker: error unmarshalling message:", slog.String("error", err.Error()))
			_ = w.reader.CommitMessages(ctx, msg) // Next to new message
			continue
		}

		slog.Info("Payment worker: processing payment for booking:", slog.String("booking_id", event.BookingID.String()), slog.String("status", event.Status))
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
			slog.Error("Payment worker: error handling payment after %d attempts: %v", slog.Int("attempts", constants.MaxKafkaEventRetry), slog.String("error", processErr.Error()))
			dlqMsg := gkafka.Message{
				Key:   msg.Key,
				Value: msg.Value,
				Headers: []gkafka.Header{{Key: "error_reason", Value: []byte(processErr.Error())},
					{Key: "retry_attempts", Value: []byte(strconv.Itoa(constants.MaxKafkaEventRetry))},
					{Key: "failed_at", Value: []byte(time.Now().Format(time.RFC3339))},
				},
			}

			if err := w.dlqWriter.WriteMessages(ctx, dlqMsg); err != nil {
				slog.Error("Payment worker: error writing to DLQ:", slog.String("error", err.Error()))
				continue
			}
		}

		if err := w.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("Payment worker: error committing message:", slog.String("error", err.Error()))
			continue
		}
	}
}
