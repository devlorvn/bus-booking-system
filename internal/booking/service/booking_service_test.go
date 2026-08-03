package service

import (
	"testing"
	"time"

	busDomain "booking-system/internal/bus/domain"

	"github.com/stretchr/testify/assert"
)

func TestCalculateDiscount(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	futureDate := now.Add(24 * time.Hour)
	pastDate := now.Add(-24 * time.Hour)

	tests := []struct {
		name          string
		baseAmount    float64
		discount      *Discount
		now           time.Time
		expectedValue float64
		expectedErr   error
	}{
		{
			name:          "Nil discount code",
			baseAmount:    100.0,
			discount:      nil,
			now:           now,
			expectedValue: 0.0,
			expectedErr:   nil,
		},
		{
			name:          "Empty discount code string",
			baseAmount:    100.0,
			discount:      &Discount{Code: "", Type: DiscountTypePercentage, Value: 10},
			now:           now,
			expectedValue: 0.0,
			expectedErr:   nil,
		},
		{
			name:          "Valid percentage discount",
			baseAmount:    200.0,
			discount:      &Discount{Code: "PROMO20", Type: DiscountTypePercentage, Value: 20, ExpiresAt: futureDate},
			now:           now,
			expectedValue: 40.0,
			expectedErr:   nil,
		},
		{
			name:          "Valid 100% percentage discount",
			baseAmount:    150.0,
			discount:      &Discount{Code: "FREE100", Type: DiscountTypePercentage, Value: 100, ExpiresAt: futureDate},
			now:           now,
			expectedValue: 150.0,
			expectedErr:   nil,
		},
		{
			name:          "Valid fixed amount discount",
			baseAmount:    200.0,
			discount:      &Discount{Code: "FIXED50", Type: DiscountTypeFixed, Value: 50, ExpiresAt: futureDate},
			now:           now,
			expectedValue: 50.0,
			expectedErr:   nil,
		},
		{
			name:          "Fixed discount exceeding base amount",
			baseAmount:    30.0,
			discount:      &Discount{Code: "FIXED50", Type: DiscountTypeFixed, Value: 50, ExpiresAt: futureDate},
			now:           now,
			expectedValue: 30.0,
			expectedErr:   nil,
		},
		{
			name:        "Expired discount code",
			baseAmount:  100.0,
			discount:    &Discount{Code: "EXPIRED", Type: DiscountTypePercentage, Value: 10, ExpiresAt: pastDate},
			now:         now,
			expectedErr: ErrDiscountExpired,
		},
		{
			name:        "Negative discount value",
			baseAmount:  100.0,
			discount:    &Discount{Code: "NEG", Type: DiscountTypeFixed, Value: -10, ExpiresAt: futureDate},
			now:         now,
			expectedErr: ErrInvalidDiscount,
		},
		{
			name:        "Percentage discount over 100%",
			baseAmount:  100.0,
			discount:    &Discount{Code: "OVER100", Type: DiscountTypePercentage, Value: 150, ExpiresAt: futureDate},
			now:         now,
			expectedErr: ErrInvalidDiscount,
		},
		{
			name:        "Unknown discount type",
			baseAmount:  100.0,
			discount:    &Discount{Code: "UNKNOWN", Type: DiscountType("INVALID"), Value: 10, ExpiresAt: futureDate},
			now:         now,
			expectedErr: ErrInvalidDiscount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := CalculateDiscount(tt.baseAmount, tt.discount, tt.now)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Equal(t, 0.0, val)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, val)
			}
		})
	}
}

func TestCalculateTotalPrice(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	futureDate := now.Add(24 * time.Hour)
	pastDate := now.Add(-24 * time.Hour)

	tests := []struct {
		name           string
		basePrice      float64
		seatCount      int
		passengerTypes []PassengerType
		discount       *Discount
		now            time.Time
		expectedTotal  float64
		expectedErr    error
	}{
		{
			name:           "Basic ticket price without discount code (1 seat)",
			basePrice:      100.0,
			seatCount:      1,
			passengerTypes: nil,
			discount:       nil,
			now:            now,
			expectedTotal:  100.0,
			expectedErr:    nil,
		},
		{
			name:           "Basic ticket price without discount code (multiple seats)",
			basePrice:      150.0,
			seatCount:      3,
			passengerTypes: nil,
			discount:       nil,
			now:            now,
			expectedTotal:  450.0,
			expectedErr:    nil,
		},
		{
			name:           "Percentage discount applied",
			basePrice:      100.0,
			seatCount:      2,
			passengerTypes: nil,
			discount:       &Discount{Code: "SAVE10", Type: DiscountTypePercentage, Value: 10, ExpiresAt: futureDate},
			now:            now,
			expectedTotal:  180.0, // 200 - 10% = 180
			expectedErr:    nil,
		},
		{
			name:           "Fixed amount discount applied",
			basePrice:      200.0,
			seatCount:      1,
			passengerTypes: nil,
			discount:       &Discount{Code: "FLAT30", Type: DiscountTypeFixed, Value: 30, ExpiresAt: futureDate},
			now:            now,
			expectedTotal:  170.0, // 200 - 30 = 170
			expectedErr:    nil,
		},
		{
			name:           "Edge case: Child ticket discount (50% base price)",
			basePrice:      100.0,
			seatCount:      1,
			passengerTypes: []PassengerType{PassengerTypeChild},
			discount:       nil,
			now:            now,
			expectedTotal:  50.0,
			expectedErr:    nil,
		},
		{
			name:           "Edge case: Senior ticket discount (70% base price)",
			basePrice:      100.0,
			seatCount:      1,
			passengerTypes: []PassengerType{PassengerTypeSenior},
			discount:       nil,
			now:            now,
			expectedTotal:  70.0,
			expectedErr:    nil,
		},
		{
			name:           "Edge case: Mixed passengers (Adult + Child + Senior) with percentage discount",
			basePrice:      100.0,
			seatCount:      3,
			passengerTypes: []PassengerType{PassengerTypeAdult, PassengerTypeChild, PassengerTypeSenior}, // 100 + 50 + 70 = 220
			discount:       &Discount{Code: "PROMO10", Type: DiscountTypePercentage, Value: 10, ExpiresAt: futureDate}, // 220 - 10% = 198
			now:            now,
			expectedTotal:  198.0,
			expectedErr:    nil,
		},
		{
			name:           "Edge case: Expired discount code",
			basePrice:      100.0,
			seatCount:      1,
			passengerTypes: nil,
			discount:       &Discount{Code: "EXPIRED", Type: DiscountTypePercentage, Value: 20, ExpiresAt: pastDate},
			now:            now,
			expectedErr:    ErrDiscountExpired,
		},
		{
			name:           "Edge case: Seat quantity equal to 0",
			basePrice:      100.0,
			seatCount:      0,
			passengerTypes: nil,
			discount:       nil,
			now:            now,
			expectedErr:    ErrInvalidSeatQuantity,
		},
		{
			name:           "Edge case: Negative seat quantity",
			basePrice:      100.0,
			seatCount:      -2,
			passengerTypes: nil,
			discount:       nil,
			now:            now,
			expectedErr:    ErrInvalidSeatQuantity,
		},
		{
			name:           "Edge case: Negative base price",
			basePrice:      -50.0,
			seatCount:      1,
			passengerTypes: nil,
			discount:       nil,
			now:            now,
			expectedErr:    ErrInvalidBasePrice,
		},
		{
			name:           "Edge case: Passenger count mismatch",
			basePrice:      100.0,
			seatCount:      2,
			passengerTypes: []PassengerType{PassengerTypeAdult},
			discount:       nil,
			now:            now,
			expectedErr:    ErrPassengerCountMismatch,
		},
		{
			name:           "Edge case: Invalid discount configuration",
			basePrice:      100.0,
			seatCount:      1,
			passengerTypes: nil,
			discount:       &Discount{Code: "BAD_DISC", Type: DiscountTypePercentage, Value: -10, ExpiresAt: futureDate},
			now:            now,
			expectedErr:    ErrInvalidDiscount,
		},
		{
			name:           "Edge case: Unknown passenger type defaults to full price",
			basePrice:      100.0,
			seatCount:      1,
			passengerTypes: []PassengerType{PassengerType("OTHER")},
			discount:       nil,
			now:            now,
			expectedTotal:  100.0,
			expectedErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, err := CalculateTotalPrice(tt.basePrice, tt.seatCount, tt.passengerTypes, tt.discount, tt.now)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Equal(t, 0.0, total)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expectedTotal, total, 0.001)
			}
		})
	}
}

func TestBookingService_Methods(t *testing.T) {
	svc := NewBookingService()
	assert.NotNil(t, svc)

	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	futureDate := now.Add(24 * time.Hour)

	// Test CalculateDiscount via interface
	discVal, err := svc.CalculateDiscount(100, &Discount{Code: "TEST", Type: DiscountTypePercentage, Value: 10, ExpiresAt: futureDate}, now)
	assert.NoError(t, err)
	assert.Equal(t, 10.0, discVal)

	// Test CalculateTotalPrice via interface
	totalVal, err := svc.CalculateTotalPrice(100, 2, nil, &Discount{Code: "TEST", Type: DiscountTypePercentage, Value: 10, ExpiresAt: futureDate}, now)
	assert.NoError(t, err)
	assert.Equal(t, 180.0, totalVal)
}

func TestPricingService_Calculate(t *testing.T) {
	svc := NewPricingService()
	assert.NotNil(t, svc)

	bus := &busDomain.Bus{
		Price: 50.0,
	}
	seats := []*busDomain.Seat{
		{},
		{},
	}

	price, err := svc.Calculate(bus, seats)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, price)
}


