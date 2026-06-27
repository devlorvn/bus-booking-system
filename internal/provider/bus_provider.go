package provider

import (
	busDomain "booking-system/internal/bus/domain"
	busRepo "booking-system/internal/bus/repository"
	"context"

	"github.com/google/uuid"
)

type BusProvider struct {
	busRepository  busRepo.BusRepository
	seatRepository busRepo.SeatRepository
}

func NewBusProvider(busRepository busRepo.BusRepository, seatRepository busRepo.SeatRepository) *BusProvider {
	return &BusProvider{
		busRepository:  busRepository,
		seatRepository: seatRepository,
	}
}

func (p *BusProvider) GetBus(
	ctx context.Context,
	busID uuid.UUID,
) (*busDomain.Bus, error) {
	return p.busRepository.GetByID(ctx, busID)
}

func (p *BusProvider) GetSeatsByCodes(
	ctx context.Context,
	busID uuid.UUID,
	codes []string,
) ([]*busDomain.Seat, error) {
	return p.seatRepository.GetByBusAndCodes(
		ctx,
		busID,
		codes,
	)
}

func (p *BusProvider) BookSeats(
	ctx context.Context,
	busID uuid.UUID,
	codes []string,
) error {
	return p.seatRepository.BookSeats(ctx, busID, codes)
}
