package costwatch

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/tailbits/costwatch/internal/clickstore"
	appsvc "github.com/tailbits/costwatch/internal/costwatch/app"
	chinfra "github.com/tailbits/costwatch/internal/costwatch/infra/clickhouse"
	envinfra "github.com/tailbits/costwatch/internal/costwatch/infra/env"
	notinfr "github.com/tailbits/costwatch/internal/costwatch/infra/notifier"
	sqlinfra "github.com/tailbits/costwatch/internal/costwatch/infra/sqlite"
	"github.com/tailbits/costwatch/internal/costwatch/port"
	"github.com/tailbits/costwatch/internal/sqlstore"
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

// FetchMetricsForService is like FetchMetrics but only for the given service.
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
