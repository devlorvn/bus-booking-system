package service

import (
	"errors"
	"time"
)

var (
	ErrInvalidSeatQuantity    = errors.New("seat quantity must be greater than zero")
	ErrInvalidBasePrice       = errors.New("base price must not be negative")
	ErrDiscountExpired        = errors.New("discount code has expired")
	ErrInvalidDiscount        = errors.New("invalid discount configuration")
	ErrPassengerCountMismatch = errors.New("passenger count must match seat quantity")
)

type DiscountType string

const (
	DiscountTypePercentage DiscountType = "PERCENTAGE"
	DiscountTypeFixed      DiscountType = "FIXED"
)

type Discount struct {
	Code      string
	Type      DiscountType
	Value     float64
	ExpiresAt time.Time
}

type PassengerType string

const (
	PassengerTypeAdult  PassengerType = "ADULT"
	PassengerTypeChild  PassengerType = "CHILD"
	PassengerTypeSenior PassengerType = "SENIOR"
)

type BookingService interface {
	CalculateDiscount(baseAmount float64, discount *Discount, now time.Time) (float64, error)
	CalculateTotalPrice(basePrice float64, seatCount int, passengerTypes []PassengerType, discount *Discount, now time.Time) (float64, error)
}

type bookingService struct{}

func NewBookingService() BookingService {
	return &bookingService{}
}

func (s *bookingService) CalculateDiscount(baseAmount float64, discount *Discount, now time.Time) (float64, error) {
	return CalculateDiscount(baseAmount, discount, now)
}

func (s *bookingService) CalculateTotalPrice(basePrice float64, seatCount int, passengerTypes []PassengerType, discount *Discount, now time.Time) (float64, error) {
	return CalculateTotalPrice(basePrice, seatCount, passengerTypes, discount, now)
}

// CalculateDiscount calculates the discount amount given a base total and discount code details.
func CalculateDiscount(baseAmount float64, discount *Discount, now time.Time) (float64, error) {
	if discount == nil || discount.Code == "" {
		return 0, nil
	}

	if !discount.ExpiresAt.IsZero() && now.After(discount.ExpiresAt) {
		return 0, ErrDiscountExpired
	}

	if discount.Value < 0 {
		return 0, ErrInvalidDiscount
	}

	var discountAmount float64
	switch discount.Type {
	case DiscountTypePercentage:
		if discount.Value > 100 {
			return 0, ErrInvalidDiscount
		}
		discountAmount = baseAmount * (discount.Value / 100.0)
	case DiscountTypeFixed:
		discountAmount = discount.Value
		if discountAmount > baseAmount {
			discountAmount = baseAmount
		}
	default:
		return 0, ErrInvalidDiscount
	}

	return discountAmount, nil
}

// CalculateTotalPrice calculates total booking price based on base price, seat quantity, passenger categories, and discounts.
func CalculateTotalPrice(basePrice float64, seatCount int, passengerTypes []PassengerType, discount *Discount, now time.Time) (float64, error) {
	if seatCount <= 0 {
		return 0, ErrInvalidSeatQuantity
	}

	if basePrice < 0 {
		return 0, ErrInvalidBasePrice
	}

	var subtotal float64
	if len(passengerTypes) > 0 {
		if len(passengerTypes) != seatCount {
			return 0, ErrPassengerCountMismatch
		}
		for _, pType := range passengerTypes {
			switch pType {
			case PassengerTypeChild:
				subtotal += basePrice * 0.5
			case PassengerTypeSenior:
				subtotal += basePrice * 0.7
			case PassengerTypeAdult, "":
				subtotal += basePrice
			default:
				subtotal += basePrice
			}
		}
	} else {
		subtotal = basePrice * float64(seatCount)
	}

	discountAmount, err := CalculateDiscount(subtotal, discount, now)
	if err != nil {
		return 0, err
	}

	totalPrice := subtotal - discountAmount
	if totalPrice < 0 {
		totalPrice = 0
	}

	return totalPrice, nil
}
