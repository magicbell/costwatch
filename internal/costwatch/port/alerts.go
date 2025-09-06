package port

import "context"

// AlertRule defines a per-hour threshold for a service/metric.
type AlertRule struct {
	Service   string
	Metric    string
	Threshold float64
}

// AlertRuleRepo stores and retrieves alert rules.
type AlertsRepo interface {
	ListRules(ctx context.Context) ([]AlertRule, error)
	UpsertRule(ctx context.Context, r AlertRule) error
	GetLastNotified(ctx context.Context, service, metric string) (int64, bool, error)
	SetLastNotified(ctx context.Context, service, metric string, timeUnix int64) error
}
