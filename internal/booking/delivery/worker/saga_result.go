package worker

import (
	processsagaresult "booking-system/internal/booking/usecase/process_saga_result"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"log/slog"

	gkafka "github.com/segmentio/kafka-go"
)

type SagaResultWorker struct {
	reader  *gkafka.Reader
	usecase *processsagaresult.ProcessSagaResultUsecase
}

func NewSagaResultWorker(
	reader *gkafka.Reader,
	usecase *processsagaresult.ProcessSagaResultUsecase,
) *SagaResultWorker {
	return &SagaResultWorker{
		reader:  reader,
		usecase: usecase,
	}
}

func (w *SagaResultWorker) Start(ctx context.Context) error {
	slog.Info("SagaResultWorker started")
	defer w.reader.Close()

	for {
		msg, err := w.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Error("SagaResultWorker: Error fetching message:", slog.String("error", err.Error()))
			continue
		}

		// Check event type header
		eventType := ""
		for _, h := range msg.Headers {
			if h.Key == "event_type" {
				eventType = string(h.Value)
				break
			}
		}

		switch eventType {
		case constants.EventTypeSeatsReserved:
			var event events.SeatsReservedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				slog.Error("SagaResultWorker: Error unmarshalling seats reserved event:", slog.String("error", err.Error()))
				_ = w.reader.CommitMessages(ctx, msg)
				continue
			}

			slog.Info("SagaResultWorker: Processing seats reserved event", slog.String("booking_id", event.BookingID.String()))
			if err := w.usecase.OnSeatsReserved(ctx, event.BookingID); err != nil {
				slog.Error("SagaResultWorker: OnSeatsReserved failed:", slog.String("error", err.Error()))
				continue
			}

		case constants.EventTypeSeatsReservationFailed:
			var event events.SeatsReservationFailedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				slog.Error("SagaResultWorker: Error unmarshalling seats failed event:", slog.String("error", err.Error()))
				_ = w.reader.CommitMessages(ctx, msg)
				continue
			}

			slog.Info("SagaResultWorker: Processing seats reservation failed event", slog.String("booking_id", event.BookingID.String()))
			if err := w.usecase.OnSeatsReservationFailed(ctx, event.BookingID, event.SeatCodes); err != nil {
				slog.Error("SagaResultWorker: OnSeatsReservationFailed failed:", slog.String("error", err.Error()))
				continue
			}

		default:
			// Skip other event types on this topic
			_ = w.reader.CommitMessages(ctx, msg)
			continue
		}

		if err := w.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("SagaResultWorker: Error committing message:", slog.String("error", err.Error()))
			continue
		}
	}
}
