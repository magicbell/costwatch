package clickstore
import (
	"context"
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

var _ driver.Conn = (*TestStore)(nil)

// emptyRows implements driver.Rows and represents an empty result set.
type emptyRows struct{}

func (e *emptyRows) Next() bool                     { return false }
func (e *emptyRows) Scan(dest ...any) error         { return nil }
func (e *emptyRows) ScanStruct(dest any) error      { return nil }
func (e *emptyRows) Columns() []string              { return nil }
func (e *emptyRows) ColumnTypes() []driver.ColumnType { return nil }
func (e *emptyRows) Totals(dest ...any) error       { return nil }
func (e *emptyRows) Err() error                     { return nil }
func (e *emptyRows) Close() error                   { return nil }

// emptyRow implements driver.Row and represents an empty single-row result.
type emptyRow struct{}

func (e *emptyRow) Scan(dest ...any) error     { return nil }
func (e *emptyRow) ScanStruct(dest any) error  { return nil }
func (e *emptyRow) Err() error                 { return nil }

// TestStore is a lightweight in-memory test connection that logs queries.
type TestStore struct {
	log *slog.Logger
}

func NewTestStore(log *slog.Logger) *TestStore {
	return &TestStore{
		log: log,
	}
}

// AsyncInsert implements driver.Conn.
func (t *TestStore) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	t.log.Debug("teststore.AsyncInsert", "query", query, "wait", wait, "args", args)
	return nil
}

// Close implements driver.Conn.
func (t *TestStore) Close() error {
	t.log.Debug("Close called")
	return nil
}

// Contributors implements driver.Conn.
func (t *TestStore) Contributors() []string {
	t.log.Debug("Contributors called")
	return []string{"Test Contributor"}
}

// Exec implements driver.Conn.
func (t *TestStore) Exec(ctx context.Context, query string, args ...any) error {
	t.log.Debug("Exec called with query: %s, args: %v", query, args)
	return nil
}

// Ping implements driver.Conn.
func (t *TestStore) Ping(context.Context) error {
	t.log.Debug("Ping called")
	return nil
}

// PrepareBatch implements driver.Conn.
func (t *TestStore) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	t.log.Debug("PrepareBatch called with query: %s, opts: %v", query, opts)
	return nil, nil
}

// Query implements driver.Conn.
func (t *TestStore) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	t.log.Debug("Query called with query: %s, args: %v", query, args)
	return &emptyRows{}, nil
}

// QueryRow implements driver.Conn.
func (t *TestStore) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	t.log.Debug("QueryRow called with query: %s, args: %v", query, args)
	return &emptyRow{}
}

// Select implements driver.Conn.
func (t *TestStore) Select(ctx context.Context, dest any, query string, args ...any) error {
	t.log.Debug("Select called with dest: %v, query: %s, args: %v", dest, query, args)
	return nil
}

// ServerVersion implements driver.Conn.
func (t *TestStore) ServerVersion() (*driver.ServerVersion, error) {
	t.log.Debug("ServerVersion called")
	return nil, nil
}

// Stats implements driver.Conn.
func (t *TestStore) Stats() driver.Stats {
	t.log.Debug("Stats called")
	return driver.Stats{}
}
