package constants

import "time"

const (
	BookingLockTTL                = 15 * time.Minute
	BookingExpirationQueue string = "booking_expiration_queue"
)

const WsChanel = "ws:events"

const (
	MaxKafkaEventRetry = 5
	DelayRetry         = 1 * time.Second
)
