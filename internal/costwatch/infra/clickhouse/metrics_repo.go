package clickhouse

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/costwatchai/costwatch/internal/clickstore"
	"github.com/costwatchai/costwatch/internal/costwatch/port"
)

type MetricsRepo struct {
	db *clickstore.Client
}

func NewMetricsRepo(db *clickstore.Client) *MetricsRepo {
	return &MetricsRepo{
		db: db,
	}
}

//go:embed sql/aggregate.sql
var aggregateSQL string

func (q *MetricsRepo) Aggregate(ctx context.Context, start, end time.Time, bucket time.Duration) ([]port.MetricBucket, error) {
	var rows []struct {
		Service   string    `ch:"service"`
		Metric    string    `ch:"metric"`
		Timestamp time.Time `ch:"ts"`
		Units     float64   `ch:"units"`
	}

	excludes := []string{}
	if disableDemo() {
		excludes = append(excludes, "coingecko")
	}

	if err := q.db.Select(ctx, &rows, aggregateSQL, int(bucket.Seconds()), start, end, excludes); err != nil {
		return nil, fmt.Errorf("clickhouse.Select: %w", err)
	}

	out := make([]port.MetricBucket, 0, len(rows))
	for _, r := range rows {
		out = append(out, port.MetricBucket{
			Service:   r.Service,
			Metric:    r.Metric,
			Timestamp: r.Timestamp,
			Units:     r.Units,
		})
	}

	return out, nil
}

//go:embed sql/percentiles.sql
var percentilesSQL string

func (q *MetricsRepo) Percentiles(ctx context.Context, start, end time.Time, bucket time.Duration) ([]port.MetricPercentiles, error) {
	var rows []struct {
		Service string  `ch:"service"`
		Metric  string  `ch:"metric"`
		P50     float64 `ch:"p50"`
		P90     float64 `ch:"p90"`
		P95     float64 `ch:"p95"`
		PMax    float64 `ch:"pmax"`
	}

	excludes := []string{}
	if disableDemo() {
		excludes = append(excludes, "coingecko")
	}

	if err := q.db.Select(ctx, &rows, percentilesSQL, start, end, int(bucket.Seconds()), excludes); err != nil {
		return nil, fmt.Errorf("clickhouse.Select: %w", err)
	}

	out := make([]port.MetricPercentiles, 0, len(rows))
	for _, r := range rows {
		out = append(out, port.MetricPercentiles{
			Service: r.Service,
			Metric:  r.Metric,
			P50:     r.P50,
			P90:     r.P90,
			P95:     r.P95,
			PMax:    r.PMax,
		})
	}

	return out, nil
}

func disableDemo() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("DEMO")))
	if v == "false" || v == "0" || v == "no" || v == "off" {
		return false
	}
	return true
}
