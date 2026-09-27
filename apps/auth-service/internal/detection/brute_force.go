package detection

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"auth-service/internal/db"

	"github.com/google/uuid"
)

type BruteForceRule struct {
	tracker        StateTracker
	threshold      int           // Number of failed attempts threshold (e.g., 3)
	windowDuration time.Duration // Time window (e.g., 5 minutes)
}

func NewBruteForceRule(tracker StateTracker, threshold int, windowDuration time.Duration) *BruteForceRule {
	if threshold <= 0 {
		threshold = 3
	}
	if windowDuration <= 0 {
		windowDuration = 5 * time.Minute
	}
	return &BruteForceRule{
		tracker:        tracker,
		threshold:      threshold,
		windowDuration: windowDuration,
	}
}

func (r *BruteForceRule) Name() string {
	return "brute_force"
}

func (r *BruteForceRule) Evaluate(ctx context.Context, event db.SecurityEvent) (*db.CreateAlertParams, bool, error) {
	// Check if this event represents a failed login attempt
	eventType := strings.ToUpper(event.EventType)
	action := strings.ToUpper(event.Action)

	isFailedLogin := eventType == "LOGIN_FAILED" || eventType == "USER_LOGIN_FAILED" ||
		action == "LOGIN_FAILED" || action == "FAILED_LOGIN"

	if !isFailedLogin {
		return nil, false, nil
	}

	// Identify target by UserID or SourceIp
	targetKey := event.UserID
	if targetKey == "" {
		targetKey = event.SourceIp
	}
	if targetKey == "" {
		targetKey = "unknown"
	}

	tenantStr := event.TenantID.String()

	// Increment failed login count
	count, err := r.tracker.IncrementFailedLogin(ctx, tenantStr, targetKey, r.windowDuration)
	if err != nil {
		return nil, false, fmt.Errorf("failed to increment failed login counter: %w", err)
	}

	// Check if alert has already been generated in this window to avoid duplicate alerts
	alertSent, err := r.tracker.HasAlertBeenSent(ctx, tenantStr, targetKey)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check alert status: %w", err)
	}

	// Only generate an alert when count exceeds threshold AND no alert has been sent yet
	if count > int64(r.threshold) && !alertSent {
		// Mark alert as sent for the duration of the window
		if err := r.tracker.SetAlertSent(ctx, tenantStr, targetKey, r.windowDuration); err != nil {
			// Log error but proceed with alert creation
			_ = err
		}

		asset := event.Resource
		if asset == "" {
			asset = targetKey
		}

		metaMap := map[string]any{
			"attempt_count": count,
			"threshold":     r.threshold,
			"window":        r.windowDuration.String(),
			"source_ip":     event.SourceIp,
			"user_id":       event.UserID,
		}
		metaBytes, _ := json.Marshal(metaMap)

		alertParams := &db.CreateAlertParams{
			ID:            uuid.New(),
			TenantID:      event.TenantID,
			Severity:      "HIGH",
			DetectionType: "brute_force",
			SourceIp:      event.SourceIp,
			AffectedAsset: asset,
			Description:   fmt.Sprintf("Brute force attack detected: %d failed login attempts for target '%s' within %s window.", count, targetKey, r.windowDuration),
			Status:        "OPEN",
			Metadata:      metaBytes,
			Timestamp:     time.Now(),
		}

		return alertParams, true, nil
	}

	return nil, false, nil
}
