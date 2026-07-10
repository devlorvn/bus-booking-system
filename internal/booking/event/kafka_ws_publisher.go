package event

import (
	"booking-system/pkg/shared/constants"
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type KafkaWsPublisher struct {
	writer *kafka.Writer
}

func NewKafkaWsPublisher(writer *kafka.Writer) *KafkaWsPublisher {
	return &KafkaWsPublisher{
		writer: writer,
	}
}

func (p *KafkaWsPublisher) PublishSeatLocked(busId string, seatID string, seatCode string, tempUserId string) error {
	payload := map[string]interface{}{
		"event":  constants.EventTypeSeatLocked,
		"bus_id": busId,
		"data": map[string]string{
			"bus_id":       busId,
			"seat_id":      seatID,
			"seat_code":    seatCode,
			"temp_user_id": tempUserId,
		},
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(context.Background(), kafka.Message{
		Value: bytes,
		Key:   []byte(busId),
	})
}

func (p *KafkaWsPublisher) PublishSeatReleased(busId string, seatCode string) error {
	payload := map[string]interface{}{
		"event":  constants.EventTypeSeatUnlocked,
		"bus_id": busId,
		"data": map[string]string{
			"bus_id":    busId,
			"seat_code": seatCode,
		},
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(context.Background(), kafka.Message{
		Value: bytes,
		Key:   []byte(busId),
	})
}
