package provider

import (
	busDomain "booking-system/internal/bus/domain"
	buspb "booking-system/proto/bus/v1"
	"context"
	"time"

	"github.com/google/uuid"
)

type BusProvider struct {
	client buspb.BusServiceClient
}

func NewBusProvider(client buspb.BusServiceClient) *BusProvider {
	return &BusProvider{
		client: client,
	}
}

func (p *BusProvider) GetBus(
	ctx context.Context,
	busID uuid.UUID,
) (*busDomain.Bus, error) {
	resp, err := p.client.GetBus(ctx, &buspb.GetBusRequest{
		BusId: busID.String(),
	})
	if err != nil {
		return nil, err
	}
	depTime, err := time.Parse(time.RFC3339, resp.Bus.DepartureTime)
	if err != nil {
		return nil, err
	}
	return &busDomain.Bus{
		ID:             busID,
		LicensePlate:   resp.Bus.LicensePlate,
		DepartureTime:  depTime,
		FromLocation:   resp.Bus.FromLocation,
		ToLocation:     resp.Bus.ToLocation,
		Price:          resp.Bus.Price,
		TotalSeats:     int(resp.Bus.TotalSeats),
		AvailableSeats: int(resp.Bus.AvailableSeats),
	}, nil
}

func (p *BusProvider) GetSeatsByCodes(
	ctx context.Context,
	busID uuid.UUID,
	codes []string,
) ([]*busDomain.Seat, error) {
	resp, err := p.client.GetSeatsByCodes(ctx, &buspb.GetSeatsByCodesRequest{
		BusId:     busID.String(),
		SeatCodes: codes,
	})
	if err != nil {
		return nil, err
	}
	seats := make([]*busDomain.Seat, len(resp.Seats))
	for i, s := range resp.Seats {
		seatUUID, err := uuid.Parse(s.Id)
		if err != nil {
			return nil, err
		}

		busUUID, err := uuid.Parse(s.BusId)
		if err != nil {
			return nil, err
		}

		seats[i] = &busDomain.Seat{
			ID:       seatUUID,
			BusID:    busUUID,
			SeatCode: s.SeatCode,
			Status:   s.Status,
		}
	}
	return seats, nil
}

func (p *BusProvider) BookSeats(
	ctx context.Context,
	busID uuid.UUID,
	codes []string,
) error {
	_, err := p.client.BookSeats(ctx, &buspb.BookSeatsRequest{
		BusId:     busID.String(),
		SeatCodes: codes,
	})
	return err
}

func (p *BusProvider) MarkBookedByBookingID(
	ctx context.Context,
	bookingID uuid.UUID,
	busID uuid.UUID,
	count int,
) error {
	_, err := p.client.MarkBookedByBookingID(ctx, &buspb.MarkBookedByBookingIDRequest{
		BookingId: bookingID.String(),
		BusId:     busID.String(),
		Count:     int32(count),
	})
	return err
}

func (p *BusProvider) GetSeatByBookingID(
	ctx context.Context,
	bookingID uuid.UUID,
) ([]*busDomain.Seat, error) {
	resp, err := p.client.GetSeatByBookingID(ctx, &buspb.GetSeatByBookingIDRequest{
		BookingId: bookingID.String(),
	})
	if err != nil {
		return nil, err
	}
	seats := make([]*busDomain.Seat, len(resp.Seats))
	for i, s := range resp.Seats {
		seatUUID, err := uuid.Parse(s.Id)
		if err != nil {
			return nil, err
		}
		seats[i] = &busDomain.Seat{
			ID:       seatUUID,
			BusID:    uuid.MustParse(s.BusId),
			SeatCode: s.SeatCode,
			Status:   s.Status,
		}
	}
	return seats, nil
}
