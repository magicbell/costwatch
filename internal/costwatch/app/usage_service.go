package app

import (
	"context"
	"time"

	"github.com/costwatchai/costwatch/internal/costwatch/port"
)

// UsageItem represents a costed metric bucket for API consumption.
type UsageItem struct {
	Service   string
	Metric    string
	Timestamp time.Time
	Cost      float64
}

// PercentileCost represents cost percentiles per service/metric.
type PercentileCost struct {
	Service string
	Metric  string
	P50     float64
	P90     float64
	P95     float64
	PMax    float64
}

type UsageService struct {
	Metrics port.MetricsRepo
	Catalog port.Catalog
}

func NewUsageService(metrics port.MetricsRepo, catalog port.Catalog) *UsageService {
	return &UsageService{Metrics: metrics, Catalog: catalog}
}

// Usage aggregates units per bucket and converts them to cost via the catalog.
func (s *UsageService) Usage(ctx context.Context, start, end time.Time, bucket time.Duration) ([]UsageItem, error) {
	recs, err := s.Metrics.Aggregate(ctx, start, end, bucket)
	if err != nil {
		return nil, err
	}
	out := make([]UsageItem, 0, len(recs))
	for _, r := range recs {
		cost, _ := s.Catalog.ComputeCost(r.Service, r.Metric, r.Units)
		out = append(out, UsageItem{Service: r.Service, Metric: r.Metric, Timestamp: r.Timestamp, Cost: cost})
	}
	return out, nil
}

// UsagePercentiles computes unit percentiles and converts them to cost via the catalog.
func (s *UsageService) UsagePercentiles(ctx context.Context, start, end time.Time, bucket time.Duration) ([]PercentileCost, error) {
	recs, err := s.Metrics.Percentiles(ctx, start, end, bucket)
	if err != nil {
		return nil, err
	}
	out := make([]PercentileCost, 0, len(recs))
	for _, r := range recs {
		c50, _ := s.Catalog.ComputeCost(r.Service, r.Metric, r.P50)
		c90, _ := s.Catalog.ComputeCost(r.Service, r.Metric, r.P90)
		c95, _ := s.Catalog.ComputeCost(r.Service, r.Metric, r.P95)
		cmax, _ := s.Catalog.ComputeCost(r.Service, r.Metric, r.PMax)
		out = append(out, PercentileCost{Service: r.Service, Metric: r.Metric, P50: c50, P90: c90, P95: c95, PMax: cmax})
	}
	return out, nil
}
