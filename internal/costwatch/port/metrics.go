package port

import (
	"context"
	"time"
)

// MetricBucket is an aggregated data point for a service/metric at a specific bucket timestamp.
type MetricBucket struct {
	Service   string
	Metric    string
	Timestamp time.Time
	Units     float64
}

// MetricPercentiles holds usage percentiles per service/metric over a time range.
type MetricPercentiles struct {
	Service string
	Metric  string
	P50     float64
	P90     float64
	P95     float64
	PMax    float64
}

// MetricsQueryPort defines access to aggregated metrics storage (e.g., ClickHouse).
type MetricsRepo interface {
	Aggregate(ctx context.Context, start, end time.Time, bucket time.Duration) ([]MetricBucket, error)
	Percentiles(ctx context.Context, start, end time.Time, bucket time.Duration) ([]MetricPercentiles, error)
}
