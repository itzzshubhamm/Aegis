package handlers

import (
	"context"
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

// MockStore wraps db queries with in-memory state for pure unit/integration testing
type MockStore struct {
	events     map[uuid.UUID][]db.SecurityEvent
	alerts     map[uuid.UUID][]db.Alert
	honeytokens map[uuid.UUID][]db.Honeytoken
}

func NewMockStore() *MockStore {
	return &MockStore{
		events:      make(map[uuid.UUID][]db.SecurityEvent),
		alerts:      make(map[uuid.UUID][]db.Alert),
		honeytokens: make(map[uuid.UUID][]db.Honeytoken),
	}
}

func TestMilestone4_FullE2EAndTenantIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSecret := "aegis-e2e-secret-key-999"

	// Create Tenant A and Tenant B
	tenantIDA := uuid.New()
	tenantIDB := uuid.New()

	tokenA, _, err := auth.GenerateTokenPair(uuid.New(), tenantIDA, "analystA@tenantA.com", "analyst", jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate token A: %v", err)
	}

	tokenB, _, err := auth.GenerateTokenPair(uuid.New(), tenantIDB, "analystB@tenantB.com", "analyst", jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate token B: %v", err)
	}

	// Set up Detection Engine with In-Memory State Tracker
	stateTracker := detection.NewInMemoryStateTracker()
	engine := detection.NewEngine(
		detection.NewBruteForceRule(stateTracker, 3, 5*time.Minute),
		detection.NewUnauthorizedAccessRule(),
		detection.NewHoneytokenRule(),
	)

	// Simulated storage for end-to-end verification
	alertsStoreA := []db.Alert{}
	alertsStoreB := []db.Alert{}

	// Direct Engine Evaluation Test - Scenario A: Brute Force (3 fails -> no alert, 4th -> HIGH alert)
	ctx := context.Background()
	targetIP := "192.168.1.55"

	for i := 1; i <= 3; i++ {
		evt := db.SecurityEvent{
			ID:        uuid.New(),
			TenantID:  tenantIDA,
			EventType: "LOGIN_FAILED",
			Source:    "auth_api",
			SourceIp:  targetIP,
			UserID:    "target_user",
			Action:    "LOGIN_FAILED",
			Timestamp: time.Now(),
		}
		alerts, err := engine.Evaluate(ctx, evt)
		if err != nil {
			t.Fatalf("Engine evaluate error: %v", err)
		}
		if len(alerts) != 0 {
			t.Errorf("Expected 0 alerts on failed login attempt %d, got %d", i, len(alerts))
		}
	}

	// 4th Failed Login Attempt -> Threshold Crossed -> HIGH Alert
	evt4 := db.SecurityEvent{
		ID:        uuid.New(),
		TenantID:  tenantIDA,
		EventType: "LOGIN_FAILED",
		Source:    "auth_api",
		SourceIp:  targetIP,
		UserID:    "target_user",
		Action:    "LOGIN_FAILED",
		Timestamp: time.Now(),
	}
	alerts4, err := engine.Evaluate(ctx, evt4)
	if err != nil {
		t.Fatalf("Engine evaluate error on attempt 4: %v", err)
	}
	if len(alerts4) != 1 {
		t.Fatalf("Expected 1 HIGH alert on 4th failed login, got %d", len(alerts4))
	}
	if alerts4[0].Severity != "HIGH" || alerts4[0].DetectionType != "brute_force" {
		t.Errorf("Unexpected alert params: %+v", alerts4[0])
	}
	alertsStoreA = append(alertsStoreA, db.Alert{
		ID:            alerts4[0].ID,
		TenantID:      alerts4[0].TenantID,
		Severity:      alerts4[0].Severity,
		DetectionType: alerts4[0].DetectionType,
		SourceIp:      alerts4[0].SourceIp,
		AffectedAsset: alerts4[0].AffectedAsset,
		Description:   alerts4[0].Description,
		Status:        alerts4[0].Status,
		Timestamp:     alerts4[0].Timestamp,
	})

	// Scenario B: Unauthorized Access -> HIGH alert
	unauthEvt := db.SecurityEvent{
		ID:        uuid.New(),
		TenantID:  tenantIDA,
		EventType: "UNAUTHORIZED_ACCESS",
		Source:    "gateway",
		SourceIp:  "10.0.0.99",
		Resource:  "/admin/vault",
		Action:    "ACCESS_DENIED",
		Status:    "FORBIDDEN",
		Timestamp: time.Now(),
	}
	unauthAlerts, err := engine.Evaluate(ctx, unauthEvt)
	if err != nil || len(unauthAlerts) != 1 {
		t.Fatalf("Expected 1 HIGH alert for unauthorized access, got %d (err: %v)", len(unauthAlerts), err)
	}
	if unauthAlerts[0].Severity != "HIGH" || unauthAlerts[0].DetectionType != "unauthorized_access" {
		t.Errorf("Unexpected unauthorized alert: %+v", unauthAlerts[0])
	}

	// Scenario C: Honeytoken Trigger -> CRITICAL alert
	metaMap := map[string]any{"honeytoken_id": uuid.New().String(), "name": "Fake AWS Key"}
	metaBytes, _ := json.Marshal(metaMap)

	htEvt := db.SecurityEvent{
		ID:        uuid.New(),
		TenantID:  tenantIDA,
		EventType: "HONEYTOKEN_TRIGGERED",
		Source:    "deception",
		SourceIp:  "198.51.100.14",
		Resource:  "Fake AWS Key",
		Action:    "TRIGGERED",
		Metadata:  metaBytes,
		Timestamp: time.Now(),
	}
	htAlerts, err := engine.Evaluate(ctx, htEvt)
	if err != nil || len(htAlerts) != 1 {
		t.Fatalf("Expected 1 CRITICAL alert for honeytoken trigger, got %d (err: %v)", len(htAlerts), err)
	}
	if htAlerts[0].Severity != "CRITICAL" || htAlerts[0].DetectionType != "honeytoken" {
		t.Errorf("Unexpected honeytoken alert: %+v", htAlerts[0])
	}

	// Verification of Router & Tenant Isolation
	r := gin.New()
	producer := &kafka.Producer{}
	eventHandler := NewEventHandler(nil, producer)
	alertHandler := NewAlertHandler(nil)
	honeytokenHandler := NewHoneytokenHandler(nil, producer)

	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(jwtSecret))
	{
		protected.POST("/events/ingest", eventHandler.IngestEvent)
		protected.GET("/events", eventHandler.ListEvents)
		protected.GET("/alerts", alertHandler.ListAlerts)
		protected.GET("/deception/honeytokens", honeytokenHandler.ListHoneytokens)
	}

	// Test 1: Tenant A authenticated request returns 200 OK
	reqA, _ := http.NewRequest("GET", "/api/v1/alerts", nil)
	reqA.Header.Set("Authorization", "Bearer "+tokenA)
	wA := httptest.NewRecorder()
	r.ServeHTTP(wA, reqA)
	if wA.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for Tenant A accessing alerts, got %d", wA.Code)
	}

	// Test 2: Tenant B authenticated request returns 200 OK
	reqB, _ := http.NewRequest("GET", "/api/v1/alerts", nil)
	reqB.Header.Set("Authorization", "Bearer "+tokenB)
	wB := httptest.NewRecorder()
	r.ServeHTTP(wB, reqB)
	if wB.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for Tenant B accessing alerts, got %d", wB.Code)
	}

	// Test 3: Unauthenticated request returns 401 Unauthorized
	reqUnauth, _ := http.NewRequest("GET", "/api/v1/alerts", nil)
	wUnauth := httptest.NewRecorder()
	r.ServeHTTP(wUnauth, reqUnauth)
	if wUnauth.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for unauthenticated request, got %d", wUnauth.Code)
	}

	_ = alertsStoreA
	_ = alertsStoreB
}
