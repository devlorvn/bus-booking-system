package lockseat

import "errors"

var (
	ErrBusNotFound       = errors.New("BUS_NOT_FOUND")
	ErrSeatNotFound      = errors.New("SEAT_NOT_FOUND")
	ErrSeatAlreadyBooked = errors.New("SEAT_ALREADY_BOOKED")
	ErrSeatAlreadyLocked = errors.New("SEAT_ALREADY_LOCKED")
	ErrNoSeatSelected    = errors.New("NO_SEAT_SELECTED")
)
