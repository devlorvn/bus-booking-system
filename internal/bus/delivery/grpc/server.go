package grpc

import (
	"booking-system/internal/bus/dto"
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

func (s *BusGRPCServer) CreateBus(ctx context.Context, req *buspb.CreateBusRequest) (*buspb.CreateBusResponse, error) {
	depTime, err := time.Parse(time.RFC3339, req.DepartureTime)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid departure time: %v", err)
	}
	dtoReq := dto.CreateBusRequest{
		LicensePlate:  req.LicensePlate,
		FromLocation:  req.FromLocation,
		ToLocation:    req.ToLocation,
		DepartureTime: depTime,
		Price:         req.Price,
		RowName:       req.RowName,
		SeatsPerRow:   int(req.SeatsPerRow),
	}
	bus, err := s.busUsecase.Create(ctx, dtoReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create bus: %v", err)
	}
	return &buspb.CreateBusResponse{Bus: &buspb.Bus{
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

func (s *BusGRPCServer) ListBuses(ctx context.Context, req *buspb.ListBusesRequest) (*buspb.ListBusesResponse, error) {
	buses, err := s.busUsecase.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list buses: %v", err)
	}
	pbBuses := make([]*buspb.Bus, 0, len(buses))
	for _, bus := range buses {
		pbBuses = append(pbBuses, &buspb.Bus{
			Id:             bus.ID.String(),
			LicensePlate:   bus.LicensePlate,
			FromLocation:   bus.FromLocation,
			ToLocation:     bus.ToLocation,
			DepartureTime:  bus.DepartureTime.Format(time.RFC3339),
			TotalSeats:     int32(bus.TotalSeats),
			AvailableSeats: int32(bus.AvailableSeats),
			Price:          bus.Price,
		})
	}
	return &buspb.ListBusesResponse{Buses: pbBuses}, nil
}
func (s *BusGRPCServer) GetSeats(ctx context.Context, req *buspb.GetSeatsRequest) (*buspb.GetSeatsResponse, error) {
	busUUID, err := uuid.Parse(req.BusId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid bus ID: %v", err)
	}
	seats, err := s.busUsecase.GetSeats(ctx, busUUID)
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
	return &buspb.GetSeatsResponse{Seats: pbSeats}, nil
}
func (s *BusGRPCServer) DeleteBus(ctx context.Context, req *buspb.DeleteBusRequest) (*buspb.DeleteBusResponse, error) {
	busUUID, err := uuid.Parse(req.BusId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid bus ID: %v", err)
	}
	err = s.busUsecase.Delete(ctx, busUUID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete bus: %v", err)
	}
	return &buspb.DeleteBusResponse{Success: true}, nil
}

func (s *BusGRPCServer) MarkBookedByBookingID(
	ctx context.Context,
	req *buspb.MarkBookedByBookingIDRequest,
) (*buspb.MarkBookedByBookingIDResponse, error) {
	bookingUUID, err := uuid.Parse(req.BookingId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid bus ID: %v", err)
	}

	busUUID, err := uuid.Parse(req.BusId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid bus ID: %v", err)
	}
	dtoReq := dto.MarkBookedRequest{
		BookingID: bookingUUID,
		BusID:     busUUID,
		SeatCount: int(req.Count),
	}
	err = s.busUsecase.MarkBookedByBookingID(ctx, dtoReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark booked: %v", err)
	}
	return &buspb.MarkBookedByBookingIDResponse{Success: true}, nil
}

func (s *BusGRPCServer) ReleaseSeatsByBookingID(
	ctx context.Context,
	req *buspb.ReleaseSeatsByBookingIDRequest,
) (*buspb.ReleaseSeatsByBookingIDResponse, error) {
	bookingUUID, err := uuid.Parse(req.BookingId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid bus ID: %v", err)
	}
	err = s.seatUsecase.ReleaseSeatsByBookingID(ctx, bookingUUID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to release seats: %v", err)
	}
	return &buspb.ReleaseSeatsByBookingIDResponse{Success: true}, nil
}

func (s *BusGRPCServer) GetSeatByBookingID(
	ctx context.Context,
	req *buspb.GetSeatByBookingIDRequest,
) (*buspb.GetSeatByBookingIDResponse, error) {
	bookingUUID, err := uuid.Parse(req.BookingId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid bus ID: %v", err)
	}
	seats, err := s.seatUsecase.GetSeatByBookingID(ctx, bookingUUID)
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
	return &buspb.GetSeatByBookingIDResponse{Seats: pbSeats}, nil
}
