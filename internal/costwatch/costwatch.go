package costwatch

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/costwatchai/costwatch/internal/clickstore"
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
	log      *slog.Logger
	cs       *clickstore.Client
	tenantID string
	svcs     map[string]Service
}

func New(ctx context.Context, log *slog.Logger, cs *clickstore.Client, tenantID string) (*CostWatch, error) {
	return &CostWatch{
		log:      log,
		cs:       cs,
		tenantID: tenantID,
		svcs:     make(map[string]Service),
	}, nil
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
	batch, err := cw.cs.PrepareBatch(ctx, "insert into metrics (tenant_id, service, metric, value, timestamp)")
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
					cw.tenantID,
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
		where tenant_id = ? 
			and service = ? 
			and timestamp >= ? 
			and timestamp <= ?
		group by metric
	`

	rows, err := cw.cs.Query(ctx, query, cw.tenantID, svc.Label(), start, end)
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
