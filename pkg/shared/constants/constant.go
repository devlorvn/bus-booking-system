package constants

import "time"

const (
	BookingLockTTL                = 15 * time.Minute
	BookingExpirationQueue string = "booking_expiration_queue"
)
