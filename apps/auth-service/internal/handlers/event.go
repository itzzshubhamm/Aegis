package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"auth-service/internal/db"
	"auth-service/internal/kafka"
	"auth-service/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventHandler struct {
	store    *db.Store
	producer *kafka.Producer
}

func NewEventHandler(store *db.Store, producer *kafka.Producer) *EventHandler {
	return &EventHandler{
		store:    store,
		producer: producer,
	}
}

type IngestEventRequest struct {
	EventID   string         `json:"event_id"`
	EventType string         `json:"event_type" binding:"required"`
	Source    string         `json:"source" binding:"required"`
	SourceIP  string         `json:"source_ip" binding:"required"`
	Resource  string         `json:"resource"`
	Asset     string         `json:"asset"` // alias for resource
	Action    string         `json:"action"`
	Severity  string         `json:"severity"`
	Metadata  map[string]any `json:"metadata"`
	Timestamp time.Time      `json:"timestamp"`
}

// IngestEvent ingests normalized security events into DB and Kafka
func (h *EventHandler) IngestEvent(c *gin.Context) {
	var req IngestEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid security event payload: " + err.Error()})
		return
	}

	// 1. Authenticate & Resolve Tenant strictly from JWT Context
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant identity not resolved from context"})
		return
	}

	// 2. Validate & Normalize Event Fields
	eventID := uuid.New()
	if req.EventID != "" {
		if parsed, err := uuid.Parse(req.EventID); err == nil {
			eventID = parsed
		}
	}

	resource := req.Resource
	if resource == "" {
		resource = req.Asset
	}

	severity := req.Severity
	if severity == "" {
		severity = "INFO"
	}

	eventTimestamp := req.Timestamp
	if eventTimestamp.IsZero() {
		eventTimestamp = time.Now()
	}

	metadataBytes, err := json.Marshal(req.Metadata)
	if err != nil {
		metadataBytes = []byte("{}")
	}

	// 3. Persist Event to PostgreSQL
	event, err := h.store.CreateSecurityEvent(c.Request.Context(), db.CreateSecurityEventParams{
		ID:        eventID,
		TenantID:  tenantID,
		EventType: req.EventType,
		Source:    req.Source,
		SourceIp:  req.SourceIP,
		UserID:    "",
		Resource:  resource,
		Action:    req.Action,
		Severity:  severity,
		Status:    "INFO",
		Metadata:  metadataBytes,
		Timestamp: eventTimestamp,
	})

	if err != nil {
		slog.Error("Failed to persist security event to database", "error", err, "tenant_id", tenantID, "event_id", eventID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to persist security event"})
		return
	}

	// 4. Publish Event to Kafka (aegis.events topic)
	kafkaPayload, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to serialize security event for Kafka", "error", err, "event_id", eventID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize event payload"})
		return
	}

	correlationID, _ := c.Get("correlationId")

	// Must NOT swallow Kafka failure - return appropriate HTTP error on failure
	err = kafka.PublishEvent(c.Request.Context(), h.producer, tenantID.String(), kafkaPayload)
	if err != nil {
		slog.Error("Kafka event publish failed",
			"correlationId", correlationID,
			"tenant_id", tenantID,
			"event_id", eventID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Failed to publish event to message stream",
			"event_id": eventID,
		})
		return
	}

	c.JSON(http.StatusCreated, event)
}
