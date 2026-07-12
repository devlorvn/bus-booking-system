package expirebooking

import (
	bookingDomain "booking-system/internal/booking/domain"
	"booking-system/internal/booking/ports"
	"booking-system/pkg/shared"
	"context"
	"errors"

	"github.com/google/uuid"
)

type SeatRepository interface {
	ReleaseSeatsByBookingID(ctx context.Context, bookingID uuid.UUID) error
}

type ExpireBookingUsecase struct {
	bookingRepo ports.BookingRepository
	busProvider ports.BusProvider
	tx          shared.Transaction
}

func New(
	bookingRepo ports.BookingRepository,
	busProvider ports.BusProvider,
	tx shared.Transaction,
) *ExpireBookingUsecase {
	return &ExpireBookingUsecase{
		bookingRepo: bookingRepo,
		busProvider: busProvider,
		tx:          tx,
	}
}

func (u *ExpireBookingUsecase) Execute(ctx context.Context, bookingID uuid.UUID) (*bookingDomain.Booking, error) {
	var booking *bookingDomain.Booking
	err := u.tx.Execute(ctx, func(txCtx context.Context) error {
		var err error
		booking, err = u.bookingRepo.GetByID(txCtx, bookingID)
		if err != nil {
			return err
		}

		if booking.Status != "PENDING_PAYMENT" {
			return errors.New("BOOKING_STATUS_NOT_PENDING")
		}

		booking.Status = "EXPIRED"
		booking.PaymentStatus = "FAILED"

		if err := u.bookingRepo.Update(txCtx, booking); err != nil {
			return err
		}

		if err := u.busProvider.ReleaseSeatsByBookingID(txCtx, bookingID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return booking, nil
}
