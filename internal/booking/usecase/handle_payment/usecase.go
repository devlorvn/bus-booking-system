package handlepayment

import (
	bookingDomain "booking-system/internal/booking/domain"
	"booking-system/internal/booking/ports"
	busDomain "booking-system/internal/bus/domain"
	"booking-system/pkg/shared"
	"context"
	"log"

	"github.com/google/uuid"
)

type HandlePaymentSuccessUsecase struct {
	bookingRepo    ports.BookingRepository
	seatRepo       SeatRepository
	busRepo        BusRepository
	bookingLock    ports.BookingLockPort
	eventPublisher ports.EventPublisher
	tx             shared.Transaction
}

func New(
	bookingRepo ports.BookingRepository,
	seatRepo SeatRepository,
	busRepo BusRepository,
	bookingLock ports.BookingLockPort,
	eventPublisher ports.EventPublisher,
	tx shared.Transaction,
) *HandlePaymentSuccessUsecase {
	return &HandlePaymentSuccessUsecase{
		bookingRepo:    bookingRepo,
		seatRepo:       seatRepo,
		busRepo:        busRepo,
		bookingLock:    bookingLock,
		eventPublisher: eventPublisher,
		tx:             tx,
	}
}

func (u *HandlePaymentSuccessUsecase) Success(
	ctx context.Context,
	bookingID uuid.UUID,
) error {
	var booking *bookingDomain.Booking
	var seats []*busDomain.Seat
	var shouldRelease bool

	err := u.tx.Execute(ctx, func(txCtx context.Context) error {
		var err error
		booking, err = u.bookingRepo.GetByID(txCtx, bookingID)
		if err != nil {
			return err
		}

		if booking.Status != "PENDING_PAYMENT" {
			return nil
		}

		seats, err = u.seatRepo.GetSeatByBookingID(txCtx, bookingID)
		if err != nil {
			return err
		}

		booking.Status = "PAID"
		booking.PaymentStatus = "COMPLETED"

		if err := u.bookingRepo.Update(txCtx, booking); err != nil {
			return err
		}
		if err := u.seatRepo.MarkBookedByBookingID(txCtx, bookingID); err != nil {
			return err
		}

		if err := u.busRepo.DecrementAvailableSeats(txCtx, booking.BusID, booking.TotalSeats); err != nil {
			return err
		}

		shouldRelease = true
		return nil
	})
	if err != nil {
		return err
	}

	if shouldRelease {
		seatCodes := make([]string, len(seats))
		for i, seat := range seats {
			seatCodes[i] = seat.SeatCode
		}

		_ = u.bookingLock.Release(ctx, bookingID, seatCodes)
	}

	return nil
}

func (u *HandlePaymentSuccessUsecase) Failed(
	ctx context.Context,
	bookingID uuid.UUID,
) error {
	var booking *bookingDomain.Booking
	var seats []*busDomain.Seat
	var shouldRelease bool

	err := u.tx.Execute(ctx, func(txCtx context.Context) error {
		var err error
		booking, err = u.bookingRepo.GetByID(txCtx, bookingID)
		if err != nil {
			return err
		}

		if booking.Status != "PENDING_PAYMENT" {
			return nil
		}

		seats, err = u.seatRepo.GetSeatByBookingID(txCtx, bookingID)
		if err != nil {
			return err
		}

		booking.Status = "FAILED"
		booking.PaymentStatus = "FAILED"

		if err := u.bookingRepo.Update(txCtx, booking); err != nil {
			return err
		}

		if err := u.seatRepo.ReleaseSeatsByBookingID(txCtx, bookingID); err != nil {
			return err
		}

		shouldRelease = true
		return nil
	})
	if err != nil {
		return err
	}

	if shouldRelease {
		seatCodes := make([]string, len(seats))
		for i, seat := range seats {
			seatCodes[i] = seat.SeatCode
		}

		_ = u.bookingLock.Release(ctx, bookingID, seatCodes)

		for _, seatCode := range seatCodes {
			err := u.eventPublisher.PublishSeatReleased(booking.BusID.String(), seatCode)
			if err != nil {
				log.Printf("[HandlePayment Usecase] Error publishing seat_unlocked event: %v", err)
			}
		}
	}

	return nil
}
