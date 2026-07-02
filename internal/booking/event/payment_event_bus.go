package event

type PaymentEventBus struct {
	BookingCreated chan BookingCreatedEvent
	PaymentSuccess chan PaymentSuccessEvent
	PaymentFailed  chan PaymentFailedEvent
	DLQ            chan DeadLetterEvent
}

func NewPaymentEventBus() *PaymentEventBus {
	return &PaymentEventBus{
		BookingCreated: make(chan BookingCreatedEvent, 100),
		PaymentSuccess: make(chan PaymentSuccessEvent),
		PaymentFailed:  make(chan PaymentFailedEvent),
		DLQ:            make(chan DeadLetterEvent, 100),
	}
}
