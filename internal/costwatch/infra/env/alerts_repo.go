package env

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/costwatchai/costwatch/internal/costwatch/port"
)

// AlertsRepos implements port.AlertsRepo backed by environment variables.
//
// It reads alert rules from the ALERT_RULES environment variable as a JSON array
// of objects: [{"service":"aws.CloudWatch","metric":"IncomingBytes","threshold":0.47}].
// This provider is read-only: UpsertRule returns an error. Notification state is ignored.
//
// Environment key: ALERT_RULES

type AlertsRepos struct{}

func NewAlertsRepos() *AlertsRepos { return &AlertsRepos{} }

type rule struct {
	Service   string  `json:"service"`
	Metric    string  `json:"metric"`
	Threshold float64 `json:"threshold"`
}

var ErrReadOnly = errors.New("env alerts repo is read-only")

func (r *AlertsRepos) ListRules(_ context.Context) ([]port.AlertRule, error) {
	val := os.Getenv("ALERT_RULES")
	if val == "" {
		return nil, nil
	}
	var rs []rule
	if err := json.Unmarshal([]byte(val), &rs); err != nil {
		return nil, err
	}
	out := make([]port.AlertRule, 0, len(rs))
	for _, it := range rs {
		out = append(out, port.AlertRule{Service: it.Service, Metric: it.Metric, Threshold: it.Threshold})
	}
	return out, nil
}

func (r *AlertsRepos) UpsertRule(_ context.Context, _ port.AlertRule) error {
	return ErrReadOnly
}

func (r *AlertsRepos) GetLastNotified(_ context.Context, _, _ string) (int64, bool, error) {
	return 0, false, nil
}

func (r *AlertsRepos) SetLastNotified(_ context.Context, _, _ string, _ int64) error {
	return nil
}
