package constants

type EventType string

const (
	BookingTopic                   = "booking-events"
	PaymentTopic                   = "payment-events"
	NotificationTopic              = "booking-websocket-events"
	PaymentServicePollGroup        = "payment-service-group"
	BookingServicePollGroup        = "booking-service-group"
	APIGatewayWebSocketPollGroup   = "api-gateway-ws-group"
	NotificationServiceWSPollGroup = "notification-service-ws-group"
)

const (
	EventTypeSeatLocked   EventType = "seat_locked"
	EventTypeSeatUnlocked EventType = "seat_unlocked"
	EventBookingConfirmed           = "booking_confirmed"
)
