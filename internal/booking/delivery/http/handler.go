package http

import (
	httpErrors "booking-system/pkg/shared/errors"
	bookingpb "booking-system/proto/booking/v1"
	"errors"

	"net/http"

	"booking-system/internal/booking/dto"
	confirmbooking "booking-system/internal/booking/usecase/confirm_booking"
	lockseat "booking-system/internal/booking/usecase/lock_seat"
	"booking-system/pkg/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type BookingHandler struct {
	lockSeatUC    *lockseat.LockSeatUsecase
	bookingClient bookingpb.BookingServiceClient
	// paymentPublisher confirmbooking.PaymentEventPublisher
}

func NewBookingHandler(
	lockSeatUC *lockseat.LockSeatUsecase,
	bookingClient bookingpb.BookingServiceClient,
	// paymentPublisher confirmbooking.PaymentEventPublisher,
) *BookingHandler {
	return &BookingHandler{
		lockSeatUC:    lockSeatUC,
		bookingClient: bookingClient,
		// paymentPublisher: paymentPublisher,
	}
}

func (h *BookingHandler) LockSeat(c *gin.Context) {
	var req dto.LockSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(httpErrors.BadRequest("invalid request body"))
		return
	}

	resp, err := h.lockSeatUC.Execute(c.Request.Context(), req)
	if err != nil {
		var appError *httpErrors.AppError
		switch {
		case errors.Is(err, lockseat.ErrBusNotFound):
			appError = httpErrors.BadRequest(err.Error())
		case errors.Is(err, lockseat.ErrNoSeatSelected), errors.Is(err, lockseat.ErrTempUserIDRequired):
			appError = httpErrors.BadRequest(err.Error())
		case errors.Is(err, lockseat.ErrSomeSeatsNotFound):
			appError = httpErrors.NotFound(err.Error())
		case errors.Is(err, lockseat.ErrSeatAlreadyLocked):
			appError = httpErrors.Conflict(err.Error())
		case errors.Is(err, lockseat.ErrSeatAlreadyBooked):
			appError = httpErrors.Conflict(err.Error())
		default:
			appError = httpErrors.InternalServerError(err.Error())
		}
		c.Error(appError)
		return
	}

	response.Success(c, http.StatusOK, "lock seat sucessfully", resp)
}

func (h *BookingHandler) ConfirmBooking(c *gin.Context) {
	var req dto.ConfirmBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(httpErrors.BadRequest("invalid request body"))
		return
	}

	resp, err := h.bookingClient.ConfirmBooking(c.Request.Context(), &bookingpb.ConfirmBookingRequest{
		BusId:      req.BusID.String(),
		SeatCodes:  req.SeatCodes,
		TempUserId: req.TempUserID,
		Phone:      req.PhoneNumber,
		UserName:   req.Name,
		Email:      req.Email,
	})
	if err != nil {
		st, ok := status.FromError(err)
		var appError *httpErrors.AppError

		if ok {
			switch st.Message() {
			case confirmbooking.ErrTempUserRequired.Error():
				appError = httpErrors.BadRequest(err.Error())
			case confirmbooking.ErrPhoneRequired.Error(), confirmbooking.ErrNoSeatSelected.Error():
				appError = httpErrors.BadRequest(err.Error())
			case confirmbooking.ErrBusNotFound.Error():
				appError = httpErrors.NotFound(err.Error())
			case confirmbooking.ErrSeatNotFound.Error():
				appError = httpErrors.Conflict(err.Error())
			case confirmbooking.ErrSeatAlreadyBooked.Error():
				appError = httpErrors.Conflict(err.Error())
			case confirmbooking.ErrLockExpired.Error():
				appError = httpErrors.Conflict(err.Error())
			default:
				appError = httpErrors.InternalServerError(err.Error())
			}

		} else {
			appError = httpErrors.InternalServerError(err.Error())

		}
		c.Error(appError)
		return
	}

	bookingUUID, err := uuid.Parse(resp.BookingId)
	if err != nil {
		c.Error(httpErrors.InternalServerError("invalid booking id"))
		return
	}

	// _ = h.paymentPublisher.PublishBookingCreated(c.Request.Context(), bookingUUID)

	response.Success(c, http.StatusOK, "confirm booking sucessfully", &dto.ConfirmBookingResponse{
		BookingID:   bookingUUID,
		TotalAmount: resp.TotalAmount,
		Status:      resp.Status,
	})
}
