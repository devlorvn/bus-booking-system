package kafka

import (
	"log"
	"time"

	gkafka "github.com/segmentio/kafka-go"
)

func NewWriter(brokers []string, topic string) *gkafka.Writer {
	return &gkafka.Writer{
		Addr:         gkafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &gkafka.LeastBytes{}, // distribute msg evenly into partitions by size (bytes)
		WriteTimeout: 10 * time.Second,
		RequiredAcks: gkafka.RequireAll,
	}
}

func NewReader(brokers []string, topic string, groupID string) *gkafka.Reader {
	return gkafka.NewReader(gkafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3,              // wait for 10kb messages
		MaxBytes:       10e6,              // limit the max size of the batch messages
		CommitInterval: time.Second,       // auto commit offset after 1 second
		StartOffset:    gkafka.LastOffset, // read the new message only
	})
}

func CreateTopicIfNotExist(brokers []string, topic string, numPartitions int, replicationFactor int) {
	if len(brokers) == 0 {
		return
	}
	conn, err := gkafka.Dial("tcp", brokers[0])
	if err != nil {
		log.Printf("Warning: Fail to dial Kafka broker: %v", err)
		return
	}
	defer conn.Close()

	err = conn.CreateTopics(gkafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	})
	if err != nil {
		log.Printf("Warning: Fail to create topic: %v", err)
		return
	}

	log.Printf("Topic %s is created successfully", topic)
}
