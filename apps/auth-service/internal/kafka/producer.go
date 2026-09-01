package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

const DefaultTopic = "aegis.events"

type Producer struct {
	writer *kafka.Writer
	topic  string
}

// NewProducer initializes a segmentio/kafka-go writer with environment-configured brokers
func NewProducer(brokersStr string, topic string) *Producer {
	if topic == "" {
		topic = DefaultTopic
	}

	brokers := strings.Split(brokersStr, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.Hash{}, // Consistent partitioning by message Key (tenant_id)
		BatchTimeout:           10 * time.Millisecond,
		AllowAutoTopicCreation: true,
		Async:                  false, // Synchronous publish to ensure error handling
	}

	slog.Info("Kafka producer initialized", "brokers", brokers, "topic", topic)

	return &Producer{
		writer: w,
		topic:  topic,
	}
}

// PublishEvent publishes a message to the Kafka topic. Returns error if publishing fails.
func PublishEvent(ctx context.Context, p *Producer, key string, payload []byte) error {
	if p == nil || p.writer == nil {
		return fmt.Errorf("kafka producer is not initialized")
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
	}

	var lastErr error
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := p.writer.WriteMessages(ctx, msg)
		if err == nil {
			slog.Info("Message published to Kafka successfully", "topic", p.topic, "key", key, "bytes", len(payload))
			return nil
		}

		lastErr = err
		slog.Warn("Kafka publish attempt failed, retrying...", "attempt", attempt, "error", err)
		time.Sleep(time.Duration(attempt*100) * time.Millisecond)
	}

	slog.Error("Failed to publish message to Kafka after retries", "topic", p.topic, "key", key, "error", lastErr)
	return fmt.Errorf("kafka publish error: %w", lastErr)
}

// Close closes the underlying Kafka writer connection
func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
