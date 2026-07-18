package processsagaresult

import (
	"booking-system/internal/booking/ports"
	"booking-system/pkg/postgres/model"
	"booking-system/pkg/shared"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type ProcessSagaResultUsecase struct {
	bookingRepo ports.BookingRepository
	bookingLock ports.BookingLockPort
	outboxRepo  ports.OutboxRepository
	tx          shared.Transaction
}

func New(
	bookingRepo ports.BookingRepository,
	bookingLock ports.BookingLockPort,
	outboxRepo ports.OutboxRepository,
	tx shared.Transaction,
) *ProcessSagaResultUsecase {
	return &ProcessSagaResultUsecase{
		bookingRepo: bookingRepo,
		bookingLock: bookingLock,
		outboxRepo:  outboxRepo,
		tx:          tx,
	}
}

func (u *ProcessSagaResultUsecase) OnSeatsReserved(ctx context.Context, bookingID uuid.UUID) error {
	return u.tx.Execute(ctx, func(txCtx context.Context) error {
		booking, err := u.bookingRepo.GetByID(txCtx, bookingID)
		if err != nil {
			return err
		}

		if booking.Status != "PENDING_CONFIRMATION" {
			// Already processed or state mismatch
			return nil
		}

		booking.Status = "PENDING_PAYMENT"
		if _, err = u.bookingRepo.Update(txCtx, booking); err != nil {
			return err
		}

		// Save BookingPendingPaymentEvent to outbox table inside db transaction
		eventMsg := events.BookingPendingPaymentEvent{
			BookingID: bookingID,
		}
		eventBytes, err := json.Marshal(eventMsg)
		if err != nil {
			return err
		}

		outboxRecord := &model.Outbox{
			ID:            uuid.New(),
			AggregateType: "booking",
			AggregateID:   bookingID.String(),
			EventType:     constants.EventTypeBookingPendingPayment,
			Payload:       eventBytes,
			Status:        "PENDING",
		}

		return u.outboxRepo.Save(txCtx, outboxRecord)
	})
}

func (u *ProcessSagaResultUsecase) OnSeatsReservationFailed(ctx context.Context, bookingID uuid.UUID, seatCodes []string) error {
	var shouldRelease bool

	err := u.tx.Execute(ctx, func(txCtx context.Context) error {
		booking, err := u.bookingRepo.GetByID(txCtx, bookingID)
		if err != nil {
			return err
		}

		if booking.Status != "PENDING_CONFIRMATION" {
			// Already processed or state mismatch
			return nil
		}

		booking.Status = "FAILED"
		if _, err = u.bookingRepo.Update(txCtx, booking); err != nil {
			return err
		}

		// Save BookingCancelledEvent to outbox table inside db transaction to notify WS
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
			EventType:     constants.EventTypeBookingCancelled,
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
		_ = u.bookingLock.Release(ctx, bookingID, seatCodes)
	}

	return nil
}
