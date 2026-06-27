package lockseat

import (
	"booking-system/internal/booking/dto"
	"booking-system/internal/booking/ports"
	"context"
)

type LockSeatUsecase struct {
	busPort        ports.BusProvider
	seatLockPort   SeatLockPort
	eventPublisher ports.EventPublisher
}

func New(
	busPort ports.BusProvider,
	seatLockPort SeatLockPort,
	eventPublisher ports.EventPublisher,
) *LockSeatUsecase {
	return &LockSeatUsecase{
		busPort:        busPort,
		seatLockPort:   seatLockPort,
		eventPublisher: eventPublisher,
	}
}

func (u *LockSeatUsecase) Execute(
	ctx context.Context,
	input dto.LockSeatRequest,
) (*dto.LockSeatResponse, error) {
	if len(input.TempUserID) == 0 {
		return nil, ErrTempUserIDRequired
	}

	if len(input.SeatCodes) == 0 {
		return nil, ErrNoSeatSelected
	}

	bus, err := u.busPort.GetBus(ctx, input.BusID)
	if err != nil {
		return nil, err
	}
	if bus == nil {
		return nil, ErrBusNotFound
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
		return nil, ErrSomeSeatsNotFound
	}

	for _, seat := range seats {
		if seat.Status == "BOOKED" {
			return nil, ErrSeatAlreadyBooked
		}
	}

	// Lock the seats using the SeatLockPort first
	err = u.seatLockPort.AcquireSeatLocks(
		ctx,
		input.BusID,
		input.SeatCodes,
		input.TempUserID.String(),
	)
	if err != nil {
		return nil, err
	}

	// Broadcast events only after locks are successfully acquired
	for _, seat := range seats {
		_ = u.eventPublisher.PublishSeatLocked(
			input.BusID.String(),
			seat.ID.String(),
			seat.SeatCode,
			input.TempUserID.String(),
		)
	}

	return &dto.LockSeatResponse{Success: true}, nil
}
