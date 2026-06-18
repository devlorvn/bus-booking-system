package usecase

import (
	"booking-system/internal/booking/repository"
)

type ConfirmBookingUsecase struct {
	bookingPort repository.BookingPort
	// seatPort    busReppository.SeatPort
	// busPort     busReppository.BusPort
}
