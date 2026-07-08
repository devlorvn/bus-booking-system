package grpc

import (
	"booking-system/internal/bus/usecase"
	buspb "booking-system/proto/bus/v1"
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BusGRPCServer struct {
	buspb.UnimplementedBusServiceServer

	busUsecase  *usecase.BusUsecase
	seatUsecase *usecase.SeatUsecase
}

func NewBusGRPCServer(
	busUsecase *usecase.BusUsecase,
	seatUsecase *usecase.SeatUsecase,
) *BusGRPCServer {
	return &BusGRPCServer{
		busUsecase:  busUsecase,
		seatUsecase: seatUsecase,
	}
}

func (s *BusGRPCServer) GetBus(
	ctx context.Context,
	req *buspb.GetBusRequest,
) (*buspb.GetBusResponse, error) {
	busUUID, err := uuid.Parse(req.BusId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid bus ID: %v", err)
	}

	bus, err := s.busUsecase.GetByID(ctx, busUUID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get bus: %v", err)
	}

	return &buspb.GetBusResponse{Bus: &buspb.Bus{
		Id:             bus.ID.String(),
		LicensePlate:   bus.LicensePlate,
		FromLocation:   bus.FromLocation,
		ToLocation:     bus.ToLocation,
		DepartureTime:  bus.DepartureTime.Format(time.RFC3339),
		TotalSeats:     int32(bus.TotalSeats),
		AvailableSeats: int32(bus.AvailableSeats),
		Price:          bus.Price,
	}}, nil
}

func (s *BusGRPCServer) GetSeatsByCodes(ctx context.Context, req *buspb.GetSeatsByCodesRequest) (*buspb.GetSeatsByCodesResponse, error) {
	busUUID, err := uuid.Parse(req.BusId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid bus id format: %v", err)
	}
	seats, err := s.seatUsecase.GetByBusAndCodes(ctx, busUUID, req.SeatCodes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get seats: %v", err)
	}
	pbSeats := make([]*buspb.Seat, 0, len(seats))
	for _, seat := range seats {
		pbSeats = append(pbSeats, &buspb.Seat{
			Id:       seat.ID.String(),
			BusId:    seat.BusID.String(),
			SeatCode: seat.SeatCode,
			Status:   seat.Status,
		})
	}
	return &buspb.GetSeatsByCodesResponse{
		Seats: pbSeats,
	}, nil
}

func (s *BusGRPCServer) BookSeats(ctx context.Context, req *buspb.BookSeatsRequest) (*buspb.BookSeatsResponse, error) {
	busUUID, err := uuid.Parse(req.BusId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid bus id format: %v", err)
	}

	err = s.seatUsecase.BookSeats(ctx, busUUID, req.SeatCodes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to book seats: %v", err)
	}

	return &buspb.BookSeatsResponse{Success: true}, nil
}
