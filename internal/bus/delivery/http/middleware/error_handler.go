package middleware

import (
	httpErrors "booking-system/internal/bus/delivery/http/errors"
	"booking-system/internal/bus/delivery/http/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		if appErr, ok := err.(*httpErrors.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Code)
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, "bus not found", nil)
			return
		}

		response.Error(c, http.StatusInternalServerError, "internal server error", nil)

	}
}
