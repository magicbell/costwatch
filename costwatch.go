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

type svcWithMetric struct {
	svc  Service
	mtrs map[string]*mtrWithUsage
}

var _ Usage = (*mtrWithUsage)(nil)

type mtrWithUsage struct {
	svc   Service
	units float64
	cost  float64
}

func (s *mtrWithUsage) Units() float64 {
	return s.units
}

func (s *mtrWithUsage) Cost() float64 {
	return s.cost
}

type CostWatch struct {
	log      *slog.Logger
	cs       *clickstore.Client
	tenantID string
	svcs     map[string]*svcWithMetric
}

func New(ctx context.Context, log *slog.Logger, cs *clickstore.Client, tenantID string) (*CostWatch, error) {
	return &CostWatch{
		log:      log,
		cs:       cs,
		tenantID: tenantID,
		svcs:     make(map[string]*svcWithMetric),
	}, nil
}

func (cw *CostWatch) RegisterService(svc Service) error {
	if _, exists := cw.svcs[svc.Label()]; exists {
		return ErrServiceAlreadyRegistered
	}

	cw.svcs[svc.Label()] = &svcWithMetric{
		svc:  svc,
		mtrs: make(map[string]*mtrWithUsage),
	}

	return nil
}

func (cw *CostWatch) FetchMetrics(ctx context.Context, start time.Time, end time.Time) error {
	// Start all services
	for _, s := range cw.svcs {
		for _, m := range s.svc.Metrics() {
			if _, exists := s.mtrs[m.Label()]; !exists {
				s.mtrs[m.Label()] = &mtrWithUsage{
					svc:   s.svc,
					units: 0,
					cost:  0,
				}
			}

			dps, err := m.Datapoints(ctx, m.Label(), start, end)
			if err != nil {
				return fmt.Errorf("m.Datapoints: %w", err)
			}

			mwu := s.mtrs[m.Label()]
			for _, dp := range dps {
				err = cw.cs.AsyncInsert(
					ctx,
					"insert into metrics (tenant_id, service, metric, value, timestamp) values (?, ?, ?, ?, ?)",
					false,
					cw.tenantID, s.svc.Label(), m.Label(), dp.Value, dp.Timestamp,
				)
				if err != nil {
					return fmt.Errorf("clickhouse.insert metrics: %w", err)
				}

				mwu.units += dp.Value
				mwu.cost += (dp.Value / m.UnitsPerPrice()) * m.Price()
			}
		}
	}

	return nil
}

func (cw *CostWatch) ServiceUsage(ctx context.Context, label string, start time.Time, end time.Time) (map[string]Usage, error) {
	svc := cw.svcs[label]

	usg := make(map[string]Usage)
	for l, m := range svc.mtrs {
		usg[l] = m
	}

	return usg, nil
}
