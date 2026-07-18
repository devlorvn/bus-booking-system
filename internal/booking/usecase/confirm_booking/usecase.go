package confirmbooking

import (
	bookingDomain "booking-system/internal/booking/domain"
	"booking-system/internal/booking/dto"
	"booking-system/internal/booking/ports"
	"booking-system/internal/booking/service"
	userDomain "booking-system/internal/user/domain"
	"booking-system/pkg/postgres/model"
	"booking-system/pkg/shared"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

const (
	BOOKING_STATUS_PENDING_PAYMENT = "PENDING_PAYMENT"
	BOOKING_STATUS_PAID            = "PAID"
	BOOKING_STATUS_FAILED          = "FAILED"
	BOOKING_STATUS_CANCELLED       = "CANCELLED"
	BOOKING_STATUS_EXPIRED         = "EXPIRED"
)

type ConfirmBookingUsecase struct {
	bookingRepo     ports.BookingRepository
	bookingSeatRepo ports.BookingSeatRepository
	userPort        UserPort
	seatLockPort    SeatLockPort
	busProvider     ports.BusProvider
	pricingService  service.PricingService
	tx              shared.Transaction
	outboxRepo      ports.OutboxRepository
	bookingLockPort ports.BookingLockPort
}

func New(
	bookingRepo ports.BookingRepository,
	bookingSeatRepo ports.BookingSeatRepository,
	userPort UserPort,
	seatLockPort SeatLockPort,
	busProvider ports.BusProvider,
	pricingService service.PricingService,
	tx shared.Transaction,
	bookingLockPort ports.BookingLockPort,
	outboxRepo ports.OutboxRepository,
) *ConfirmBookingUsecase {
	return &ConfirmBookingUsecase{
		bookingRepo:     bookingRepo,
		bookingSeatRepo: bookingSeatRepo,
		userPort:        userPort,
		seatLockPort:    seatLockPort,
		pricingService:  pricingService,
		tx:              tx,
		outboxRepo:      outboxRepo,
		busProvider:     busProvider,
		bookingLockPort: bookingLockPort,
	}
}

func (u *ConfirmBookingUsecase) Execute(
	ctx context.Context,
	req dto.ConfirmBookingRequest,
) (*dto.ConfirmBookingResponse, error) {

	err := u.bookingLockPort.AcquireConfirmLock(ctx, req.TempUserID)
	if err != nil {
		return nil, err
	}

	defer u.bookingLockPort.ReleaseConfirmLock(
		ctx,
		req.TempUserID,
	)

	if req.TempUserID == "" {
		return nil, ErrTempUserRequired
	}

	if req.PhoneNumber == "" {
		return nil, ErrPhoneRequired
	}

	if len(req.SeatCodes) == 0 {
		return nil, ErrNoSeatSelected
	}

	err = u.seatLockPort.ValidateLockOwner(
		ctx,
		req.BusID,
		req.SeatCodes,
		req.TempUserID,
	)
	if err != nil {
		return nil, ErrLockExpired
	}

	bus, err := u.busProvider.GetBus(
		ctx,
		req.BusID,
	)
	if err != nil {
		return nil, err
	}

	if bus == nil {
		return nil, ErrBusNotFound
	}

	seats, err := u.busProvider.GetSeatsByCodes(
		ctx,
		req.BusID,
		req.SeatCodes,
	)
	if err != nil {
		return nil, err
	}

	if len(seats) != len(req.SeatCodes) {
		return nil, ErrSeatNotFound
	}

	totalAmount, err := u.pricingService.Calculate(
		bus,
		seats,
	)
	if err != nil {
		return nil, err
	}

	user, err := u.userPort.FindByPhone(
		ctx,
		req.PhoneNumber,
	)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user = &userDomain.User{
			ID:          uuid.New(),
			Name:        req.Name,
			Email:       req.Email,
			PhoneNumber: req.PhoneNumber,
		}

		if err := u.userPort.Create(ctx, user); err != nil {
			return nil, err
		}
	} else {
		user.Name = req.Name
		user.Email = req.Email

		if err := u.userPort.Update(ctx, user); err != nil {
			return nil, err
		}
	}

	var bookingID uuid.UUID
	err = u.tx.Execute(ctx, func(txCtx context.Context) error {
		bookingID = uuid.New()
		booking := &bookingDomain.Booking{
			ID:            bookingID,
			BusID:         req.BusID,
			UserID:        user.ID,
			Status:        "PENDING_CONFIRMATION",
			PaymentStatus: "PENDING",
			TotalSeats:    len(req.SeatCodes),
			TotalAmount:   totalAmount,
		}

		if _, err := u.bookingRepo.Create(txCtx, booking); err != nil {
			return err
		}
		items := make(
			[]*bookingDomain.BookingSeat,
			0,
			len(seats),
		)
		for _, seat := range seats {
			items = append(items, &bookingDomain.BookingSeat{
				BookingID: bookingID,
				SeatID:    seat.ID,
				SeatCode:  seat.SeatCode,
			})
		}

		err = u.bookingSeatRepo.BulkCreate(
			txCtx,
			items,
		)
		if err != nil {
			return err
		}

		// Save BookingCreatedEvent to outbox table inside db transaction
		eventMsg := events.BookingCreatedEvent{
			BookingID: bookingID,
			BusID:     req.BusID,
			SeatCodes: req.SeatCodes,
		}
		eventBytes, err := json.Marshal(eventMsg)
		if err != nil {
			return err
		}

		outboxRecord := &model.Outbox{
			ID:            uuid.New(),
			AggregateType: "booking",
			AggregateID:   bookingID.String(),
			EventType:     constants.EventTypeBookingCreated,
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

	u.seatLockPort.ReleaseSeatLocks(
		ctx,
		req.BusID,
		req.SeatCodes,
		req.TempUserID,
	)

	u.bookingLockPort.Create(ctx, bookingID, req.SeatCodes)

	return &dto.ConfirmBookingResponse{
		BookingID:   bookingID,
		Status:      "PENDING_CONFIRMATION",
		TotalAmount: totalAmount,
	}, nil
}
