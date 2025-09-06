package sqlstore

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

// DB returns the underlying sql.DB. Intended for infra adapters to run queries.
func (s *Store) DB() *sql.DB { return s.db }


// Open opens (and creates if necessary) a SQLite database at the given path and ensures schema.
func Open() (*Store, error) {
	path := os.Getenv("SQLITE_DB_PATH")
	if path == "" {
		path = ".db/costwatch.db"
	}

	// Ensure parent directory exists if using nested path (e.g., .db/costwatch.db)
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

//go:embed sql/schema.sql
var schemaSQL string

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
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

