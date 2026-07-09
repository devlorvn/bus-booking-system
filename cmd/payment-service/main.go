package main

import (
	"booking-system/configs"
	"booking-system/pkg/kafka"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	gkafka "github.com/segmentio/kafka-go"
)

func main() {
	config := configs.LoadConfig()

	kafka.CreateTopicIfNotExist(config.Kafka.Brokers, constants.BookingTopic, 1, 1)
	kafka.CreateTopicIfNotExist(config.Kafka.Brokers, constants.PaymentTopic, 1, 1)

	reader := kafka.NewReader(config.Kafka.Brokers, constants.BookingTopic, constants.PaymentServicePoll)
	defer reader.Close()

	writer := kafka.NewWriter(config.Kafka.Brokers, constants.PaymentTopic)
	defer writer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Printf("Payment service is started and listening on %s", constants.BookingTopic)
		for {
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("Context cancelled")
					return // Close loop if context closed
				}
				log.Printf("Payment service: Error reading message: %v", err)
				continue
			}

			var event events.BookingCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Payment service: Error unmarshalling message: %v", err)
				continue
			}

			time.Sleep(time.Duration(1000+rand.IntN(1000)) * time.Millisecond)
			// Giả lập 90% thành công, 10% thất bại
			status := "SUCCESS"
			reason := ""
			if rand.Float32() < 0.1 {
				status = "FAILED"
				reason = "insufficient_balance"
			}
			processedEvent := events.PaymentProcessedEvent{
				BookingID: event.BookingID,
				Status:    status,
				Reason:    reason,
			}
			bytes, _ := json.Marshal(processedEvent)
			err = writer.WriteMessages(ctx, gkafka.Message{
				Key:   []byte(event.BookingID.String()),
				Value: bytes,
			})
			if err != nil {
				log.Printf("Payment Service: Failed to publish payment result: %v", err)
			} else {
				log.Printf("Payment Service: Processed booking %s: Status %s", event.BookingID, status)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down payment service")
}
