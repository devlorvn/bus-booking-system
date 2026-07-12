package grpc

import (
	"booking-system/internal/booking/dto"
	confirmbooking "booking-system/internal/booking/usecase/confirm_booking"
	lockseat "booking-system/internal/booking/usecase/lock_seat"
	bookingpb "booking-system/proto/booking/v1"
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BookingGRPCServer struct {
	bookingpb.UnimplementedBookingServiceServer
	confirmBookingUsecase *confirmbooking.ConfirmBookingUsecase
	lockSeatUsecase       *lockseat.LockSeatUsecase
}

func NewBookingGRPCServer(
	confirmBookingUsecase *confirmbooking.ConfirmBookingUsecase,
	lockSeatUsecase *lockseat.LockSeatUsecase,
) *BookingGRPCServer {
	return &BookingGRPCServer{
		confirmBookingUsecase: confirmBookingUsecase,
		lockSeatUsecase:       lockSeatUsecase,
	}
}

func (s *BookingGRPCServer) ConfirmBooking(ctx context.Context, req *bookingpb.ConfirmBookingRequest) (*bookingpb.ConfirmBookingResponse, error) {
	busUUID, err := uuid.Parse(req.BusId)
	if err != nil {
		return nil, err
	}

	dtoReq := dto.ConfirmBookingRequest{
		BusID:       busUUID,
		TempUserID:  req.TempUserId,
		SeatCodes:   req.SeatCodes,
		Name:        req.UserName,
		Email:       req.Email,
		PhoneNumber: req.Phone,
	}

	resp, err := s.confirmBookingUsecase.Execute(ctx, dtoReq)

	if err != nil {
		// 4. Ánh xạ các lỗi miền nghiệp vụ sang mã lỗi gRPC tương ứng
		switch {
		case errors.Is(err, confirmbooking.ErrTempUserRequired):
			return nil, status.Errorf(codes.InvalidArgument, "TEMP_USER_REQUIRED")
		case errors.Is(err, confirmbooking.ErrPhoneRequired):
			return nil, status.Errorf(codes.InvalidArgument, "PHONE_REQUIRED")
		case errors.Is(err, confirmbooking.ErrNoSeatSelected):
			return nil, status.Errorf(codes.InvalidArgument, "NO_SEAT_SELECTED")
		case errors.Is(err, confirmbooking.ErrBusNotFound):
			return nil, status.Errorf(codes.NotFound, "BUS_NOT_FOUND")
		case errors.Is(err, confirmbooking.ErrSeatNotFound):
			return nil, status.Errorf(codes.AlreadyExists, "SEAT_NOT_FOUND")
		case errors.Is(err, confirmbooking.ErrSeatAlreadyBooked):
			return nil, status.Errorf(codes.AlreadyExists, "SEAT_ALREADY_BOOKED")
		case errors.Is(err, confirmbooking.ErrLockExpired):
			return nil, status.Errorf(codes.Aborted, "LOCK_EXPIRED")
		default:
			return nil, status.Errorf(codes.Internal, "failed to confirm booking: %v", err)
		}
	}
	return &bookingpb.ConfirmBookingResponse{
		BookingId:   resp.BookingID.String(),
		TotalAmount: resp.TotalAmount,
		Status:      resp.Status,
	}, nil
}

func (s *BookingGRPCServer) LockSeat(ctx context.Context, req *bookingpb.LockSeatRequest) (*bookingpb.LockSeatResponse, error) {
	busUUID, err := uuid.Parse(req.BusId)
	if err != nil {
		return nil, err
	}

	temUserUUID, err := uuid.Parse(req.TempUserId)
	if err != nil {
		return nil, err
	}

	lockSeatDto := dto.LockSeatRequest{
		BusID:      busUUID,
		TempUserID: temUserUUID,
		SeatCodes:  req.SeatCodes,
	}

	_, err = s.lockSeatUsecase.Execute(ctx, lockSeatDto)
	if err != nil {
		switch {
		case errors.Is(err, lockseat.ErrBusNotFound):
			return nil, status.Errorf(codes.NotFound, "BUS_NOT_FOUND")
		case errors.Is(err, lockseat.ErrNoSeatSelected):
			return nil, status.Errorf(codes.InvalidArgument, "NO_SEAT_SELECTED")
		case errors.Is(err, lockseat.ErrTempUserIDRequired):
			return nil, status.Errorf(codes.InvalidArgument, "TEMP_USER_REQUIRED")
		case errors.Is(err, lockseat.ErrSomeSeatsNotFound):
			return nil, status.Errorf(codes.NotFound, "SOME_SEATS_NOT_FOUND")
		case errors.Is(err, lockseat.ErrSeatAlreadyLocked):
			return nil, status.Errorf(codes.AlreadyExists, "SEAT_ALREADY_LOCKED")
		case errors.Is(err, lockseat.ErrSeatAlreadyBooked):
			return nil, status.Errorf(codes.AlreadyExists, "SEAT_ALREADY_BOOKED")
		default:
			return nil, status.Errorf(codes.Internal, "failed to lock seat: %v", err)
		}
	}
	return &bookingpb.LockSeatResponse{
		Success: true,
	}, nil
}
