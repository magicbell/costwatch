package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/magicbell/mason/model"
)

type PercentilesResponse QueryResult[PercentileRecord]

var _ model.Entity = (*PercentilesResponse)(nil)

//go:embed schemas/usage_percentiles_response.schema.json
var percentilesResponseSchema []byte

//go:embed schemas/usage_percentiles_response.example.json
var percentilesResponseExample []byte

func (r *PercentilesResponse) Name() string { return "UsagePercentilesResponse" }
func (r *PercentilesResponse) Schema() []byte {
	return percentilesResponseSchema
}
func (r *PercentilesResponse) Example() []byte                   { return percentilesResponseExample }
func (r *PercentilesResponse) Marshal() (json.RawMessage, error) { return json.Marshal(r) }
func (r *PercentilesResponse) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, r)
}

func (a *API) Percentiles(ctx context.Context, _ *http.Request, _ model.Nil) (res *PercentilesResponse, err error) {
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

	res = &PercentilesResponse{
		Items:    items,
		FromDate: start,
		ToDate:   end,
		Interval: 0,
	}

	return res, nil
}
