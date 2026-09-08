package detection

import (
	"context"

	"auth-service/internal/db"
)

// Rule defines the interface for detection rules evaluated against security events
type Rule interface {
	// Name returns unique identifier for the rule
	Name() string

	// Evaluate evaluates a security event and returns an alert parameter if rule triggers
	Evaluate(ctx context.Context, event db.SecurityEvent) (*db.CreateAlertParams, bool, error)
}
