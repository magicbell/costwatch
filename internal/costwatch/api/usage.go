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

func (m *API) SelectAnomalies(ctx context.Context, start time.Time, end time.Time, interval int) (res []costwatch.AnomalyRecord, err error) {
	qry := `
		WITH
			toIntervalSecond(?) AS bucket,
			toDateTime(?)       AS start_ts,
			toDateTime(?)       AS end_ts,
			toStartOfInterval(end_ts, bucket) AS last_complete_bucket,
			20 AS window,
			3.0 AS threshold,
			agg AS (
				SELECT
					service,
					metric,
					toStartOfInterval(timestamp, bucket) AS timestamp,
					sum(value) AS value
				FROM metrics FINAL
				GROUP BY service, metric, timestamp
			),
			d AS (
				SELECT
					service,
					metric,
					timestamp,
					value AS sum,
					value - lagInFrame(value, 1) OVER (PARTITION BY service, metric ORDER BY timestamp) AS diff
				FROM agg
			),
			rs AS (
				SELECT
					service AS s2,
					metric  AS m2,
					timestamp AS t2,
					avg(diff)       OVER w AS mu,
					stddevPop(diff) OVER w AS sigma
				FROM (
					SELECT
						service,
						metric,
						timestamp,
						value
						  - lagInFrame(value, 1) OVER (PARTITION BY service, metric ORDER BY timestamp)
						  AS diff
					FROM agg
				)
				WINDOW w AS (PARTITION BY service, metric ORDER BY timestamp ROWS BETWEEN window PRECEDING AND 1 PRECEDING)
			)
		SELECT
			d.service,
			d.metric,
			d.timestamp,
			d.sum,
			d.diff,
			(d.diff - rs.mu) / nullIf(rs.sigma, 0) AS z_score
		FROM d
		LEFT JOIN rs
			ON rs.s2 = d.service AND rs.m2 = d.metric AND rs.t2 = d.timestamp
		WHERE
			d.diff IS NOT NULL
			AND rs.sigma IS NOT NULL
			AND abs((d.diff - rs.mu) / nullIf(rs.sigma, 0)) > threshold
			AND d.timestamp >= start_ts
			AND d.timestamp <  last_complete_bucket
		ORDER BY d.timestamp desc;
	`

	if err = m.store.Select(ctx, &res, qry, interval, start, end); err != nil {
		err = fmt.Errorf("click.Select: %w", err)
		return
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

func (a *API) Anomalies(ctx context.Context, _ *http.Request, _ model.Nil) (res *QueryResult[AnomalyRecord], err error) {
	end := time.Now().UTC()
	start := end.Add(-28 * 24 * time.Hour)
	interval := 3600

	recs, err := a.SelectAnomalies(ctx, start, end, interval)
	if err != nil {
		return res, fmt.Errorf("SelectAnomalies: %w", err)
	}

	items := make([]AnomalyRecord, 0, len(recs))
	for _, rec := range recs {
		cost, _ := costwatch.ComputeCost(rec.Service, rec.Metric, rec.Sum)
		items = append(items, AnomalyRecord{
			Service:   rec.Service,
			Metric:    rec.Metric,
			Timestamp: rec.Timestamp,
			Sum:       rec.Sum,
			Diff:      rec.Diff,
			ZScore:    rec.ZScore,
			Cost:      cost,
		})
	}

	res = &QueryResult[AnomalyRecord]{
		Items:    items,
		FromDate: start,
		ToDate:   end,
		Interval: interval,
	}

	return
}
