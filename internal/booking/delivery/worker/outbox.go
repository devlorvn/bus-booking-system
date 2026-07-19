package worker

import (
	"booking-system/pkg/postgres/repository"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"log/slog"
	"time"
)

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
	pendingEvents, err := w.repo.GetPending(ctx, 50)
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
				_ = w.repo.MarkProcessed(ctx, event.ID)
				continue
			}

			pubErr = w.publisher.PublishBookingCreated(ctx, bookingCreated)

		case constants.EventTypeBookingCancelled:
			var bookingCancelled events.BookingCancelledEvent
			if err := json.Unmarshal(event.Payload, &bookingCancelled); err != nil {
				slog.Error("Outbox worker: deserialize booking cancelled event err:", slog.String("error", err.Error()), slog.Any("payload", event.Payload))
				_ = w.repo.MarkProcessed(ctx, event.ID)
				continue
			}

			pubErr = w.publisher.PublishBookingCancelled(ctx, bookingCancelled)
		case constants.EventTypeBookingConfirmed:
			var bookingConfirmed events.BookingConfirmedEvent
			if err := json.Unmarshal(event.Payload, &bookingConfirmed); err != nil {
				slog.Error("Outbox worker: deserialize booking confirmed event err:", slog.String("error", err.Error()))
				_ = w.repo.MarkProcessed(ctx, event.ID)
				continue
			}

			pubErr = w.publisher.PublishBookingConfirmed(ctx, bookingConfirmed)

		case constants.EventTypeBookingPendingPayment:
			var bookingPendingPayment events.BookingPendingPaymentEvent
			if err := json.Unmarshal(event.Payload, &bookingPendingPayment); err != nil {
				slog.Error("Outbox worker: deserialize booking pending payment event err:", slog.String("error", err.Error()))
				_ = w.repo.MarkProcessed(ctx, event.ID)
				continue
			}

			pubErr = w.publisher.PublishBookingPendingPayment(ctx, bookingPendingPayment)

		default:
			slog.Error("Outbox worker: unknown event type", slog.String("event_type", event.EventType))
			_ = w.repo.MarkProcessed(ctx, event.ID)
			continue
		}

		if pubErr != nil {
			slog.Error("Outbox worker: publish event err:", slog.String("error", pubErr.Error()))
			continue
		}

		if err := w.repo.MarkProcessed(ctx, event.ID); err != nil {
			slog.Error("Outbox worker: mark processed err:", slog.String("error", err.Error()))
		}
	}
}

func (w *OutboxWorker) cleanupProcessedEvents(ctx context.Context) {
	olderThan := time.Now().Add(-24 * time.Hour) // Keep processed events for 24 hours for troubleshooting
	err := w.repo.DeleteProcessed(ctx, olderThan)
	if err != nil {
		slog.Error("Outbox worker: failed to cleanup processed events", slog.String("error", err.Error()))
		return
	}
	slog.Info("Outbox worker: cleaned up processed events successfully")
}
