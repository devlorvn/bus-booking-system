package confirmbooking

import (
	"context"

	userDomain "booking-system/internal/user/domain"

	"github.com/google/uuid"
)



type UserPort interface {
	FindByPhone(
		ctx context.Context,
		phone string,
	) (*userDomain.User, error)

	Create(
		ctx context.Context,
		user *userDomain.User,
	) error

	Update(
		ctx context.Context,
		user *userDomain.User,
	) error
}

type SeatLockPort interface {
	ValidateLockOwner(
		ctx context.Context,
		busID uuid.UUID,
		seatCodes []string,
		tempUserID string,
	) error

	ReleaseSeatLocks(
		ctx context.Context,
		busID uuid.UUID,
		seatCodes []string,
		tempUserID string,
	) error
}

type PaymentEventPublisher interface {
	PublishBookingCreated(
		ctx context.Context,
		bookingID uuid.UUID,
	) error
}
