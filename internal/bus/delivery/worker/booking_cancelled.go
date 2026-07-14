package worker

import (
	"booking-system/internal/bus/usecase"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"log"

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
			log.Printf("[BookingCancelledWorker] Error fetching message: %v", err)
			continue
		}

		var event events.BookingCancelledEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("[BookingCancelledWorker] Error unmarshalling message: %v", err)
			w.reader.CommitMessages(ctx, msg)
			continue
		}

		if err := w.usecase.ReleaseSeatsByBookingID(ctx, event.BookingID); err != nil {
			log.Printf("[BookingCancelledWorker] Error releasing seats: %v", err)
			continue
		}

		if err := w.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("[BookingCancelledWorker] Error committing messages: %v", err)
			continue
		}
	}
}
