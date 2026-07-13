package main

import (
	"booking-system/configs"
	"booking-system/pkg/kafka"
	"booking-system/pkg/redis"
	"booking-system/pkg/shared/constants"
	"context"
	"log"
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

	// inital kafka reader for notification topic
	kafkaReader := kafka.NewReader(
		cfg.Kafka.Brokers,
		constants.NotificationTopic,
		constants.NotificationServiceWSPollGroup,
	)
	defer kafkaReader.Close()

	go func() {
		log.Println("Notification service: Started")
		for {
			msg, err := kafkaReader.ReadMessage(workerCtx)
			if err != nil {
				if workerCtx.Err() != nil {
					return
				}
				log.Printf("Notification Service: Error reading Kafka message: %v", err)
				continue
			}
			log.Printf("Notification Service: Relaying event to Redis Pub/Sub: %s", string(msg.Value))

			err = redisClient.Publish(workerCtx, constants.WsChanel, msg.Value).Err()
			if err != nil {
				log.Printf("Notification Service: Failed to publish to Redis: %v", err)
				continue
			}
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")
}
