package detection

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"auth-service/internal/db"

	"github.com/google/uuid"
)

func TestDetectionEngine_BruteForceRule(t *testing.T) {
	ctx := context.Background()
	stateTracker := NewInMemoryStateTracker()
	rule := NewBruteForceRule(stateTracker, 3, 5*time.Minute)
	engine := NewEngine(rule)

	tenantID := uuid.New()
	targetIP := "192.168.1.100"

	createEvent := func() db.SecurityEvent {
		return db.SecurityEvent{
			ID:        uuid.New(),
			TenantID:  tenantID,
			EventType: "LOGIN_FAILED",
			Source:    "auth_service",
			SourceIp:  targetIP,
			UserID:    "user_123",
			Resource:  "login_endpoint",
			Action:    "LOGIN_FAILED",
			Severity:  "WARN",
			Timestamp: time.Now(),
		}
	}

	// Attempts 1, 2, 3: Should NOT generate alert
	for i := 1; i <= 3; i++ {
		event := createEvent()
		alerts, err := engine.Evaluate(ctx, event)
		if err != nil {
			t.Fatalf("Evaluate error on attempt %d: %v", i, err)
		}
		if len(alerts) != 0 {
			t.Errorf("Expected 0 alerts on attempt %d (below/at threshold 3), got %d", i, len(alerts))
		}
	}

	// Attempt 4: Threshold crossed -> Should generate HIGH alert
	event4 := createEvent()
	alerts4, err := engine.Evaluate(ctx, event4)
	if err != nil {
		t.Fatalf("Evaluate error on attempt 4: %v", err)
	}
	if len(alerts4) != 1 {
		t.Fatalf("Expected 1 alert on attempt 4 (threshold crossed), got %d", len(alerts4))
	}

	alert := alerts4[0]
	if alert.Severity != "HIGH" {
		t.Errorf("Expected severity HIGH, got %s", alert.Severity)
	}
	if alert.DetectionType != "brute_force" {
		t.Errorf("Expected detection_type brute_force, got %s", alert.DetectionType)
	}
	if alert.TenantID != tenantID {
		t.Errorf("Expected tenantID %s, got %s", tenantID, alert.TenantID)
	}

	// Attempt 5: Duplicate in same window -> Should NOT generate additional alert
	event5 := createEvent()
	alerts5, err := engine.Evaluate(ctx, event5)
	if err != nil {
		t.Fatalf("Evaluate error on attempt 5: %v", err)
	}
	if len(alerts5) != 0 {
		t.Errorf("Expected 0 alerts on attempt 5 (alert already sent), got %d", len(alerts5))
	}
}

func TestDetectionEngine_UnauthorizedAccessRule(t *testing.T) {
	ctx := context.Background()
	rule := NewUnauthorizedAccessRule()
	engine := NewEngine(rule)

	tenantID := uuid.New()
	event := db.SecurityEvent{
		ID:        uuid.New(),
		TenantID:  tenantID,
		EventType: "UNAUTHORIZED_ACCESS",
		Source:    "api_gateway",
		SourceIp:  "10.0.0.5",
		UserID:    "unauth_user",
		Resource:  "/api/v1/admin/secrets",
		Action:    "ACCESS_DENIED",
		Severity:  "HIGH",
		Status:    "FORBIDDEN",
		Timestamp: time.Now(),
	}

	alerts, err := engine.Evaluate(ctx, event)
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert for unauthorized access, got %d", len(alerts))
	}

	alert := alerts[0]
	if alert.Severity != "HIGH" {
		t.Errorf("Expected severity HIGH, got %s", alert.Severity)
	}
	if alert.DetectionType != "unauthorized_access" {
		t.Errorf("Expected detection_type unauthorized_access, got %s", alert.DetectionType)
	}
	if alert.AffectedAsset != "/api/v1/admin/secrets" {
		t.Errorf("Expected affected asset /api/v1/admin/secrets, got %s", alert.AffectedAsset)
	}
}

func TestDetectionEngine_HoneytokenRule(t *testing.T) {
	ctx := context.Background()
	rule := NewHoneytokenRule()
	engine := NewEngine(rule)

	tenantID := uuid.New()
	tokenID := uuid.New()

	metaMap := map[string]any{
		"honeytoken_id": tokenID.String(),
		"name":          "Production DB Pass",
		"type":          "DB_PASS",
	}
	metaBytes, _ := json.Marshal(metaMap)

	event := db.SecurityEvent{
		ID:        uuid.New(),
		TenantID:  tenantID,
		EventType: "HONEYTOKEN_TRIGGERED",
		Source:    "deception_engine",
		SourceIp:  "203.0.113.42",
		Resource:  "Production DB Pass",
		Action:    "TRIGGERED",
		Severity:  "CRITICAL",
		Metadata:  metaBytes,
		Timestamp: time.Now(),
	}

	alerts, err := engine.Evaluate(ctx, event)
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert for honeytoken trigger, got %d", len(alerts))
	}

	alert := alerts[0]
	if alert.Severity != "CRITICAL" {
		t.Errorf("Expected severity CRITICAL, got %s", alert.Severity)
	}
	if alert.DetectionType != "honeytoken" {
		t.Errorf("Expected detection_type honeytoken, got %s", alert.DetectionType)
	}
	if alert.TenantID != tenantID {
		t.Errorf("Expected tenantID %s, got %s", tenantID, alert.TenantID)
	}
}
