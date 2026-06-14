package usecase

import (
	"booking-system/internal/booking/repository"
	busReppository "booking-system/internal/bus/repository"
)

type ConfirmBookingUsecase struct {
	bookingPort repository.BookingPort
	seatPort    busReppository.SeatPort
	busPort     busReppository.BusPort
}
