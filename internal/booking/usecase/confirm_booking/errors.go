package confirmbooking

import "errors"

var (
	ErrLockExpired      = errors.New("LOCK_EXPIRED")
	ErrTempUserRequired = errors.New("TEMP_USER_REQUIRED")
	ErrNoSeatSelected   = errors.New("NO_SEAT_SELECTED")
	ErrPhoneRequired    = errors.New("PHONE_REQUIRED")
	ErrBusNotFound      = errors.New("BUS_NOT_FOUND")
	ErrSeatNotFound     = errors.New("SEAT_NOT_FOUND")
	ErrSeatAlreadyBooked = errors.New("SEAT_ALREADY_BOOKED")
)
