package provider

import (
	bookingDomain "booking-system/internal/booking/domain"
	userDomain "booking-system/internal/user/domain"
	postgresRepo "booking-system/pkg/postgres/repository"
	"context"

	"github.com/google/uuid"
)

type BookingRepoAdapter struct {
	Repo *postgresRepo.BookingRepository
}

func (a *BookingRepoAdapter) Create(ctx context.Context, booking *bookingDomain.Booking) error {
	_, err := a.Repo.Create(ctx, booking)
	return err
}

func (a *BookingRepoAdapter) GetByID(ctx context.Context, id uuid.UUID) (*bookingDomain.Booking, error) {
	return a.Repo.GetByID(ctx, id)
}

func (a *BookingRepoAdapter) Update(ctx context.Context, booking *bookingDomain.Booking) error {
	_, err := a.Repo.Update(ctx, booking)
	return err
}

type UserPortAdapter struct {
	Repo *postgresRepo.UserRepository
}

func (a *UserPortAdapter) FindByPhone(ctx context.Context, phone string) (*userDomain.User, error) {
	return a.Repo.FindByPhone(ctx, phone)
}

func (a *UserPortAdapter) Create(ctx context.Context, user *userDomain.User) error {
	_, err := a.Repo.Create(ctx, user)
	return err
}

func (a *UserPortAdapter) Update(ctx context.Context, user *userDomain.User) error {
	_, err := a.Repo.Update(ctx, user)
	return err
}

type SeatPortAdapter struct {
	Repo *postgresRepo.SeatRepository
}

func (a *SeatPortAdapter) MarkBookedByBookingID(ctx context.Context, bookingID uuid.UUID) error {
	return a.Repo.MarkBookedByBookingID(ctx, bookingID)
}

type BusPortAdapter struct {
	Repo *postgresRepo.BusRepository
}

func (a *BusPortAdapter) DecrementAvailableSeats(ctx context.Context, busID uuid.UUID, count int) error {
	return a.Repo.DecrementAvailableSeats(ctx, busID, count)
}
