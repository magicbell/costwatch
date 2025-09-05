package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/costwatchai/costwatch/internal/costwatch"
	"github.com/magicbell/mason/model"
)

var _ model.Entity = (*QueryResult[UsageRecord])(nil)

func (r *QueryResult[UsageRecord]) Example() []byte {
	return []byte(`{}`)
}

func (r *QueryResult[UsageRecord]) Marshal() (json.RawMessage, error) { return json.Marshal(r) }
func (r *QueryResult[UsageRecord]) Name() string                      { return "UsageResponse" }
func (r *QueryResult[UsageRecord]) Schema() []byte {
	return []byte(`{}`)
}
func (r *QueryResult[UsageRecord]) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, r)
}

func (m *API) SelectMetrics(ctx context.Context, start time.Time, end time.Time, interval int) (res []costwatch.MetricRecord, err error) {
	qry := `
		SELECT service, metric, toStartOfInterval(timestamp, toIntervalSecond(?)) AS timestamp, sum(value) AS value
		FROM metrics FINAL
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY service, metric, timestamp
		ORDER BY service, metric, timestamp;
	`

	if err = m.store.Select(ctx, &res, qry, interval, start, end); err != nil {
		err = fmt.Errorf("click.Select: %w", err)
		return
	}

	return
}

// Usage returns service usage + cost per interval for the past 28 days.
func (a *API) Usage(ctx context.Context, _ *http.Request, _ model.Nil) (res *QueryResult[UsageRecord], err error) {
	end := time.Now().UTC()
	start := end.Add(-28 * 24 * time.Hour)
	interval := 3600

	recs, err := a.SelectMetrics(ctx, start, end, interval)
	if err != nil {
		return res, fmt.Errorf("SelectMetrics: %w", err)
	}

	items := make([]UsageRecord, 0, len(recs))
	for _, rec := range recs {
		cost, _ := costwatch.ComputeCost(rec.Service, rec.Metric, rec.Value)
		items = append(items, UsageRecord{
			Service:   rec.Service,
			Metric:    rec.Metric,
			Cost:      cost,
			Timestamp: rec.Timestamp,
		})
	}

	res = &QueryResult[UsageRecord]{
		Items:    items,
		FromDate: start,
		ToDate:   end,
		Interval: interval,
	}

	return
}

func (m *API) SelectUsagePercentiles(ctx context.Context, start time.Time, end time.Time, interval int) (res []costwatch.PercentileRecord, err error) {
	qry := `
		WITH
			? AS start_ts,
			? AS end_ts,
			toIntervalSecond(?)                AS bucket,
			toStartOfInterval(end_ts, bucket)  AS end_bucket   
		
		, by_bucket AS (
			SELECT
				service,
				metric,
				toStartOfInterval(timestamp, bucket)     AS bucket_ts,
				sum(value)                               AS bucket_usage
			FROM metrics FINAL
			WHERE timestamp >= start_ts
			  AND timestamp <  end_bucket
			GROUP BY service, metric, bucket_ts
		)
		
		SELECT
			service,
			metric,
			toFloat64(quantileTDigest(0.50)(bucket_usage)) AS p50,
			toFloat64(quantileTDigest(0.90)(bucket_usage)) AS p90,
			toFloat64(quantileTDigest(0.95)(bucket_usage)) AS p95,
			toFloat64(max(bucket_usage))                   AS pmax
		FROM by_bucket
		GROUP BY service, metric
		ORDER BY service, metric;
	`

	if err = m.store.Select(ctx, &res, qry, start, end, interval); err != nil {
		err = fmt.Errorf("click.Select: %w", err)
		return
	}
	return
}

func (a *API) UsagePercentiles(ctx context.Context, _ *http.Request, _ model.Nil) (res *QueryResult[PercentileRecord], err error) {
	end := time.Now().UTC()
	start := end.Add(-7 * 24 * time.Hour)
	interval := 3600

	recs, err := a.SelectUsagePercentiles(ctx, start, end, interval)
	if err != nil {
		return res, fmt.Errorf("SelectUsagePercentiles: %w", err)
	}

	// Convert averages (units) to costs, similar to other handlers.
	items := make([]PercentileRecord, 0, len(recs))
	for _, r := range recs {
		c50, _ := costwatch.ComputeCost(r.Service, r.Metric, r.P50)
		c90, _ := costwatch.ComputeCost(r.Service, r.Metric, r.P90)
		c95, _ := costwatch.ComputeCost(r.Service, r.Metric, r.P95)
		cMax, _ := costwatch.ComputeCost(r.Service, r.Metric, r.PMax)

		items = append(items, PercentileRecord{
			Service: r.Service,
			Metric:  r.Metric,
			P50:     c50,
			P90:     c90,
			P95:     c95,
			PMax:    cMax,
		})
	}

	res = &QueryResult[PercentileRecord]{
		Items:    items,
		FromDate: start,
		ToDate:   end,
		Interval: 0,
	}
	return
}
