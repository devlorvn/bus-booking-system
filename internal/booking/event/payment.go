package event

type BookingCreatedEvent struct {
	BookingID string `json:"booking_id"`
}

type PaymentSuccessEvent struct {
	BookingID string `json:"booking_id"`
}

type PaymentFailedEvent struct {
	BookingID string `json:"booking_id"`
	Reason    string `json:"reason"`
}
