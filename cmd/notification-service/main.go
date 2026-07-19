package main

import (
	"booking-system/configs"
	"booking-system/pkg/kafka"
	"booking-system/pkg/redis"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	cfg := configs.LoadConfig()

	redisClient := redis.NewClient(&cfg.Redis)
	defer redisClient.Close()
	// Create worker context
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	// Initial kafka reader for notification topic (transient seat locks)
	notiKafkaReader := kafka.NewReader(
		cfg.Kafka.Brokers,
		constants.NotificationTopic,
		constants.NotificationServicePollGroup,
	)
	defer notiKafkaReader.Close()

	// Initial kafka reader for booking events topic (business events)
	bookingReader := kafka.NewReader(
		cfg.Kafka.Brokers,
		constants.BookingTopic,
		constants.NotificationServicePollGroup,
	)
	defer bookingReader.Close()

	var wg sync.WaitGroup

	// Goroutine 1: Relay transient WS events
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("Notification service: Relaying transient WS events")
		for {
			msg, err := notiKafkaReader.ReadMessage(workerCtx)
			if err != nil {
				if workerCtx.Err() != nil {
					return
				}
				slog.Error("Notification Service: Error reading transient WS event: %v", slog.String("error", err.Error()))
				continue
			}
			slog.Info("Notification Service: Relaying event to Redis Pub/Sub: %s", slog.String("msg", string(msg.Value)))

			err = redisClient.Publish(workerCtx, constants.WsChanel, msg.Value).Err()
			if err != nil {
				slog.Error("Notification Service: Failed to publish transient WS event to Redis: %v", slog.String("error", err.Error()))
				continue
			}
		}
	}()

	// Goroutine 2: Relay business events (like booking_cancelled) to WS
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("Notification service: Relaying business events to WS")
		for {
			msg, err := bookingReader.ReadMessage(workerCtx)
			if err != nil {
				if workerCtx.Err() != nil {
					return
				}
				slog.Error("Notification Service: Error reading booking event: %v", slog.String("error", err.Error()))
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

			if eventType == constants.EventTypeBookingCancelled {
				var event events.BookingCancelledEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					slog.Error("Notification Service: Error unmarshalling cancelled event: %v", slog.String("error", err.Error()))
					continue
				}

				// Format into WebSocket JSON structure expected by client
				payload := map[string]interface{}{
					"event":  constants.EventTypeBookingFailed,
					"bus_id": event.BusID.String(),
					"data": map[string]interface{}{
						"booking_id": event.BookingID.String(),
						"bus_id":     event.BusID.String(),
						"seat_codes": event.SeatCodes,
					},
				}

				payloadBytes, err := json.Marshal(payload)
				if err != nil {
					slog.Error("Notification Service: Failed to marshal WS payload: %v", slog.String("error", err.Error()))
					continue
				}

				slog.Info("Notification Service: Relaying booking cancelled event to Redis Pub/Sub")
				err = redisClient.Publish(workerCtx, constants.WsChanel, payloadBytes).Err()
				if err != nil {
					slog.Error("Notification Service: Failed to publish relayed booking cancelled to Redis: %v", slog.String("error", err.Error()))
					continue
				}
			} else if eventType == constants.EventTypeBookingConfirmed {
				var event events.BookingConfirmedEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					slog.Error("Notification Service: Error unmarshalling confirmed event: %v", slog.String("error", err.Error()))
					continue
				}

				// Format into WebSocket JSON structure expected by client
				payload := map[string]interface{}{
					"event":  constants.EventTypeBookingConfirmed,
					"bus_id": event.BusID.String(),
					"data": map[string]interface{}{
						"booking_id": event.BookingID.String(),
						"bus_id":     event.BusID.String(),
					},
				}

				payloadBytes, err := json.Marshal(payload)
				if err != nil {
					slog.Error("Notification Service: Failed to marshal WS payload: %v", slog.String("error", err.Error()))
					continue
				}

				slog.Info("Notification Service: Relaying booking confirmed event to Redis Pub/Sub")
				err = redisClient.Publish(workerCtx, constants.WsChanel, payloadBytes).Err()
				if err != nil {
					slog.Error("Notification Service: Failed to publish relayed booking confirmed to Redis: %v", slog.String("error", err.Error()))
					continue
				}
			}
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server gracefully...")

	cancelWorkers()
	wg.Wait()
	slog.Info("Notification service exited")
}
