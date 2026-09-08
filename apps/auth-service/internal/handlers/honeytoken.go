package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"auth-service/internal/db"
	"auth-service/internal/kafka"
	"auth-service/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HoneytokenHandler struct {
	store    *db.Store
	producer *kafka.Producer
}

func NewHoneytokenHandler(store *db.Store, producer *kafka.Producer) *HoneytokenHandler {
	return &HoneytokenHandler{
		store:    store,
		producer: producer,
	}
}

type CreateHoneytokenRequest struct {
	Name       string `json:"name" binding:"required"`
	Type       string `json:"type" binding:"required"` // API_KEY, AWS_CRED, DB_PASS, HONEY_URL
	TokenValue string `json:"token_value"`
}

func (h *HoneytokenHandler) CreateHoneytoken(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant identity not resolved"})
		return
	}

	var req CreateHoneytokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid honeytoken payload: " + err.Error()})
		return
	}

	tokenVal := req.TokenValue
	if tokenVal == "" {
		tokenVal = "ht_" + req.Type + "_" + uuid.New().String()
	}

	honeytoken, err := h.store.CreateHoneytoken(c.Request.Context(), db.CreateHoneytokenParams{
		ID:         uuid.New(),
		TenantID:   tenantID,
		Name:       req.Name,
		Type:       req.Type,
		TokenValue: tokenVal,
		Status:     "ACTIVE",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create honeytoken: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, honeytoken)
}

func (h *HoneytokenHandler) ListHoneytokens(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant identity not resolved"})
		return
	}

	if h == nil || h.store == nil {
		c.JSON(http.StatusOK, []db.Honeytoken{})
		return
	}

	tokens, err := h.store.ListHoneytokens(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve honeytokens"})
		return
	}

	if tokens == nil {
		tokens = []db.Honeytoken{}
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *HoneytokenHandler) GetHoneytokenByID(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant identity not resolved"})
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		idStr = c.Param("token_id")
	}

	tokenID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid honeytoken ID format"})
		return
	}

	token, err := h.store.GetHoneytokenByID(c.Request.Context(), db.GetHoneytokenByIDParams{
		ID:       tokenID,
		TenantID: tenantID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Honeytoken not found"})
		return
	}

	c.JSON(http.StatusOK, token)
}

func (h *HoneytokenHandler) TriggerHoneytoken(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant identity not resolved"})
		return
	}

	idStr := c.Param("token_id")
	if idStr == "" {
		idStr = c.Param("id")
	}

	tokenID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid honeytoken ID format"})
		return
	}

	// Verify honeytoken exists and belongs to authenticated tenant
	token, err := h.store.GetHoneytokenByID(c.Request.Context(), db.GetHoneytokenByIDParams{
		ID:       tokenID,
		TenantID: tenantID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Honeytoken not found or access denied"})
		return
	}

	// Build normalized security event for Honeytoken Trigger
	eventID := uuid.New()
	eventTime := time.Now()
	clientIP := c.ClientIP()
	if clientIP == "" || clientIP == "::1" {
		clientIP = "127.0.0.1"
	}

	metaMap := map[string]any{
		"honeytoken_id": token.ID.String(),
		"name":          token.Name,
		"type":          token.Type,
		"token_value":   token.TokenValue,
	}
	metaBytes, _ := json.Marshal(metaMap)

	event := db.SecurityEvent{
		ID:        eventID,
		TenantID:  tenantID,
		EventType: "HONEYTOKEN_TRIGGERED",
		Source:    "deception_engine",
		SourceIp:  clientIP,
		UserID:    "",
		Resource:  token.Name,
		Action:    "TRIGGERED",
		Severity:  "CRITICAL",
		Status:    "INFO",
		Metadata:  metaBytes,
		Timestamp: eventTime,
		CreatedAt: eventTime,
	}

	// Save security event to DB
	_, err = h.store.CreateSecurityEvent(c.Request.Context(), db.CreateSecurityEventParams{
		ID:        event.ID,
		TenantID:  event.TenantID,
		EventType: event.EventType,
		Source:    event.Source,
		SourceIp:  event.SourceIp,
		UserID:    event.UserID,
		Resource:  event.Resource,
		Action:    event.Action,
		Severity:  event.Severity,
		Status:    event.Status,
		Metadata:  event.Metadata,
		Timestamp: event.Timestamp,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to persist honeytoken event: " + err.Error()})
		return
	}

	// Publish to Kafka aegis.events topic (Detection engine will process & create alert)
	kafkaPayload, err := json.Marshal(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize event payload"})
		return
	}

	err = kafka.PublishEvent(c.Request.Context(), h.producer, tenantID.String(), kafkaPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish honeytoken event to stream: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Deception honeytoken triggered and security event dispatched",
		"honeytoken_id": token.ID,
		"event_id":      eventID,
		"status":        "DISPATCHED",
	})
}
