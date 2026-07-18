package handlepayment

import (
	bookingDomain "booking-system/internal/booking/domain"
	"booking-system/internal/booking/ports"
	"booking-system/pkg/postgres/model"
	"booking-system/pkg/shared"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type HandlePaymentUsecase struct {
	bookingRepo     ports.BookingRepository
	bookingSeatRepo ports.BookingSeatRepository
	busProvider     ports.BusProvider
	bookingLock     ports.BookingLockPort
	outboxRepo      ports.OutboxRepository
	tx              shared.Transaction
}

func New(
	bookingRepo ports.BookingRepository,
	bookingSeatRepo ports.BookingSeatRepository,
	busProvider ports.BusProvider,
	bookingLock ports.BookingLockPort,
	outboxRepo ports.OutboxRepository,
	tx shared.Transaction,
) *HandlePaymentUsecase {
	return &HandlePaymentUsecase{
		bookingRepo:     bookingRepo,
		bookingSeatRepo: bookingSeatRepo,
		busProvider:     busProvider,
		bookingLock:     bookingLock,
		outboxRepo:      outboxRepo,
		tx:              tx,
	}
}

func (u *HandlePaymentUsecase) Success(
	ctx context.Context,
	bookingID uuid.UUID,
) error {
	var booking *bookingDomain.Booking
	var seats []*bookingDomain.BookingSeat
	var shouldRelease bool
	var err error

	seats, err = u.bookingSeatRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return err
	}

	if len(seats) == 0 {
		return errors.New("SEAT_NOT_FOUND")
	}

	err = u.tx.Execute(ctx, func(txCtx context.Context) error {
		var err error
		booking, err = u.bookingRepo.GetByID(txCtx, bookingID)
		if err != nil {
			return err
		}

		if booking.Status != "PENDING_PAYMENT" {
			return nil
		}

		booking.Status = "PAID"
		booking.PaymentStatus = "COMPLETED"

		if _, err := u.bookingRepo.Update(txCtx, booking); err != nil {
			return err
		}

		shouldRelease = true
		return nil
	})
	if err != nil {
		return err
	}

	if err := u.busProvider.MarkBookedByBookingID(ctx, bookingID, booking.BusID, booking.TotalSeats); err != nil {
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

func (u *HandlePaymentUsecase) Failed(
	ctx context.Context,
	bookingID uuid.UUID,
) error {
	var booking *bookingDomain.Booking
	var seats []*bookingDomain.BookingSeat
	var shouldRelease bool
	var err error

	seats, err = u.bookingSeatRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return err
	}

	if len(seats) == 0 {
		return errors.New("SEAT_NOT_FOUND")
	}

	err = u.tx.Execute(ctx, func(txCtx context.Context) error {
		var err error
		booking, err = u.bookingRepo.GetByID(txCtx, bookingID)
		if err != nil {
			return err
		}

		if booking.Status != "PENDING_PAYMENT" {
			return nil
		}

		booking.Status = "FAILED"
		booking.PaymentStatus = "FAILED"

		if _, err := u.bookingRepo.Update(txCtx, booking); err != nil {
			return err
		}

		// Save BookingCancelledEvent to outbox table inside db transaction
		seatCodes := make([]string, len(seats))
		for i, seat := range seats {
			seatCodes[i] = seat.SeatCode
		}

		canclEvent := events.BookingCancelledEvent{
			BookingID: bookingID,
			BusID:     booking.BusID,
			SeatCodes: seatCodes,
		}

		eventBytes, err := json.Marshal(canclEvent)
		if err != nil {
			return err
		}

		outboxRecord := &model.Outbox{
			ID:            uuid.New(),
			AggregateType: "booking",
			AggregateID:   bookingID.String(),
			EventType:     "booking_cancelled",
			Payload:       eventBytes,
			Status:        "PENDING",
		}

		if err := u.outboxRepo.Save(txCtx, outboxRecord); err != nil {
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
