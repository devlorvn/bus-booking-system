package service

import (
	busDomain "booking-system/internal/bus/domain"
)

type PricingService interface {
	Calculate(
		bus *busDomain.Bus,
		seats []*busDomain.Seat,
	) (float64, error)
}

type pricingService struct{}

func (s *pricingService) Calculate(
	bus *busDomain.Bus,
	seats []*busDomain.Seat,
) (float64, error) {
	return bus.Price * float64(len(seats)), nil
}

func NewPricingService() PricingService {
	return &pricingService{}
}

