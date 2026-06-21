package ports

import (
	"context"

	busDomain "booking-system/internal/bus/domain"

	"github.com/google/uuid"
)

type BusProvider interface {
	GetBus(
		ctx context.Context,
		busID uuid.UUID,
	) (*busDomain.Bus, error)

	GetSeatsByCodes(
		ctx context.Context,
		busID uuid.UUID,
		codes []string,
	) ([]*busDomain.Seat, error)
}
