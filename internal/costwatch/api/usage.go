package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

// Usage returns service usage + cost per interval for the past 28 days.
func (a *API) Usage(ctx context.Context, _ *http.Request, _ model.Nil) (res *QueryResult[UsageRecord], err error) {
	end := time.Now().UTC()
	start := end.Add(-28 * 24 * time.Hour)
	interval := 3600

	itemsRaw, err := a.usage.Usage(ctx, start, end, time.Duration(interval)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("usage.Usage: %w", err)
	}

	items := make([]UsageRecord, 0, len(itemsRaw))
	for _, it := range itemsRaw {
		items = append(items, UsageRecord{Service: it.Service, Metric: it.Metric, Cost: it.Cost, Timestamp: it.Timestamp})
	}

	res = &QueryResult[UsageRecord]{
		Items:    items,
		FromDate: start,
		ToDate:   end,
		Interval: interval,
	}
	return res, nil
}

func (a *API) UsagePercentiles(ctx context.Context, _ *http.Request, _ model.Nil) (res *QueryResult[PercentileRecord], err error) {
	end := time.Now().UTC()
	start := end.Add(-7 * 24 * time.Hour)
	interval := 3600

	recs, err := a.usage.UsagePercentiles(ctx, start, end, time.Duration(interval)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("usage.UsagePercentiles: %w", err)
	}
	items := make([]PercentileRecord, 0, len(recs))
	for _, r := range recs {
		items = append(items, PercentileRecord{Service: r.Service, Metric: r.Metric, P50: r.P50, P90: r.P90, P95: r.P95, PMax: r.PMax})
	}
	res = &QueryResult[PercentileRecord]{
		Items:    items,
		FromDate: start,
		ToDate:   end,
		Interval: 0,
	}
	return res, nil
}
