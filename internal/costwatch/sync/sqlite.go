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
		  service     text not null,
		  metric      text not null,
		  last_synced timestamp not null,
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

// Get returns the last synced timestamp and whether it exists.
func (s *Store) Get(ctx context.Context, service, metric string) (time.Time, bool, error) {
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

// Set upserts the last synced timestamp for the given key.
func (s *Store) Set(ctx context.Context, service, metric string, t time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		insert into sync_state(service, metric, last_synced)
		values(?, ?, ?)
		on conflict(service, metric) do update set last_synced=excluded.last_synced
	`, service, metric, t.UTC())
	return err
}
