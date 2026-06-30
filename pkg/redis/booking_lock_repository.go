package redis

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

const bookingLockTTL = 15 * time.Minute

type LockBookingRepository struct {
	client *goredis.Client
}

func NewLockBookingRepository(client *goredis.Client) *LockBookingRepository {
	return &LockBookingRepository{
		client: client,
	}
}

func (r *LockBookingRepository) Create(ctx context.Context, bookingID uuid.UUID, seatCodes []string) error {
	return r.client.SetNX(
		ctx,
		buildBookingLockKey(bookingID, seatCodes),
		bookingID.String(),
		bookingLockTTL,
	).Err()
}

func (r *LockBookingRepository) Release(ctx context.Context, bookingID uuid.UUID, seatCodes []string) error {
	return r.client.Del(
		ctx,
		buildBookingLockKey(bookingID, seatCodes),
	).Err()
}

func buildBookingLockKey(bookingID uuid.UUID, seatCodes []string) string {
	sortedSeatCodes := make([]string, len(seatCodes))
	copy(sortedSeatCodes, seatCodes)
	sort.Strings(sortedSeatCodes)
	seatCodesStr := strings.Join(sortedSeatCodes, ",")
	return fmt.Sprintf("booking_lock:%s:%s", bookingID.String(), seatCodesStr)
}
