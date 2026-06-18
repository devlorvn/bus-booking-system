package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *BookingHandler,
) {
	bookings := router.Group("/bookings")

	bookings.POST("/lock", handler.LockSeat)
}
