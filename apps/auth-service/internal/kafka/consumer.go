package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"time"

	"auth-service/internal/db"
	"auth-service/internal/detection"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	engine *detection.Engine
	store  *db.Store
}

func NewConsumer(brokersStr string, topic string, groupID string, engine *detection.Engine, store *db.Store) *Consumer {
	if topic == "" {
		topic = DefaultTopic
	}
	if groupID == "" {
		groupID = "aegis-detection-group"
	}

	brokers := strings.Split(brokersStr, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		MaxWait:        500 * time.Millisecond,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 100 * time.Millisecond,
	})

	slog.Info("Kafka consumer initialized", "brokers", brokers, "topic", topic, "groupID", groupID)

	return &Consumer{
		reader: r,
		engine: engine,
		store:  store,
	}
}

// Start runs the continuous event consumption loop until context is cancelled
func (c *Consumer) Start(ctx context.Context) {
	slog.Info("Kafka detection consumer service started")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Kafka consumer shutting down context cancelled")
			return
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				return
			}
			slog.Error("Kafka error fetching message", "error", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		c.processMessage(ctx, msg)

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			slog.Warn("Kafka error committing message offset", "offset", msg.Offset, "error", err)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) {
	// Safely deserialize JSON payload
	var event db.SecurityEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.Error("Failed to deserialize security event from Kafka",
			"key", string(msg.Key),
			"offset", msg.Offset,
			"error", err,
			"raw_payload", string(msg.Value),
		)
		return
	}

	slog.Debug("Consumed security event from Kafka", "event_id", event.ID, "tenant_id", event.TenantID, "event_type", event.EventType)

	if c.engine == nil {
		slog.Warn("Detection engine not configured on consumer, skipping evaluation")
		return
	}

	// Evaluate event through Detection Engine
	alertParamsList, err := c.engine.Evaluate(ctx, event)
	if err != nil {
		slog.Error("Error processing event through detection engine", "event_id", event.ID, "error", err)
		return
	}

	// Persist generated alerts to PostgreSQL
	for _, params := range alertParamsList {
		if c.store != nil {
			alert, err := c.store.CreateAlert(ctx, params)
			if err != nil {
				slog.Error("Failed to save generated alert to database",
					"tenant_id", params.TenantID,
					"detection_type", params.DetectionType,
					"error", err,
				)
				continue
			}

			slog.Info("Alert successfully persisted to database",
				"alert_id", alert.ID,
				"tenant_id", alert.TenantID,
				"severity", alert.Severity,
				"detection_type", alert.DetectionType,
			)

			// If this is a honeytoken alert, update honeytoken status in DB if ID present
			if params.DetectionType == "honeytoken" && len(params.Metadata) > 0 {
				var meta map[string]any
				if err := json.Unmarshal(params.Metadata, &meta); err == nil {
					if tokenIDRaw, ok := meta["honeytoken_id"]; ok {
						if tokenIDStr, ok := tokenIDRaw.(string); ok {
							if tokenID, err := uuid.Parse(tokenIDStr); err == nil {
								_, _ = c.store.UpdateHoneytokenStatusByTokenID(ctx, db.UpdateHoneytokenStatusByTokenIDParams{
									ID:              tokenID,
									Status:          "TRIGGERED",
									LastTriggeredAt: sql.NullTime{Time: time.Now(), Valid: true},
								})
							}
						}
					}
				}
			}
		}
	}
}

// Close closes the underlying Kafka reader
func (c *Consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
