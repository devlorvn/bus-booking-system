package worker

import (
	"booking-system/internal/booking/event"
	"context"
	"log"
)

type DLQWorker struct {
	bus *event.PaymentEventBus
}

func NewDLQWorker(bus *event.PaymentEventBus) *DLQWorker {
	return &DLQWorker{
		bus: bus,
	}
}

func (w *DLQWorker) Start(ctx context.Context) error {
	log.Println("DLQ worker started")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case dlqEvent := <-w.bus.DLQ:
			log.Printf(
				"[DLQ] Event=%s Entity=%s Error=%s Time=%s",
				dlqEvent.EventType,
				dlqEvent.EntityId,
				dlqEvent.Error,
				dlqEvent.FailedAt,
			)
		}
	}
}
