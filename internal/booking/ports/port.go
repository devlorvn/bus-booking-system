package ports

import (
	bookingDomain "booking-system/internal/booking/domain"
	"context"

	"github.com/google/uuid"
)

type BookingRepository interface {
	Create(
		ctx context.Context,
		booking *bookingDomain.Booking,
	) error

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (*bookingDomain.Booking, error)

	Update(
		ctx context.Context,
		booking *bookingDomain.Booking,
	) error
}
