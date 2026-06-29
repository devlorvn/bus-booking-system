package ports

import (
	"context"

	"github.com/google/uuid"
)

type PaymentProcessor interface {
	Process(
		ctx context.Context,
		bookingID uuid.UUID,
	) error
}
