package events

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type KafkaPaymentPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPaymentPublisher(writer *kafka.Writer) *KafkaPaymentPublisher {
	return &KafkaPaymentPublisher{
		writer: writer,
	}
}

func (p *KafkaPaymentPublisher) PublishBookingCreated(ctx context.Context, bookingID uuid.UUID) error {
	eventMsg := BookingCreatedEvent{
		BookingID: bookingID,
	}

	bytes, err := json.Marshal(eventMsg)

	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(bookingID.String()),
		Value: bytes,
	})
}

// func (p *KafkaPaymentPublisher) PublishPaymentSuccess(ctx context.Context, bookingID uuid.UUID) error {

// }
