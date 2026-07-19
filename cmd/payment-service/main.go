package main

import (
	"booking-system/configs"
	"booking-system/pkg/kafka"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	gkafka "github.com/segmentio/kafka-go"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	config := configs.LoadConfig()

	kafka.CreateTopicIfNotExist(config.Kafka.Brokers, constants.BookingTopic, 1, 1)
	kafka.CreateTopicIfNotExist(config.Kafka.Brokers, constants.PaymentTopic, 1, 1)

	bookingKafkaReader := kafka.NewReader(config.Kafka.Brokers, constants.BookingTopic, constants.PaymentServicePollGroup)
	defer bookingKafkaReader.Close()

	writer := kafka.NewWriter(config.Kafka.Brokers, constants.PaymentTopic)
	defer writer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		slog.Info("Payment service is started and listening on %s", slog.String("topic", constants.BookingTopic))
		for {
			msg, err := bookingKafkaReader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					slog.Error("Context cancelled")
					return // Close loop if context closed
				}
				slog.Error("Payment service: Error reading message: %v", slog.String("error", err.Error()))
				continue
			}

			// Check event type header
			eventType := ""
			for _, h := range msg.Headers {
				if h.Key == "event_type" {
					eventType = string(h.Value)
					break
				}
			}

			if eventType != constants.EventTypeBookingPendingPayment {
				continue
			}

			var event events.BookingPendingPaymentEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				slog.Error("Payment service: Error unmarshalling message: %v", slog.String("error", err.Error()))
				continue
			}

			time.Sleep(time.Duration(1000+rand.IntN(1000)) * time.Millisecond)
			// Giả lập 90% thành công, 10% thất bại
			status := "FAILED"
			reason := ""
			// if rand.Float32() < 0.1 {
			// 	status = "FAILED"
			// 	reason = "insufficient_balance"
			// }
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
				slog.Error("Payment Service: Failed to publish payment result: %v", slog.String("error", err.Error()))
			} else {
				slog.Info("Payment Service: Processed booking %s: Status %s", slog.String("booking_id", event.BookingID.String()), slog.String("status", status))
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down payment service")
}
