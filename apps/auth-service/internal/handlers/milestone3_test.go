package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"auth-service/internal/auth"
	"auth-service/internal/db"
	"auth-service/internal/detection"
	"auth-service/internal/kafka"
	"auth-service/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setupTestRouterWithStore(store *db.Store) (*gin.Engine, string, uuid.UUID, string, uuid.UUID) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	jwtSecret := "test-secret-key-12345"

	tenantIDA := uuid.New()
	tokenA, _, _ := auth.GenerateTokenPair(uuid.New(), tenantIDA, "userA@tenantA.com", "analyst", jwtSecret)

	tenantIDB := uuid.New()
	tokenB, _, _ := auth.GenerateTokenPair(uuid.New(), tenantIDB, "userB@tenantB.com", "analyst", jwtSecret)

	producer := &kafka.Producer{} // dummy producer for HTTP test routing if offline

	stateTracker := detection.NewInMemoryStateTracker()
	engine := detection.NewEngine(
		detection.NewBruteForceRule(stateTracker, 3, 5*time.Minute),
		detection.NewUnauthorizedAccessRule(),
		detection.NewHoneytokenRule(),
	)

	eventHandler := NewEventHandler(store, producer)
	alertHandler := NewAlertHandler(store)
	honeytokenHandler := NewHoneytokenHandler(store, producer)

	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(jwtSecret))
	{
		protected.POST("/events/ingest", eventHandler.IngestEvent)
		protected.GET("/events", eventHandler.ListEvents)
		protected.GET("/events/:id", eventHandler.GetEventByID)

		protected.GET("/alerts", alertHandler.ListAlerts)
		protected.GET("/alerts/:id", alertHandler.GetAlertByID)
		protected.PATCH("/alerts/:id", alertHandler.UpdateAlertStatus)

		protected.POST("/deception/honeytokens", honeytokenHandler.CreateHoneytoken)
		protected.GET("/deception/honeytokens", honeytokenHandler.ListHoneytokens)
		protected.GET("/deception/honeytokens/:id", honeytokenHandler.GetHoneytokenByID)
		protected.POST("/deception/trigger/:token_id", honeytokenHandler.TriggerHoneytoken)
	}

	_ = engine

	return r, tokenA, tenantIDA, tokenB, tenantIDB
}

func TestMilestone3_Handlers(t *testing.T) {
	r, tokenA, tenantIDA, tokenB, tenantIDB := setupTestRouterWithStore(nil)

	t.Run("Alerts API Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/alerts", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized for missing token, got %d", w.Code)
		}
	})

	t.Run("Honeytoken Creation Validation", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]string{})
		req, _ := http.NewRequest("POST", "/api/v1/deception/honeytokens", bytes.NewBuffer(reqBody))
		req.Header.Set("Authorization", "Bearer "+tokenA)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request for empty honeytoken payload, got %d", w.Code)
		}
	})

	_ = tenantIDA
	_ = tokenB
	_ = tenantIDB
}
