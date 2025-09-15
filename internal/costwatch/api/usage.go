package api

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/magicbell/mason/model"
)

type UsageResponse QueryResult[UsageRecord]

var _ model.Entity = (*UsageResponse)(nil)

//go:embed schemas/usage_response.schema.json
var usageResponseSchema []byte

//go:embed schemas/usage_response.example.json
var usageResponseExample []byte

func (r *UsageResponse) Name() string                      { return "UsageResponse" }
func (r *UsageResponse) Schema() []byte                    { return usageResponseSchema }
func (r *UsageResponse) Example() []byte                   { return usageResponseExample }
func (r *UsageResponse) Marshal() (json.RawMessage, error) { return json.Marshal(r) }
func (r *UsageResponse) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, r)
}

// Usage returns service usage + cost per interval for the past 7 days.
func (a *API) Usage(ctx context.Context, _ *http.Request, _ model.Nil) (res *UsageResponse, err error) {
	end := time.Now().UTC()
	start := end.Add(-7 * 24 * time.Hour)
	interval := 3600

	itemsRaw, err := a.usage.Usage(ctx, start, end, time.Duration(interval)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("usage.Usage: %w", err)
	}

	items := make([]UsageRecord, 0, len(itemsRaw))
	for _, it := range itemsRaw {
		items = append(items, UsageRecord{Service: it.Service, Metric: it.Metric, Cost: it.Cost, Timestamp: it.Timestamp})
	}

	res = &UsageResponse{
		Items:    items,
		FromDate: start,
		ToDate:   end,
		Interval: interval,
	}
	return res, nil
}

type UsagePercentilesResponse QueryResult[PercentileRecord]

var _ model.Entity = (*UsagePercentilesResponse)(nil)

//go:embed schemas/usage_percentiles_response.schema.json
var usagePercentilesResponseSchema []byte

//go:embed schemas/usage_percentiles_response.example.json
var usagePercentilesResponseExample []byte

func (r *UsagePercentilesResponse) Name() string { return "UsagePercentilesResponse" }
func (r *UsagePercentilesResponse) Schema() []byte {
	return usagePercentilesResponseSchema
}
func (r *UsagePercentilesResponse) Example() []byte                   { return usagePercentilesResponseExample }
func (r *UsagePercentilesResponse) Marshal() (json.RawMessage, error) { return json.Marshal(r) }
func (r *UsagePercentilesResponse) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, r)
}

func (a *API) UsagePercentiles(ctx context.Context, _ *http.Request, _ model.Nil) (res *UsagePercentilesResponse, err error) {
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
	res = &UsagePercentilesResponse{
		Items:    items,
		FromDate: start,
		ToDate:   end,
		Interval: 0,
	}
	return res, nil
}
