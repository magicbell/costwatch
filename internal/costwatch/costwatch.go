package costwatch

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/costwatchai/costwatch/internal/clickstore"
	appsvc "github.com/costwatchai/costwatch/internal/costwatch/app"
	chinfra "github.com/costwatchai/costwatch/internal/costwatch/infra/clickhouse"
	envinfra "github.com/costwatchai/costwatch/internal/costwatch/infra/env"
	notinfr "github.com/costwatchai/costwatch/internal/costwatch/infra/notifier"
	sqlinfra "github.com/costwatchai/costwatch/internal/costwatch/infra/sqlite"
	"github.com/costwatchai/costwatch/internal/costwatch/port"
	"github.com/costwatchai/costwatch/internal/sqlstore"
)

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
	log *slog.Logger
	cs  *clickstore.Client
	db  *sqlstore.Store
}

func New(ctx context.Context, log *slog.Logger, cs *clickstore.Client) (*CostWatch, error) {
	return &CostWatch{
		log: log,
		cs:  cs,
	}, nil
}

// registryCatalog implements MetricCatalog using the global registry in this package.
type registryCatalog struct{}

func (registryCatalog) ComputeCost(service, metric string, units float64) (float64, bool) {
	return ComputeCost(service, metric, units)
}

// getSyncStore returns a lazily-initialized sync-state store.
func (cw *CostWatch) getSyncStore() (*sqlstore.Store, error) {
	if cw.db != nil {
		return cw.db, nil
	}

	db, err := sqlstore.Open()
	if err != nil {
		return nil, err
	}

	cw.db = db
	return db, nil
}

func (cw *CostWatch) FetchMetrics(ctx context.Context, start time.Time, end time.Time) error {
	// Prepare batch for bulk insertion
	batch, err := cw.cs.PrepareBatch(ctx, "insert into metrics (service, metric, value, timestamp)")
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	// Collect all datapoints and add to batch
	for _, s := range ListServices() {
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
	svc, exists := FindService(svc.Label())
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

// sendAlerts uses the AlertService with ports/adapters to compute windows and send notifications.
func (cw *CostWatch) sendAlerts(ctx context.Context) error {
	st, err := cw.getSyncStore()
	if err != nil {
		return err
	}

	// Wire ports
	m := chinfra.NewMetricsRepo(cw.cs)
	var a port.AlertsRepo
	if os.Getenv("ALERT_RULES") != "" {
		a = envinfra.NewAlertsRepos()
	} else {
		a = sqlinfra.NewAlertsRepos(st)
	}
	n := notinfr.NewWebhookNotifierFromEnv()
	c := registryCatalog{}
	alerts := appsvc.NewAlertService(m, a, n, c)
	return alerts.SendAlerts(ctx)
}
