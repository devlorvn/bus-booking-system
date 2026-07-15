package worker

import (
	"booking-system/internal/bus/usecase"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"log/slog"

	gkafka "github.com/segmentio/kafka-go"
)

type BookingCancelledWorker struct {
	reader  *gkafka.Reader
	usecase *usecase.SeatUsecase
}

func NewBookingCancelledWorker(reader *gkafka.Reader, usecase *usecase.SeatUsecase) *BookingCancelledWorker {
	return &BookingCancelledWorker{reader: reader, usecase: usecase}
}

func (w *BookingCancelledWorker) Start(ctx context.Context) error {
	defer w.reader.Close()

	for {
		msg, err := w.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Error("BookingCancelledWorker: Error fetching message:", slog.String("error", err.Error()))
			continue
		}

		var event events.BookingCancelledEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			slog.Error("BookingCancelledWorker: Error unmarshalling message:", slog.String("error", err.Error()))
			w.reader.CommitMessages(ctx, msg)
			continue
		}

		if err := w.usecase.ReleaseSeatsByBookingID(ctx, event.BookingID); err != nil {
			slog.Error("BookingCancelledWorker: Error releasing seats:", slog.String("error", err.Error()))
			continue
		}

		if err := w.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("BookingCancelledWorker: Error committing messages:", slog.String("error", err.Error()))
			continue
		}
	}
}
