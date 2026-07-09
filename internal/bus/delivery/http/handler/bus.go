package handler

import (
	"booking-system/internal/bus/dto"
	"booking-system/pkg/shared/errors"
	"booking-system/pkg/shared/response"
	"net/http"
	"time"

	buspb "booking-system/proto/bus/v1"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BusHandler struct {
	busClient buspb.BusServiceClient
}

func NewBusHandler(busClient buspb.BusServiceClient) *BusHandler {
	return &BusHandler{busClient: busClient}
}

func (h *BusHandler) Create(c *gin.Context) {
	// Implementation for creating a bus
	var req dto.CreateBusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.BadRequest("invalid request body"))
		return
	}

	resp, err := h.busClient.CreateBus(c.Request.Context(), &buspb.CreateBusRequest{
		LicensePlate:  req.LicensePlate,
		FromLocation:  req.FromLocation,
		ToLocation:    req.ToLocation,
		DepartureTime: req.DepartureTime.Format(time.RFC3339),
		Price:         req.Price,
		RowName:       req.RowName,
		SeatsPerRow:   int32(req.SeatsPerRow),
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "bus created successfully", resp.Bus)
}
func (h *BusHandler) List(c *gin.Context) {
	// Implementation for listing all buses
	resp, err := h.busClient.ListBuses(c.Request.Context(), &buspb.ListBusesRequest{})
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "buses retrieved successfully", resp.Buses)
}

func (h *BusHandler) GetSeats(c *gin.Context) {
	id := c.Param("id")
	busID, err := uuid.Parse(id)
	if err != nil {
		c.Error(errors.BadRequest("invalid bus ID"))
		return
	}

	resp, err := h.busClient.GetSeats(c.Request.Context(), &buspb.GetSeatsRequest{BusId: busID.String()})
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "seats retrieved successfully", resp.Seats)
}

func (h *BusHandler) GetByID(c *gin.Context) {
	// Implementation for getting a bus by ID
	id := c.Param("id")
	busID, err := uuid.Parse(id)
	if err != nil {
		c.Error(errors.BadRequest("invalid bus ID"))
		return
	}

	resp, err := h.busClient.GetBus(c.Request.Context(), &buspb.GetBusRequest{BusId: busID.String()})
	if err != nil {
		c.Error(err)
		return
	}

	if resp.Bus == nil {
		c.Error(errors.NotFound("bus not found"))
		return
	}

	response.Success(c, http.StatusOK, "bus retrieved successfully", resp.Bus)
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
	_, err = h.busClient.DeleteBus(c.Request.Context(), &buspb.DeleteBusRequest{BusId: busID.String()})
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusNoContent, "bus deleted successfully", nil)
}
