package http

import (
	httpErrors "booking-system/pkg/shared/errors"
	"errors"

	"net/http"

	"booking-system/internal/booking/dto"
	lockseat "booking-system/internal/booking/usecase/lock_seat"
	"booking-system/pkg/shared/response"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	lockSeatUC *lockseat.LockSeatUsecase
}

func NewBookingHandler(lockSeatUC *lockseat.LockSeatUsecase) *BookingHandler {
	return &BookingHandler{
		lockSeatUC: lockSeatUC,
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
