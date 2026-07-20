package worker

import (
	"booking-system/internal/bus/usecase"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"log/slog"

	gkafka "github.com/segmentio/kafka-go"
)

type BookingWorker struct {
	reader  *gkafka.Reader
	writer  *gkafka.Writer
	usecase *usecase.SeatUsecase
}

func NewBookingWorker(
	reader *gkafka.Reader,
	writer *gkafka.Writer,
	usecase *usecase.SeatUsecase,
) *BookingWorker {
	return &BookingWorker{
		reader:  reader,
		writer:  writer,
		usecase: usecase,
	}
}

func (w *BookingWorker) Start(ctx context.Context) error {
	slog.Info("BookingWorker started")
	defer w.reader.Close()
	if w.writer != nil {
		defer w.writer.Close()
	}

	for {
		msg, err := w.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Error("BookingWorker: Error fetching message:", slog.String("error", err.Error()))
			continue
		}

		// Parse event type header
		eventType := ""
		for _, h := range msg.Headers {
			if h.Key == "event_type" {
				eventType = string(h.Value)
				break
			}
		}

		switch eventType {
		case constants.EventTypeBookingCreated:
			var event events.BookingCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				slog.Error("BookingWorker: Error unmarshalling created event:", slog.String("error", err.Error()))
				_ = w.reader.CommitMessages(ctx, msg)
				continue
			}
			w.handleBookingCreated(ctx, event)

		case constants.EventTypeBookingCancelled:
			var event events.BookingCancelledEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				slog.Error("BookingWorker: Error unmarshalling cancelled event:", slog.String("error", err.Error()))
				_ = w.reader.CommitMessages(ctx, msg)
				continue
			}
			w.handleBookingCancelled(ctx, event)

		default:
			// Commit other events without processing
			_ = w.reader.CommitMessages(ctx, msg)
			continue
		}

		if err := w.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("BookingWorker: Error committing message:", slog.String("error", err.Error()))
			continue
		}
	}
}

func (w *BookingWorker) handleBookingCreated(ctx context.Context, event events.BookingCreatedEvent) {
	slog.Info("BookingWorker: Booking created event received", slog.String("booking_id", event.BookingID.String()))
	err := w.usecase.BookSeats(ctx, event.BusID, event.SeatCodes)
	if err != nil {
		slog.Error("BookingWorker: Seat reservation failed", slog.String("booking_id", event.BookingID.String()), slog.String("error", err.Error()))
		w.publishFailed(ctx, event, err)
	} else {
		slog.Info("BookingWorker: Seat reservation succeeded", slog.String("booking_id", event.BookingID.String()))
		w.publishSuccess(ctx, event)
	}
}

func (w *BookingWorker) handleBookingCancelled(ctx context.Context, event events.BookingCancelledEvent) {
	slog.Info("BookingWorker: Booking cancelled event received", slog.String("booking_id", event.BookingID.String()))
	err := w.usecase.ReleaseSeats(ctx, event.BusID, event.SeatCodes)
	if err != nil {
		slog.Error("BookingWorker: Failed to release seats:", slog.String("error", err.Error()))
	}
}

func (w *BookingWorker) publishSuccess(ctx context.Context, event events.BookingCreatedEvent) {
	if w.writer == nil {
		return
	}
	reservedEvent := events.SeatsReservedEvent{
		BookingID: event.BookingID,
		BusID:     event.BusID,
		SeatCodes: event.SeatCodes,
	}
	bytes, _ := json.Marshal(reservedEvent)
	err := w.writer.WriteMessages(ctx, gkafka.Message{
		Key:   []byte(event.BookingID.String()),
		Value: bytes,
		Headers: []gkafka.Header{
			{Key: "event_type", Value: []byte(constants.EventTypeSeatsReserved)},
		},
	})
	if err != nil {
		slog.Error("BookingWorker: Failed to publish reservation success event", slog.String("error", err.Error()))
	}
}

func (w *BookingWorker) publishFailed(ctx context.Context, event events.BookingCreatedEvent, err error) {
	if w.writer == nil {
		return
	}
	failedEvent := events.SeatsReservationFailedEvent{
		BookingID: event.BookingID,
		BusID:     event.BusID,
		SeatCodes: event.SeatCodes,
		Reason:    err.Error(),
	}
	bytes, _ := json.Marshal(failedEvent)
	pubErr := w.writer.WriteMessages(ctx, gkafka.Message{
		Key:   []byte(event.BookingID.String()),
		Value: bytes,
		Headers: []gkafka.Header{
			{Key: "event_type", Value: []byte(constants.EventTypeSeatsReservationFailed)},
		},
	})
	if pubErr != nil {
		slog.Error("BookingWorker: Failed to publish reservation failure event", slog.String("error", pubErr.Error()))
	}
}
