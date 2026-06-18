package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

const lockTTL = 5 * time.Minute

type LockSeatRepository struct {
	client *goredis.Client
}

func NewLockSeatRepository(client *goredis.Client) *LockSeatRepository {
	return &LockSeatRepository{
		client: client,
	}
}

func (r *LockSeatRepository) AcquireSeatLocks(
	ctx context.Context,
	busID uuid.UUID,
	seatCodes []string,
	tempUserID string,
) error {
	var locked []string

	for _, seatCode := range seatCodes {
		key := buildSeatLockKey(busID, seatCode)
		existing, err := r.client.Get(ctx, key).Result()
		if err == nil {
			if existing == tempUserID {
				r.client.Expire(ctx, key, lockTTL)
				continue
			}

			_ = r.ReleaseSeatLocks(
				ctx,
				busID,
				locked,
				tempUserID,
			)

			return errors.New("SEAT_ALREADY_LOCKED")
		}

		ok, err := r.client.SetNX(
			ctx,
			key,
			tempUserID,
			lockTTL,
		).Result()

		if err != nil {
			return err
		}

		if !ok {
			_ = r.ReleaseSeatLocks(
				ctx,
				busID,
				locked,
				tempUserID,
			)
			return errors.New("SEAT_ALREADY_LOCKED")
		}

		locked = append(locked, seatCode)

	}
	return nil
}

func (r *LockSeatRepository) ReleaseSeatLocks(
	ctx context.Context,
	busID uuid.UUID,
	seatCodes []string,
	tempUserID string,
) error {
	for _, seatCode := range seatCodes {
		key := buildSeatLockKey(busID, seatCode)
		owner, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		if owner != tempUserID {
			continue
		}
		_, err = r.client.Del(ctx, key).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func buildSeatLockKey(
	busID uuid.UUID,
	seatCode string,
) string {
	return fmt.Sprintf(
		"seat_lock:%s:%s",
		busID.String(),
		seatCode,
	)
}
