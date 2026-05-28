package http

import (
	"booking-system/internal/delivery/http/handler"

	"github.com/gin-gonic/gin"
)

type HandlerHttpGroup struct {
	BusHandler *handler.BusHandler
}

func NewRouter(h *HandlerHttpGroup) *gin.Engine {
	r := gin.Default()

	bus := r.Group("/buses")
	{
		bus.GET("/", h.BusHandler.List)
		bus.POST("/", h.BusHandler.Create)
		bus.GET("/:id", h.BusHandler.GetByID)
		bus.PUT("/:id", h.BusHandler.Update)
		bus.DELETE("/:id", h.BusHandler.Delete)
	}

	return r
}
