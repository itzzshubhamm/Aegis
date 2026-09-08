package detection

import (
	"context"
	"log/slog"

	"auth-service/internal/db"
)

type Engine struct {
	rules []Rule
}

func NewEngine(rules ...Rule) *Engine {
	return &Engine{
		rules: rules,
	}
}

func (e *Engine) RegisterRule(rule Rule) {
	e.rules = append(e.rules, rule)
}

func (e *Engine) Evaluate(ctx context.Context, event db.SecurityEvent) ([]db.CreateAlertParams, error) {
	var alerts []db.CreateAlertParams

	for _, rule := range e.rules {
		alertParams, triggered, err := rule.Evaluate(ctx, event)
		if err != nil {
			slog.Error("Error evaluating detection rule", "rule", rule.Name(), "event_id", event.ID, "error", err)
			continue
		}

		if triggered && alertParams != nil {
			slog.Info("Detection rule triggered alert",
				"rule", rule.Name(),
				"event_id", event.ID,
				"tenant_id", event.TenantID,
				"severity", alertParams.Severity,
				"detection_type", alertParams.DetectionType,
			)
			alerts = append(alerts, *alertParams)
		}
	}

	return alerts, nil
}
