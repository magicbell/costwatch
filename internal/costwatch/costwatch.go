package costwatch

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/costwatchai/costwatch/internal/clickstore"
	cwsync "github.com/costwatchai/costwatch/internal/costwatch/sync"
)

var ErrServiceAlreadyRegistered = fmt.Errorf("service already registered")

type Usage interface {
	Units() float64
	Cost() float64
}

var _ Usage = (*ServiceUsage)(nil)

type ServiceUsage struct {
	units float64
	cost  float64
}

func (s *ServiceUsage) Units() float64 {
	return s.units
}

func (s *ServiceUsage) Cost() float64 {
	return s.cost
}

type CostWatch struct {
	log       *slog.Logger
	cs        *clickstore.Client
	svcs      map[string]Service
	syncStore *cwsync.Store
}

// Services returns a snapshot of registered services.
func (cw *CostWatch) Services() []Service {
	res := make([]Service, 0, len(cw.svcs))
	for _, s := range cw.svcs {
		res = append(res, s)
	}
	return res
}

func New(ctx context.Context, log *slog.Logger, cs *clickstore.Client) (*CostWatch, error) {
	return &CostWatch{
		log:  log,
		cs:   cs,
		svcs: make(map[string]Service),
	}, nil
}

// getSyncStore returns a lazily-initialized sync-state store.
func (cw *CostWatch) getSyncStore() (*cwsync.Store, error) {
	if cw.syncStore != nil {
		return cw.syncStore, nil
	}
	path := os.Getenv("COSTWATCH_STATE_PATH")
	if path == "" {
		path = ".db/costwatch_state.db"
	}
	st, err := cwsync.Open(path)
	if err != nil {
		return nil, err
	}
	cw.syncStore = st
	return st, nil
}

// Sync performs a full sync across services/metrics using the sync-state to compute windows.
// Rules:
// - End is always "now" (UTC).
// - Start is the oldest (earliest) of [last-synced, now-15m] to ensure we fetch at least 15 minutes.
// - If no last-synced exists, default to a first-time lookback (currently 48 hours) to backfill history.
func (cw *CostWatch) Sync(ctx context.Context) error {
	st, err := cw.getSyncStore()
	if err != nil {
		return fmt.Errorf("open syncstate: %w", err)
	}
	now := time.Now().UTC()
	fifteenAgo := now.Add(-15 * time.Minute)
	for _, s := range cw.Services() {
		for _, m := range s.Metrics() {
   last, ok, err := st.Get(ctx, s.Label(), m.Label())
			if err != nil {
				cw.log.Error("syncstate.Get error", "service", s.Label(), "metric", m.Label(), "error", err)
				continue
			}
			start := last
			if !ok || last.IsZero() {
				start = now.Add(-48 * time.Hour)
			} else if last.Before(fifteenAgo) {
				start = fifteenAgo
			}
			end := now
			if !start.Before(end) {
				continue
			}
			cw.log.Info("fetching metric", "service", s.Label(), "metric", m.Label(), "start", start, "end", end)
			if err := cw.FetchMetricForService(ctx, s, m, start, end); err != nil {
				cw.log.Error("FetchMetricForService failed", "service", s.Label(), "metric", m.Label(), "error", err)
				continue
			}
   if err := st.Set(ctx, s.Label(), m.Label(), end); err != nil {
				cw.log.Error("syncstate.Set error", "service", s.Label(), "metric", m.Label(), "error", err)
			}
		}
	}
	return nil
}

func (cw *CostWatch) RegisterService(svc Service) error {
	if _, exists := cw.svcs[svc.Label()]; exists {
		return ErrServiceAlreadyRegistered
	}

	cw.svcs[svc.Label()] = svc

	return nil
}

func (cw *CostWatch) FetchMetrics(ctx context.Context, start time.Time, end time.Time) error {
	// Prepare batch for bulk insertion
	batch, err := cw.cs.PrepareBatch(ctx, "insert into metrics (service, metric, value, timestamp)")
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	// Collect all datapoints and add to batch
	for _, s := range cw.svcs {
		for _, m := range s.Metrics() {
			dps, err := m.Datapoints(ctx, m.Label(), start, end)
			if err != nil {
				return fmt.Errorf("m.Datapoints: %w", err)
			}

			// Add all datapoints to the batch
			for _, dp := range dps {
    if err := batch.Append(
					s.Label(),
					m.Label(),
					dp.Value,
					dp.Timestamp,
				); err != nil {
					return fmt.Errorf("batch append: %w", err)
				}
			}
		}
	}

	// Send the batch to ClickHouse
	if err := batch.Send(); err != nil {
		return fmt.Errorf("batch send: %w", err)
	}

	return nil
}

// FetchMetricsForService is like FetchMetrics but only for the given service.
func (cw *CostWatch) FetchMetricsForService(ctx context.Context, svc Service, start time.Time, end time.Time) error {
	if svc == nil {
		return fmt.Errorf("nil service")
	}

	// Prepare a batch for bulk insertion
	batch, err := cw.cs.PrepareBatch(ctx, "insert into metrics (service, metric, value, timestamp)")
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}
	for _, m := range svc.Metrics() {
		dps, err := m.Datapoints(ctx, m.Label(), start, end)
		if err != nil {
			return fmt.Errorf("m.Datapoints: %w", err)
		}
		for _, dp := range dps {
			if err := batch.Append(
				svc.Label(),
				m.Label(),
				dp.Value,
				dp.Timestamp,
			); err != nil {
				return fmt.Errorf("batch append: %w", err)
			}
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("batch send: %w", err)
	}
	return nil
}

// FetchMetricForService ingests datapoints only for the specified metric of a service.
func (cw *CostWatch) FetchMetricForService(ctx context.Context, svc Service, m Metric, start time.Time, end time.Time) error {
	if svc == nil || m == nil {
		return fmt.Errorf("nil service or metric")
	}
 batch, err := cw.cs.PrepareBatch(ctx, "insert into metrics (service, metric, value, timestamp)")
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}
	dps, err := m.Datapoints(ctx, m.Label(), start, end)
	if err != nil {
		return fmt.Errorf("m.Datapoints: %w", err)
	}
	for _, dp := range dps {
		if err := batch.Append(
			svc.Label(),
			m.Label(),
			dp.Value,
			dp.Timestamp,
		); err != nil {
			return fmt.Errorf("batch append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("batch send: %w", err)
	}
	return nil
}

func (cw *CostWatch) ServiceUsage(ctx context.Context, svc Service, start time.Time, end time.Time) (map[string]Usage, error) {
	svc, exists := cw.svcs[svc.Label()]
	if !exists {
		return nil, fmt.Errorf("service %s not registered", svc.Label())
	}

	// Query ClickHouse for aggregated metrics data
 query := `
		select 
			metric,
			sum(value) as total_units
		from metrics 
		where service = ? 
			and timestamp >= ? 
			and timestamp <= ?
		group by metric
	`

 rows, err := cw.cs.Query(ctx, query, svc.Label(), start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics from ClickHouse: %w", err)
	}
	defer rows.Close()

	usg := make(map[string]Usage)

	// Process each metric result
	for rows.Next() {
		var metricName string
		var totalUnits float64

		err := rows.Scan(&metricName, &totalUnits)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		// Find the metric definition to get pricing information
		var match Metric
		for _, m := range svc.Metrics() {
			if m.Label() == metricName {
				match = m
				break
			}
		}

		if match == nil {
			return nil, fmt.Errorf("no matching metric found for %s", metricName)
		}

		// Calculate cost using the metric's pricing info
		totalCost := (totalUnits / match.UnitsPerPrice()) * match.Price()

		usg[metricName] = &ServiceUsage{
			units: totalUnits,
			cost:  totalCost,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return usg, nil
}
