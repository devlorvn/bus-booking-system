package main

import (
	"booking-system/configs"
	"booking-system/pkg/kafka"
	"booking-system/pkg/redis"
	"booking-system/pkg/shared/constants"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := configs.LoadConfig()

	redisClient := redis.NewClient(&cfg.Redis)
	defer redisClient.Close()
	// Create worker context
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	// Initial kafka reader for notification topic
	kafkaReader := kafka.NewReader(
		cfg.Kafka.Brokers,
		constants.NotificationTopic,
		constants.NotificationServiceWSPollGroup,
	)
	defer kafkaReader.Close()

	go func() {
		slog.Info("Notification service: Started")
		for {
			msg, err := kafkaReader.ReadMessage(workerCtx)
			if err != nil {
				if workerCtx.Err() != nil {
					return
				}
				slog.Error("Notification Service: Error reading Kafka message: %v", slog.String("error", err.Error()))
				continue
			}
			slog.Info("Notification Service: Relaying event to Redis Pub/Sub: %s", slog.String("msg", string(msg.Value)))

			err = redisClient.Publish(workerCtx, constants.WsChanel, msg.Value).Err()
			if err != nil {
				slog.Error("Notification Service: Failed to publish to Redis: %v", slog.String("error", err.Error()))
				continue
			}
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server gracefully...")
}
