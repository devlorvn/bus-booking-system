package events

import (
	"booking-system/pkg/shared/constants"
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type KafkaBookingPublisher struct {
	writer *kafka.Writer
}

func NewKafkaBookingPublisher(writer *kafka.Writer) *KafkaBookingPublisher {
	return &KafkaBookingPublisher{
		writer: writer,
	}
}

func (p *KafkaBookingPublisher) PublishBookingCreated(ctx context.Context, event BookingCreatedEvent) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.BookingID.String()),
		Value: bytes,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(constants.EventTypeBookingCreated)},
		},
	})
}

func (p *KafkaBookingPublisher) PublishBookingCancelled(ctx context.Context, event BookingCancelledEvent) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.BookingID.String()),
		Value: bytes,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(constants.EventTypeBookingCancelled)},
		},
	})
}

func (p *KafkaBookingPublisher) PublishBookingPendingPayment(ctx context.Context, event BookingPendingPaymentEvent) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.BookingID.String()),
		Value: bytes,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(constants.EventTypeBookingPendingPayment)},
		},
	})
}
