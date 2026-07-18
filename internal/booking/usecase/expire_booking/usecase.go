package expirebooking

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

type ExpireBookingUsecase struct {
	bookingRepo     ports.BookingRepository
	bookingSeatRepo ports.BookingSeatRepository
	busProvider     ports.BusProvider
	outboxRepo      ports.OutboxRepository
	tx              shared.Transaction
}

func New(
	bookingRepo ports.BookingRepository,
	bookingSeatRepo ports.BookingSeatRepository,
	busProvider ports.BusProvider,
	outboxRepo ports.OutboxRepository,
	tx shared.Transaction,
) *ExpireBookingUsecase {
	return &ExpireBookingUsecase{
		bookingRepo:     bookingRepo,
		bookingSeatRepo: bookingSeatRepo,
		busProvider:     busProvider,
		outboxRepo:      outboxRepo,
		tx:              tx,
	}
}

func (u *ExpireBookingUsecase) Execute(ctx context.Context, bookingID uuid.UUID) (*bookingDomain.Booking, error) {
	var booking *bookingDomain.Booking
	var seats []*bookingDomain.BookingSeat
	var err error

	seats, err = u.bookingSeatRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if len(seats) == 0 {
		return nil, errors.New("SEAT_NOT_FOUND")
	}

	err = u.tx.Execute(ctx, func(txCtx context.Context) error {
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

		// Save BookingCancelledEvent to outbox table inside db transaction
		seatCodes := make([]string, len(seats))
		for i, seat := range seats {
			seatCodes[i] = seat.SeatCode
		}

		cancelEvent := events.BookingCancelledEvent{
			BookingID: bookingID,
			BusID:     booking.BusID,
			SeatCodes: seatCodes,
		}

		eventBytes, err := json.Marshal(cancelEvent)
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

		return nil
	})
	if err != nil {
		return nil, err
	}

	return booking, nil
}
