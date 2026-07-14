package expirebooking

import (
	bookingDomain "booking-system/internal/booking/domain"
	"booking-system/internal/booking/ports"
	busDomain "booking-system/internal/bus/domain"
	"booking-system/pkg/shared"
	"booking-system/pkg/shared/events"
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
)

type ExpireBookingUsecase struct {
	bookingRepo    ports.BookingRepository
	busProvider    ports.BusProvider
	eventPublisher ports.EventPublisher
	tx             shared.Transaction
}

func New(
	bookingRepo ports.BookingRepository,
	busProvider ports.BusProvider,
	eventPublisher ports.EventPublisher,
	tx shared.Transaction,
) *ExpireBookingUsecase {
	return &ExpireBookingUsecase{
		bookingRepo:    bookingRepo,
		busProvider:    busProvider,
		eventPublisher: eventPublisher,
		tx:             tx,
	}
}

func (u *ExpireBookingUsecase) Execute(ctx context.Context, bookingID uuid.UUID) (*bookingDomain.Booking, error) {
	var booking *bookingDomain.Booking
	var seats []*busDomain.Seat
	var shouldPublish bool

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

		seats, err = u.busProvider.GetSeatByBookingID(txCtx, bookingID)
		if err != nil {
			return err
		}

		shouldPublish = true
		return nil
	})
	if err != nil {
		return nil, err
	}

	if shouldPublish {
		seatCodes := make([]string, len(seats))
		for i, seat := range seats {
			seatCodes[i] = seat.SeatCode
		}

		cancelEvent := events.BookingCancelledEvent{
			BookingID: bookingID,
			BusID:     booking.BusID,
			SeatCodes: seatCodes,
		}

		if err := u.eventPublisher.PublishBookingCancelled(ctx, cancelEvent); err != nil {
			log.Printf("[ExpireBookingUsecase] Error publishing booking_cancelled event: %v", err)
		}
	}

	return booking, nil
}
