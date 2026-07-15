package worker

import (
	"booking-system/internal/booking/ports"
	expirebooking "booking-system/internal/booking/usecase/expire_booking"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/helpers"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

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
	slog.Info("Lock expiration worker started")

	pubsub := w.client.Subscribe(
		ctx,
		"__keyevent@0__:expired",
	)
	defer pubsub.Close()

	ch := pubsub.Channel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			w.pollingBookingExpirations(ctx)
		case msg, ok := <-ch:
			if !ok {
				slog.Error("Redis pubsub channel closed, stopping worker.")
				return errors.New("redis pubsub channel closed")
			}
			switch {
			case strings.HasPrefix(msg.Payload, "seat_lock:"):
				slog.Info("seat lock expired:", slog.String("payload", msg.Payload))

				busID, seatCode, err := helpers.ParseSeatLockKey(msg.Payload)
				if err != nil {
					slog.Error("parse error:", slog.String("error", err.Error()))
					continue
				}

				err = w.eventPublisher.PublishSeatReleased(
					busID.String(),
					seatCode,
				)
				if err != nil {
					slog.Error("publish error:", slog.String("error", err.Error()))
				}
			default:
				continue
			}
		}
	}
}

func (w *LockExpirationWorker) pollingBookingExpirations(ctx context.Context) {
	members, err := w.client.ZRangeByScoreWithScores(ctx, constants.BookingExpirationQueue, &goredis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%d", time.Now().Unix()),
	}).Result()

	if err != nil {
		if ctx.Err() != nil {
			return
		}
		slog.Error("Lock expiration worker: zrange by score error:", slog.String("error", err.Error()))
		return
	}

	for _, member := range members {
		bookingID, seatCodes, err := helpers.ParseBookingLockKey(member.Member.(string))
		if err != nil {
			slog.Error("Lock expiration worker: parse err:", slog.String("error", err.Error()))
			continue
		}

		booking, err := w.expireBooking.Execute(ctx, bookingID)
		if err != nil {
			if err.Error() == "BOOKING_STATUS_NOT_PENDING" {
				err := w.client.ZRem(ctx, constants.BookingExpirationQueue, member.Member).Err()
				if err != nil {
					slog.Error("Lock expiration worker: zrem err:", slog.String("error", err.Error()))
				}
				continue
			}

			if ctx.Err() != nil {
				return
			}
			slog.Error("Lock expiration worker: expire err:", slog.String("error", err.Error()))

			continue
		}

		for _, seatCode := range seatCodes {
			err = w.eventPublisher.PublishSeatReleased(booking.BusID.String(), seatCode)
			if err != nil {
				slog.Error("Lock expiration worker: publish seat released err:", slog.String("error", err.Error()))
			}
		}
		err = w.client.ZRem(ctx, constants.BookingExpirationQueue, member.Member).Err()
		if err != nil {
			slog.Error("Lock expiration worker: zrem err:", slog.String("error", err.Error()))
		}
	}
}
