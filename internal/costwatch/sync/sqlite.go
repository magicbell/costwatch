package cwsync

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Store persists last-synced timestamps per service/metric in SQLite.
// It is a minimal persistence layer to avoid double fetching windows.
//
// Schema:
//   create table if not exists sync_state (
//     service     text not null,
//     metric      text not null,
//     last_synced timestamp not null,
//     primary key (service, metric)
//   );
//
// Timestamps are stored in UTC.

type Store struct {
	db *sql.DB
}

type AlertRule struct {
	Service   string
	Metric    string
	Threshold float64
}

// GetAlertThreshold returns the configured threshold for a service/metric if present.
func (s *Store) GetAlertThreshold(ctx context.Context, service, metric string) (float64, bool, error) {
	row := s.db.QueryRowContext(ctx, `select threshold from alert_rules where service=? and metric=?`, service, metric)
	var th float64
	err := row.Scan(&th)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return th, true, nil
}

// GetAlertRules returns all configured alert rules from the database.
// Each alert rule consists of a service, a metric, and a threshold.
// If no rules exist, it returns an empty slice.
func (s *Store) GetAlertRules(ctx context.Context) ([]AlertRule, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT service, metric, threshold FROM alert_rules`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []AlertRule
	for rows.Next() {
		var r AlertRule
		if err := rows.Scan(&r.Service, &r.Metric, &r.Threshold); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rules, nil
}

// SetAlertThreshold upserts the threshold for service/metric.
func (s *Store) SetAlertThreshold(ctx context.Context, service, metric string, threshold float64) error {
	_, err := s.db.ExecContext(ctx, `
		insert into alert_rules(service, metric, threshold)
		values(?, ?, ?)
		on conflict(service, metric) do update set threshold=excluded.threshold
	`, service, metric, threshold)
	return err
}

// Open opens (and creates if necessary) a SQLite database at the given path and ensures schema.
func Open(path string) (*Store, error) {
	// Ensure parent directory exists if using nested path (e.g., .db/costwatch_state.db)
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// Busy timeout to reduce lock errors if multiple processes accidentally open it.
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := ensureSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(`
		create table if not exists sync_state (
		  service        text not null,
		  metric         text not null,
		  last_synced    timestamp not null,
		  last_notified  timestamp,
		  primary key (service, metric)
		);

		create table if not exists alert_rules (
		  service   text not null,
		  metric    text not null,
		  threshold real not null,
		  primary key (service, metric)
		);
	`)
	return err
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// GetLastSync returns the last synced timestamp and whether it exists.
func (s *Store) GetLastSync(ctx context.Context, service, metric string) (time.Time, bool, error) {
	row := s.db.QueryRowContext(ctx, `select last_synced from sync_state where service=? and metric=?`, service, metric)
	var ts time.Time
	err := row.Scan(&ts)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, err
	}
	return ts.UTC(), true, nil
}

// SetLastSync upserts the last synced timestamp for the given key.
func (s *Store) SetLastSync(ctx context.Context, service, metric string, t time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		insert into sync_state(service, metric, last_synced)
		values(?, ?, ?)
		on conflict(service, metric) do update set last_synced=excluded.last_synced
	`, service, metric, t.UTC())
	return err
}

// GetLastNotified returns the last time an alert was sent for the given service/metric.
func (s *Store) GetLastNotified(ctx context.Context, service, metric string) (time.Time, bool, error) {
	row := s.db.QueryRowContext(ctx, `select last_notified from sync_state where service=? and metric=?`, service, metric)
	var ts sql.NullTime
	err := row.Scan(&ts)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, err
	}
	if !ts.Valid {
		return time.Time{}, false, nil
	}
	return ts.Time.UTC(), true, nil
}

// SetLastNotified upserts the last time an alert was sent for the given service/metric.
func (s *Store) SetLastNotified(ctx context.Context, service, metric string, t time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		update sync_state set last_notified = ?
		where service = ? and metric = ?
	`, t.UTC(), service, metric)
	return err
}
