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

type UnauthorizedAccessRule struct{}

func NewUnauthorizedAccessRule() *UnauthorizedAccessRule {
	return &UnauthorizedAccessRule{}
}

func (r *UnauthorizedAccessRule) Name() string {
	return "unauthorized_access"
}

func (r *UnauthorizedAccessRule) Evaluate(ctx context.Context, event db.SecurityEvent) (*db.CreateAlertParams, bool, error) {
	eventType := strings.ToUpper(event.EventType)
	action := strings.ToUpper(event.Action)
	status := strings.ToUpper(event.Status)

	isUnauthorized := eventType == "UNAUTHORIZED_ACCESS" || eventType == "ACCESS_DENIED" || eventType == "UNAUTHORIZED" ||
		action == "UNAUTHORIZED_ACCESS" || action == "ACCESS_DENIED" ||
		status == "UNAUTHORIZED" || status == "FORBIDDEN" || status == "ACCESS_DENIED"

	if !isUnauthorized {
		return nil, false, nil
	}

	asset := event.Resource
	if asset == "" {
		asset = "Restricted Resource"
	}

	metaMap := map[string]any{
		"user_id":         event.UserID,
		"source_ip":       event.SourceIp,
		"target_resource": asset,
		"action":          event.Action,
		"event_type":      event.EventType,
	}
	if len(event.Metadata) > 0 {
		var incomingMeta map[string]any
		if err := json.Unmarshal(event.Metadata, &incomingMeta); err == nil {
			for k, v := range incomingMeta {
				metaMap[k] = v
			}
		}
	}
	metaBytes, _ := json.Marshal(metaMap)

	userContext := event.UserID
	if userContext == "" {
		userContext = "anonymous/unknown user"
	}

	alertParams := &db.CreateAlertParams{
		ID:            uuid.New(),
		TenantID:      event.TenantID,
		Severity:      "HIGH",
		DetectionType: "unauthorized_access",
		SourceIp:      event.SourceIp,
		AffectedAsset: asset,
		Description:   fmt.Sprintf("Unauthorized access attempt detected on resource '%s' by %s from IP %s.", asset, userContext, event.SourceIp),
		Status:        "OPEN",
		Metadata:      metaBytes,
		Timestamp:     time.Now(),
	}

	return alertParams, true, nil
}
