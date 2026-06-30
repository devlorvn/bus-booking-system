package helpers

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

func ParseSeatLockKey(key string) (uuid.UUID, string, error) {
	parts := strings.Split(key, ":")

	if len(parts) != 3 {
		return uuid.Nil, "", errors.New("invalid key")
	}

	busID, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.Nil, "", err
	}

	return busID, parts[2], nil
}

func ParseBookingLockKey(key string) (uuid.UUID, []string, error) {
	parts := strings.Split(key, ":")

	if len(parts) != 3 {
		return uuid.Nil, nil, errors.New("invalid key")
	}

	bookingID, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.Nil, nil, err
	}

	seatCodes := strings.Split(parts[2], ",")
	return bookingID, seatCodes, nil
}
