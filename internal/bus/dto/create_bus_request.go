package dto

import "time"

type CreateBusRequest struct {
	LicensePlate  string    `json:"license_plate" binding:"required"`
	FromLocation  string    `json:"from_location" binding:"required"`
	ToLocation    string    `json:"to_location" binding:"required"`
	DepartureTime time.Time `json:"departure_time" binding:"required"`

	Price       float64  `json:"price" binding:"required"`
	TotalSeats  int      `json:"total_seats" binding:"required"`
	RowName     []string `json:"row_name" binding:"required"`      // e.g. ["A", "B", "C", "D"]
	SeatsPerRow int      `json:"seats_per_row" binding:"required"` // e.g. 6
}
