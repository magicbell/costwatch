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

type UsageResponse QueryResult[Record]

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

	recs, err := a.usage.Usage(ctx, start, end, time.Duration(interval)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("usage.Usage: %w", err)
	}

	items := make([]Record, 0, len(recs))
	for _, it := range recs {
		items = append(items, Record{Service: it.Service, Metric: it.Metric, Cost: it.Cost, Timestamp: it.Timestamp})
	}

	res = &UsageResponse{
		Items:    items,
		FromDate: start,
		ToDate:   end,
		Interval: interval,
	}
	return res, nil
}
