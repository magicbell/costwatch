package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"time"

	"github.com/costwatchai/costwatch/internal/costwatch/port"
	"github.com/costwatchai/costwatch/internal/sqlstore"
)

type AlertsRepos struct{ st *sqlstore.Store }

func NewAlertsRepos(st *sqlstore.Store) *AlertsRepos { return &AlertsRepos{st: st} }

//go:embed sql/list_alert_rules.sql
var listAlertRulesSQL string

// Rules
func (r *AlertsRepos) ListRules(ctx context.Context) ([]port.AlertRule, error) {
	rows, err := r.st.DB().QueryContext(ctx, listAlertRulesSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []port.AlertRule
	for rows.Next() {
		var rec port.AlertRule
		if err := rows.Scan(&rec.Service, &rec.Metric, &rec.Threshold); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

//go:embed sql/upsert_alert_rule.sql
var upsertAlertRuleSQL string

func (r *AlertsRepos) UpsertRule(ctx context.Context, ar port.AlertRule) error {
	_, err := r.st.DB().ExecContext(ctx, upsertAlertRuleSQL, ar.Service, ar.Metric, ar.Threshold)
	return err
}

//go:embed sql/get_last_notified.sql
var getLastNotified string

// Notifications
func (r *AlertsRepos) GetLastNotified(ctx context.Context, s, m string) (int64, bool, error) {
	row := r.st.DB().QueryRowContext(ctx, getLastNotified, s, m)
	var ts sql.NullTime
	if err := row.Scan(&ts); err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}
	if !ts.Valid {
		return 0, false, nil
	}
	return ts.Time.UTC().Unix(), true, nil
}

//go:embed sql/set_last_notified.sql
var setLastNotified string

func (r *AlertsRepos) SetLastNotified(ctx context.Context, s, m string, unix int64) error {
	_, err := r.st.DB().ExecContext(ctx, setLastNotified, time.Unix(unix, 0).UTC(), s, m)
	return err
}
