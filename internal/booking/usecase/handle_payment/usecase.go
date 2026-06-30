package handlepayment

import (
	"booking-system/internal/booking/ports"
	"booking-system/pkg/shared"
	"context"

	"github.com/google/uuid"
)

type HandlePaymentSuccessUsecase struct {
	bookingRepo ports.BookingRepository
	seatRepo    SeatRepository
	busRepo     BusRepository
	tx          shared.Transaction
}

func New(
	bookingRepo ports.BookingRepository,
	seatRepo SeatRepository,
	busRepo BusRepository,
	tx shared.Transaction,
) *HandlePaymentSuccessUsecase {
	return &HandlePaymentSuccessUsecase{
		bookingRepo: bookingRepo,
		seatRepo:    seatRepo,
		busRepo:     busRepo,
		tx:          tx,
	}
}

func (u *HandlePaymentSuccessUsecase) Success(
	ctx context.Context,
	bookingID uuid.UUID,
) error {
	return u.tx.Execute(ctx, func(txCtx context.Context) error {
		booking, err := u.bookingRepo.GetByID(txCtx, bookingID)
		if err != nil {
			return err
		}

		if booking.Status != "PENDING_PAYMENT" {
			return nil
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
		return nil
	})
}

func (u *HandlePaymentSuccessUsecase) Failed(
	ctx context.Context,
	bookingID uuid.UUID,
) error {
	return u.tx.Execute(ctx, func(txCtx context.Context) error {
		booking, err := u.bookingRepo.GetByID(txCtx, bookingID)
		if err != nil {
			return err
		}

		if booking.Status != "PENDING_PAYMENT" {
			return nil
		}

		booking.Status = "FAILED"
		booking.PaymentStatus = "FAILED"

		if err := u.bookingRepo.Update(txCtx, booking); err != nil {
			return err
		}
		return nil
	})
}
