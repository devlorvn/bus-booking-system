package usecase

import (
	"booking-system/internal/bus/domain"
	"booking-system/internal/bus/repository"
	"booking-system/pkg/shared"
	"context"

	"github.com/google/uuid"
)

type SeatUsecase struct {
	repo repository.SeatRepository
	tx   shared.Transaction
}

func NewSeatUsecase(repo repository.SeatRepository, tx shared.Transaction) *SeatUsecase {
	return &SeatUsecase{repo: repo, tx: tx}
}

func (u *SeatUsecase) Create(ctx context.Context, Seat *domain.Seat) (*domain.Seat, error) {
	return u.repo.Create(ctx, Seat)
}

func (u *SeatUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Seat, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *SeatUsecase) Update(ctx context.Context, Seat *domain.Seat) (*domain.Seat, error) {
	Seat, err := u.repo.GetByID(ctx, Seat.ID)
	if err != nil {
		return nil, err
	}

	return u.repo.Update(ctx, Seat)
}

func (u *SeatUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	Seat, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if Seat == nil {
		return nil
	}
	return u.repo.Delete(ctx, id)
}

func (u *SeatUsecase) ListByBusID(ctx context.Context, busID uuid.UUID) ([]*domain.Seat, error) {
	return u.repo.ListByBusID(ctx, busID)
}

func (u *SeatUsecase) GetByBusAndCodes(ctx context.Context, busID uuid.UUID, codes []string) ([]*domain.Seat, error) {
	return u.repo.GetByBusAndCodes(ctx, busID, codes)
}

func (u *SeatUsecase) BookSeats(ctx context.Context, busID uuid.UUID, codes []string) error {
	return u.repo.BookSeats(ctx, busID, codes)
}

func (u *SeatUsecase) MarkBookedByBookingID(ctx context.Context, bookingID uuid.UUID, busID uuid.UUID, seatCount int) error {
	return u.repo.MarkBookedByBookingID(ctx, bookingID)
}

func (u *SeatUsecase) GetSeatByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*domain.Seat, error) {
	return u.repo.GetSeatByBookingID(ctx, bookingID)
}

func (u *SeatUsecase) ReleaseSeatsByBookingID(ctx context.Context, bookingID uuid.UUID) error {
	return u.repo.ReleaseSeatsByBookingID(ctx, bookingID)
}
