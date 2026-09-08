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

type HoneytokenRule struct{}

func NewHoneytokenRule() *HoneytokenRule {
	return &HoneytokenRule{}
}

func (r *HoneytokenRule) Name() string {
	return "honeytoken"
}

func (r *HoneytokenRule) Evaluate(ctx context.Context, event db.SecurityEvent) (*db.CreateAlertParams, bool, error) {
	eventType := strings.ToUpper(event.EventType)
	action := strings.ToUpper(event.Action)

	isHoneytoken := eventType == "HONEYTOKEN_TRIGGERED" || eventType == "HONEYTOKEN_ACCESSED" || eventType == "DECEPTION_TRIGGERED" ||
		action == "HONEYTOKEN_TRIGGERED" || action == "TRIGGERED"

	if !isHoneytoken {
		return nil, false, nil
	}

	asset := event.Resource
	if asset == "" {
		asset = "Deception Honeytoken"
	}

	metaMap := map[string]any{
		"event_id":   event.ID,
		"source_ip":  event.SourceIp,
		"resource":   asset,
		"event_type": event.EventType,
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

	tokenIDStr := ""
	if tid, ok := metaMap["honeytoken_id"]; ok {
		tokenIDStr = fmt.Sprintf("%v", tid)
	}

	desc := fmt.Sprintf("CRITICAL: Honeytoken deception mechanism triggered for asset '%s' from IP %s.", asset, event.SourceIp)
	if tokenIDStr != "" {
		desc = fmt.Sprintf("CRITICAL: Honeytoken (ID: %s) triggered for asset '%s' from IP %s.", tokenIDStr, asset, event.SourceIp)
	}

	alertParams := &db.CreateAlertParams{
		ID:            uuid.New(),
		TenantID:      event.TenantID,
		Severity:      "CRITICAL",
		DetectionType: "honeytoken",
		SourceIp:      event.SourceIp,
		AffectedAsset: asset,
		Description:   desc,
		Status:        "OPEN",
		Metadata:      metaBytes,
		Timestamp:     time.Now(),
	}

	return alertParams, true, nil
}
