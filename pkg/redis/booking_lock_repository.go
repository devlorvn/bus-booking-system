package redis

import (
	"booking-system/pkg/shared/constants"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

type LockBookingRepository struct {
	client *goredis.Client
}

func NewLockBookingRepository(client *goredis.Client) *LockBookingRepository {
	return &LockBookingRepository{
		client: client,
	}
}

func (r *LockBookingRepository) Create(ctx context.Context, bookingID uuid.UUID, seatCodes []string) error {
	key := buildBookingLockKey(bookingID, seatCodes)
	err := r.client.SetNX(
		ctx,
		key,
		bookingID.String(),
		constants.BookingLockTTL,
	).Err()
	if err != nil {
		return err
	}

	expireAt := time.Now().Add(constants.BookingLockTTL).Unix()

	err = r.client.ZAdd(
		ctx,
		constants.BookingExpirationQueue,
		goredis.Z{
			Score:  float64(expireAt),
			Member: key,
		},
	).Err()

	return err
}

func (r *LockBookingRepository) Release(ctx context.Context, bookingID uuid.UUID, seatCodes []string) error {
	key := buildBookingLockKey(bookingID, seatCodes)
	err := r.client.Del(
		ctx,
		key,
	).Err()
	if err != nil {
		return err
	}

	err = r.client.ZRem(
		ctx,
		constants.BookingExpirationQueue,
		key,
	).Err()
	return err
}

func (r *LockBookingRepository) AcquireConfirmLock(ctx context.Context, tempUserID string) error {
	return r.client.SetNX(
		ctx,
		buildConfirmLockKey(tempUserID),
		tempUserID,
		constants.BookingLockTTL,
	).Err()
}

func (r *LockBookingRepository) ReleaseConfirmLock(ctx context.Context, tempUserID string) error {
	return r.client.Del(
		ctx,
		buildConfirmLockKey(tempUserID),
	).Err()
}

func buildBookingLockKey(bookingID uuid.UUID, seatCodes []string) string {
	seatCodesStr := buildSeatCodesStr(seatCodes)
	return fmt.Sprintf("booking_lock:%s:%s", bookingID.String(), seatCodesStr)
}

func buildConfirmLockKey(tempUserID string) string {
	return fmt.Sprintf("confirm_lock:%s", tempUserID)
}

func buildSeatCodesStr(seatCodes []string) string {
	sortedSeatCodes := make([]string, len(seatCodes))
	copy(sortedSeatCodes, seatCodes)
	sort.Strings(sortedSeatCodes)
	return strings.Join(sortedSeatCodes, ",")
}
