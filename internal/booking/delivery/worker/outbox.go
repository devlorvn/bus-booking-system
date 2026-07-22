package worker

import (
	"booking-system/pkg/postgres/repository"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

const maxOutboxRetries = 5

type EventPublisher interface {
	PublishBookingCreated(ctx context.Context, event events.BookingCreatedEvent) error
	PublishBookingCancelled(ctx context.Context, event events.BookingCancelledEvent) error
	PublishBookingPendingPayment(ctx context.Context, event events.BookingPendingPaymentEvent) error
	PublishBookingConfirmed(ctx context.Context, event events.BookingConfirmedEvent) error
}

type OutboxWorker struct {
	repo         *repository.OutboxRepository
	publisher    EventPublisher
	pollInterval time.Duration
}

func NewOutboxWorker(
	repo *repository.OutboxRepository,
	publisher EventPublisher,
	pollInterval time.Duration,
) *OutboxWorker {
	return &OutboxWorker{
		repo:         repo,
		publisher:    publisher,
		pollInterval: pollInterval,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) error {
	slog.Info("Outbox worker started")
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			w.processPendingEvents(ctx)
		case <-cleanupTicker.C:
			w.cleanupProcessedEvents(ctx)
		}
	}
}

func (w *OutboxWorker) processPendingEvents(ctx context.Context) {
	pendingEvents, err := w.repo.GetPending(ctx, 50, maxOutboxRetries)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		slog.Error("Outbox worker: fetch pending events err:", slog.String("error", err.Error()))
		return
	}

	for _, event := range pendingEvents {
		var pubErr error

		switch event.EventType {
		case constants.EventTypeBookingCreated:
			var bookingCreated events.BookingCreatedEvent
			if err := json.Unmarshal(event.Payload, &bookingCreated); err != nil {
				slog.Error("Outbox worker: deserialize booking created event err:", slog.String("error", err.Error()))
				if markErr := w.repo.MarkFailed(ctx, event.ID, fmt.Sprintf("deserialize error: %v", err)); markErr != nil {
					slog.Error("Outbox worker: mark failed err:", slog.String("error", markErr.Error()))
				}
				continue
			}

			pubErr = w.publisher.PublishBookingCreated(ctx, bookingCreated)

		case constants.EventTypeBookingCancelled:
			var bookingCancelled events.BookingCancelledEvent
			if err := json.Unmarshal(event.Payload, &bookingCancelled); err != nil {
				slog.Error("Outbox worker: deserialize booking cancelled event err:", slog.String("error", err.Error()), slog.Any("payload", event.Payload))
				if markErr := w.repo.MarkFailed(ctx, event.ID, fmt.Sprintf("deserialize error: %v", err)); markErr != nil {
					slog.Error("Outbox worker: mark failed err:", slog.String("error", markErr.Error()))
				}
				continue
			}

			pubErr = w.publisher.PublishBookingCancelled(ctx, bookingCancelled)

		case constants.EventTypeBookingConfirmed:
			var bookingConfirmed events.BookingConfirmedEvent
			if err := json.Unmarshal(event.Payload, &bookingConfirmed); err != nil {
				slog.Error("Outbox worker: deserialize booking confirmed event err:", slog.String("error", err.Error()))
				if markErr := w.repo.MarkFailed(ctx, event.ID, fmt.Sprintf("deserialize error: %v", err)); markErr != nil {
					slog.Error("Outbox worker: mark failed err:", slog.String("error", markErr.Error()))
				}
				continue
			}

			pubErr = w.publisher.PublishBookingConfirmed(ctx, bookingConfirmed)

		case constants.EventTypeBookingPendingPayment:
			var bookingPendingPayment events.BookingPendingPaymentEvent
			if err := json.Unmarshal(event.Payload, &bookingPendingPayment); err != nil {
				slog.Error("Outbox worker: deserialize booking pending payment event err:", slog.String("error", err.Error()))
				if markErr := w.repo.MarkFailed(ctx, event.ID, fmt.Sprintf("deserialize error: %v", err)); markErr != nil {
					slog.Error("Outbox worker: mark failed err:", slog.String("error", markErr.Error()))
				}
				continue
			}

			pubErr = w.publisher.PublishBookingPendingPayment(ctx, bookingPendingPayment)

		default:
			slog.Error("Outbox worker: unknown event type", slog.String("event_type", event.EventType))
			if markErr := w.repo.MarkFailed(ctx, event.ID, fmt.Sprintf("unknown event type: %s", event.EventType)); markErr != nil {
				slog.Error("Outbox worker: mark failed err:", slog.String("error", markErr.Error()))
			}
			continue
		}

		if pubErr != nil {
			slog.Error("Outbox worker: publish event err:", slog.String("error", pubErr.Error()))
			if recordErr := w.repo.RecordFailure(ctx, event.ID, pubErr.Error(), maxOutboxRetries); recordErr != nil {
				slog.Error("Outbox worker: record failure err:", slog.String("error", recordErr.Error()))
			}
			continue
		}

		if err := w.repo.MarkProcessed(ctx, event.ID); err != nil {
			slog.Error("Outbox worker: mark processed err:", slog.String("error", err.Error()))
		}
	}
}

func (w *OutboxWorker) cleanupProcessedEvents(ctx context.Context) {
	processedOlderThan := time.Now().Add(-24 * time.Hour) // Keep processed events for 24 hours
	failedOlderThan := time.Now().Add(-7 * 24 * time.Hour)  // Keep failed events for 7 days
	err := w.repo.DeleteProcessed(ctx, processedOlderThan, failedOlderThan)
	if err != nil {
		slog.Error("Outbox worker: failed to cleanup events", slog.String("error", err.Error()))
		return
	}
	slog.Info("Outbox worker: cleaned up old events successfully")
}
