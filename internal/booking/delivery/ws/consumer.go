package ws

import (
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
)

var kafkaUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsConsumer struct {
	reader *kafka.Reader
	hub    *Hub
}

func NewWsConsumer(reader *kafka.Reader, hub *Hub) *WsConsumer {
	return &WsConsumer{
		reader: reader,
		hub:    hub,
	}
}

func (c *WsConsumer) Consume(ctx context.Context) error {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Context cancelled")
				return ctx.Err()
			}
			log.Printf("Error reading message: %v", err)
			continue
		}

		log.Printf("[WsConsumer] Received Kafka message: Key=%s, Value=%s", string(msg.Key), string(msg.Value))

		if msg.Key == nil {
			log.Println("Message has no key, skipping")
			continue
		}
		var event events.KafkaWsEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}
		switch constants.EventType(event.Event) {
		case constants.EventTypeSeatLocked:
			var data map[string]interface{}
			if err := json.Unmarshal(event.Data, &data); err != nil {
				log.Printf("Error unmarshalling message: %v", err)
				continue
			}
			c.hub.Broadcast(event.BusID, BroadcastMessage{
				Event: constants.EventTypeSeatLocked,
				Data:  data,
			})
		case constants.EventTypeSeatUnlocked:
			var data map[string]interface{}
			if err := json.Unmarshal(event.Data, &data); err != nil {
				log.Printf("Error unmarshalling message: %v", err)
				continue
			}
			c.hub.Broadcast(event.BusID, BroadcastMessage{
				Event: constants.EventTypeSeatUnlocked,
				Data:  data,
			})
		}

	}
}
