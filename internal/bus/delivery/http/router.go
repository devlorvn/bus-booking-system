package http

import (
	"booking-system/internal/bus/delivery/http/handler"

	"github.com/gin-gonic/gin"
)

func RegiserBusRouter(
	router *gin.RouterGroup,
	handler *handler.BusHandler,
) {

	bus := router.Group("/buses")
	{
		bus.GET("/", handler.List)
		bus.POST("/", handler.Create)
		bus.GET("/:id", handler.GetByID)
		bus.PUT("/:id", handler.Update)
		bus.DELETE("/:id", handler.Delete)
	}
}
