package worker

import (
	"booking-system/internal/booking/ports"
	"booking-system/pkg/shared/helpers"
	"context"
	"log"
	"strings"

	goredis "github.com/redis/go-redis/v9"
)

type LockExpirationWorker struct {
	client         *goredis.Client
	eventPublisher ports.EventPublisher
}

func NewLockExpirationWorker(client *goredis.Client, eventPublisher ports.EventPublisher) *LockExpirationWorker {
	return &LockExpirationWorker{
		client:         client,
		eventPublisher: eventPublisher,
	}
}
func (w *LockExpirationWorker) Start(ctx context.Context) error {
	pubsub := w.client.Subscribe(
		ctx,
		"__keyevent@0__:expired",
	)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case msg := <-ch:
			if msg == nil {
				continue
			}

			if !strings.HasPrefix(msg.Payload, "seat_lock:") {
				continue
			}

			log.Println("seat lock expired:", msg.Payload)

			busID, seatCode, err := helpers.ParseSeatLockKey(msg.Payload)
			if err != nil {
				log.Println("parse error:", err)
				continue
			}

			err = w.eventPublisher.PublishSeatReleased(
				busID.String(),
				seatCode,
			)
			if err != nil {
				log.Println("publish error:", err)
			}
		}
	}
}
