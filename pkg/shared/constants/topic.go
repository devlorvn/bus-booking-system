package constants

type EventType string

const (
	BookingTopic                 = "booking-events"
	PaymentTopic                 = "payment-events"
	PaymentDLQTopic              = "payment-events-dlq"
	NotificationTopic            = "booking-websocket-events"
	PaymentServicePollGroup      = "payment-service-group"
	BookingServicePollGroup      = "booking-service-group"
	BusServicePollGroup          = "bus-service-group"
	APIGatewayWebSocketPollGroup = "api-gateway-ws-group"
	NotificationServicePollGroup = "notification-service-ws-group"
)

const (
	EventTypeSeatLocked             EventType = "seat_locked"
	EventTypeSeatUnlocked           EventType = "seat_unlocked"
	EventBookingConfirmed                     = "booking_confirmed"
	EventTypeBookingFailed                    = "booking_failed"
	EventTypeBookingCreated                   = "booking_created"
	EventTypeBookingCancelled                 = "booking_cancelled"
	EventTypeSeatsReserved                    = "seats_reserved"
	EventTypeSeatsReservationFailed           = "seats_reservation_failed"
	EventTypeBookingPendingPayment            = "booking_pending_payment"
)
