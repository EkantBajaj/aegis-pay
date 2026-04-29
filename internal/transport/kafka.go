package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer initializes a writer with 'RequiredAcks: -1' for financial safety
func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll, // Wait for all replicas to acknowledge
			Async:        false,           // Synchronous for immediate error handling in the gateway
		},
	}
}

// PublishFailure sends the failed transaction details to the recovery topic
func (k *KafkaProducer) PublishFailure(ctx context.Context, transactionData interface{}) error {
	payload, err := json.Marshal(transactionData)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	err = k.writer.WriteMessages(ctx, kafka.Message{
		Value: payload,
	})

	if err != nil {
		return fmt.Errorf("kafka write failed: %w", err)
	}

	log.Println("--- KAFKA: Published failed transaction event for recovery ---")
	return nil
}

func (k *KafkaProducer) Close() error {
	return k.writer.Close()
}
