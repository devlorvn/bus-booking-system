package handler

import (
	"booking-system/internal/bus/delivery/http/response"
	"booking-system/internal/bus/dto"
	"booking-system/internal/bus/usecase"
	"booking-system/pkg/shared/errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BusHandler struct {
	uc *usecase.BusUsecase
}

func NewBusHandler(uc *usecase.BusUsecase) *BusHandler {
	return &BusHandler{uc: uc}
}

func (h *BusHandler) Create(c *gin.Context) {
	// Implementation for creating a bus
	var req dto.CreateBusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.BadRequest("invalid request body"))
		return
	}

	bus, err := h.uc.Create(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "bus created successfully", bus)
}
func (h *BusHandler) List(c *gin.Context) {
	// Implementation for listing all buses
	buses, err := h.uc.List(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "buses retrieved successfully", buses)
}

func (h *BusHandler) GetByID(c *gin.Context) {
	// Implementation for getting a bus by ID
	id := c.Param("id")
	busID, err := uuid.Parse(id)
	if err != nil {
		c.Error(errors.BadRequest("invalid bus ID"))
		return
	}

	bus, err := h.uc.GetByID(c.Request.Context(), busID)
	if err != nil {
		c.Error(err)
		return
	}

	if bus == nil {
		c.Error(errors.NotFound("bus not found"))
		return
	}

	response.Success(c, http.StatusOK, "bus retrieved successfully", bus)
}

func (h *BusHandler) Update(c *gin.Context) {
	// Implementation for updating a bus
	// id := c.Param("id")
	// var req dto.UpdateBusRequest
	// if err := c.ShouldBindJSON(&req); err != nil {
	// 	c.JSON(400, gin.H{"error": err.Error()})
	// 	return
	// }

	// bus, err := h.uc.Update(c.Request.Context(), id, req)
	// if err != nil {
	// 	c.JSON(500, gin.H{"error": err.Error()})
	// 	return
	// }

	// c.JSON(200, gin.H{"data": bus})
}

func (h *BusHandler) Delete(c *gin.Context) {
	// Implementation for deleting a bus
	id := c.Param("id")
	busID, err := uuid.Parse(id)
	if err != nil {
		c.Error(errors.BadRequest("invalid bus ID"))
		return
	}
	err = h.uc.Delete(c.Request.Context(), busID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusNoContent, "bus deleted successfully", nil)
}
