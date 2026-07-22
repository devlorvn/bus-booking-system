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

	// 1. Lock the seats in Redis FIRST to prevent race conditions during DB validation
	err := u.seatLockPort.AcquireSeatLocks(
		ctx,
		input.BusID,
		input.SeatCodes,
		input.TempUserID.String(),
	)
	if err != nil {
		return nil, err
	}

	// Clean up lock if any DB validation fails
	success := false
	defer func() {
		if !success {
			_ = u.seatLockPort.ReleaseSeatLocks(
				ctx,
				input.BusID,
				input.SeatCodes,
				input.TempUserID.String(),
			)
		}
	}()

	// 2. Validate Bus & Seats in DB
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

	// Everything validated successfully; keep the lock!
	success = true

	// 3. Broadcast events only after locks and validations succeed
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
