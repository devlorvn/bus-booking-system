package lockseat

import (
	"booking-system/internal/booking/dto"
	"context"
	"errors"
)

type LockSeatUsecase struct {
	busPort      BusProvider
	seatLockPort SeatLockPort
	// eventPublisher EventPublisher
}

func NewLockSeatUsecase(
	busPort BusProvider,
	seatLockPort SeatLockPort,
	// eventPublisher EventPublisher,
) *LockSeatUsecase {
	return &LockSeatUsecase{
		busPort:      busPort,
		seatLockPort: seatLockPort,
		// eventPublisher: eventPublisher,
	}
}

func (u *LockSeatUsecase) Execute(
	ctx context.Context,
	input dto.LockSeatRequest,
) (*dto.LockSeatResponse, error) {
	if len(input.TempUserID) == 0 {
		return nil, errors.New("TEMP_USER_ID_REQUIRED")
	}

	if len(input.SeatCodes) == 0 {
		return nil, errors.New("NO_SEAT_SELECTED")
	}

	bus, err := u.busPort.GetBus(ctx, input.BusID)
	if err != nil {
		return nil, err
	}
	if bus == nil {
		return nil, errors.New("BUS_NOT_FOUND")
	}

	seats, err := u.busPort.GetSeatsByCodes(
		ctx,
		input.BusID,
		input.SeatCodes,
	)
	if err != nil {
		return nil, err
	}

	if len(seats) != len(input.SeatCodes) {
		return nil, errors.New("SOME_SEATS_NOT_FOUND")
	}

	for _, seat := range seats {
		if seat.Status == "BOOKED" {
			return nil, errors.New("SEAT_ALREADY_BOOKED")
		}
	}

	// Lock the seats using the SeatLockPort
	err = u.seatLockPort.AcquireSeatLocks(
		ctx,
		input.BusID,
		input.SeatCodes,
		input.TempUserID.String(),
	)
	if err != nil {
		return nil, err
	}
	// broadcast event to notify other services about the locked seats
	// _ = u.eventPublisher.PublishSeatLocked(
	// 	ctx,
	// 	input.BusID,
	// 	input.SeatCodes,
	// )
	return &dto.LockSeatResponse{Success: true}, nil
}
