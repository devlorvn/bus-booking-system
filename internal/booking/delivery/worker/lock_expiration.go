package worker

import (
	"booking-system/internal/booking/ports"
	expirebooking "booking-system/internal/booking/usecase/expire_booking"
	"booking-system/pkg/shared/helpers"
	"context"
	"errors"
	"log"
	"strings"

	goredis "github.com/redis/go-redis/v9"
)

type LockExpirationWorker struct {
	client         *goredis.Client
	eventPublisher ports.EventPublisher
	expireBooking  *expirebooking.ExpireBookingUsecase
}

func NewLockExpirationWorker(
	client *goredis.Client,
	eventPublisher ports.EventPublisher,
	expireBooking *expirebooking.ExpireBookingUsecase,
) *LockExpirationWorker {
	return &LockExpirationWorker{
		client:         client,
		eventPublisher: eventPublisher,
		expireBooking:  expireBooking,
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

		case msg, ok := <-ch:
			if !ok {
				log.Println("Redis pubsub channel closed, stopping worker.")
				return errors.New("redis pubsub channel closed")
			}
			switch {
			case strings.HasPrefix(msg.Payload, "seat_lock:"):
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
			case strings.HasPrefix(msg.Payload, "booking_lock:"):
				log.Println("booking lock expired:", msg.Payload)

				bookingID, seatCodes, err := helpers.ParseBookingLockKey(msg.Payload)
				if err != nil {
					log.Println("parse error:", err)
					continue
				}

				booking, err := w.expireBooking.Execute(ctx, bookingID)
				if err != nil {
					log.Println("expire booking error:", err)
					continue
				}

				for _, seatCode := range seatCodes {
					err = w.eventPublisher.PublishSeatReleased(
						booking.BusID.String(),
						seatCode,
					)
					if err != nil {
						log.Println("publish error:", err)
					}
				}
			default:
				continue
			}
		}
	}
}
