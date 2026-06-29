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

func (r *LockSeatRepository) ValidateLockOwner(
	ctx context.Context,
	busID uuid.UUID,
	seatCodes []string,
	tempUserID string,
) error {
	for _, seatCode := range seatCodes {
		key := buildSeatLockKey(busID, seatCode)
		owner, err := r.client.Get(ctx, key).Result()
		if err != nil {
			if err == goredis.Nil {
				return errors.New("LOCK_EXPIRED")
			}
			return err
		}
		if owner != tempUserID {
			return errors.New("LOCK_OWNER_MISMATCH")
		}
	}
	return nil
}


var releaseLockScript = goredis.NewScript(`
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
`)

func (r *LockSeatRepository) ReleaseSeatLocks(
	ctx context.Context,
	busID uuid.UUID,
	seatCodes []string,
	tempUserID string,
) error {
	for _, seatCode := range seatCodes {
		key := buildSeatLockKey(busID, seatCode)
		// To ensure atomicity of the operation, we use a Lua script.
		_, err := releaseLockScript.Run(ctx, r.client, []string{key}, tempUserID).Result()
		if err != nil && !errors.Is(err, goredis.Nil) {
			return err
		}
	}
	return nil
}

func (r *LockSeatRepository) IsSeatLocked(
	ctx context.Context,
	busID uuid.UUID,
	seatCode string,
) (bool, error) {
	key := buildSeatLockKey(busID, seatCode)
	exist, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exist == 1, nil
}

func (r *LockSeatRepository) GetLockedSeats(
	ctx context.Context,
	busID uuid.UUID,
	seatCodes []string,
) (map[string]bool, error) {
	if len(seatCodes) == 0 {
		return nil, nil
	}
	keys := make([]string, len(seatCodes))
	for i, seatCode := range seatCodes {
		keys[i] = buildSeatLockKey(busID, seatCode)
	}

	results, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	lockedSeats := make(map[string]bool)
	for i, val := range results {
		if val != nil {
			lockedSeats[seatCodes[i]] = true
		}
	}
	return lockedSeats, nil
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
