package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/costwatchai/costwatch/internal/costwatch/port"
	"github.com/magicbell/mason/model"
)

var _ model.Entity = (*ListResult[AlertRule])(nil)

func (r *ListResult[AlertRule]) Example() []byte {
	return []byte(`{}`)
}

func (r *ListResult[AlertRule]) Marshal() (json.RawMessage, error) { return json.Marshal(r) }
func (r *ListResult[AlertRule]) Name() string                      { return "UsageResponse" }
func (r *ListResult[AlertRule]) Schema() []byte {
	return []byte(`{}`)
}
func (r *ListResult[AlertRule]) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, r)
}

// UpdateAlertThresholdRequest is the payload to upsert an alert threshold.
type UpdateAlertThresholdRequest struct {
	Service   string  `json:"service"`
	Metric    string  `json:"metric"`
	Threshold float64 `json:"threshold"`
}

// UpdateAlertRule upserts a threshold keyed by service+metric.
func (a *API) UpdateAlertRule(ctx context.Context, _ *http.Request, ent *AlertRule, _ model.Nil) (res *AlertRule, err error) {
	if err := a.alerts.Alerts.UpsertRule(ctx, port.AlertRule{Service: ent.Service, Metric: ent.Metric, Threshold: ent.Threshold}); err != nil {
		return nil, fmt.Errorf("rules.Upsert: %w", err)
	}
	return ent, nil
}

func (a *API) AlertRules(ctx context.Context, _ *http.Request, _ model.Nil) (res *ListResult[AlertRule], err error) {
	recs, err := a.alerts.Alerts.ListRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("rules.List: %w", err)
	}
	items := make([]AlertRule, 0, len(recs))
	for _, rec := range recs {
		items = append(items, AlertRule{Service: rec.Service, Metric: rec.Metric, Threshold: rec.Threshold})
	}
	res = &ListResult[AlertRule]{Items: items}
	return res, nil
}

func (a *API) computeAlertWindows(ctx context.Context, start, end time.Time, interval int) ([]AlertWindow, error) {
	wins, err := a.alerts.ComputeWindows(ctx, start, end, time.Duration(interval)*time.Second)
	if err != nil {
		return nil, err
	}
	res := make([]AlertWindow, 0, len(wins))
	for _, w := range wins {
		// If this window ends in the last bucket, expose end as null to indicate an ongoing anomaly.
		// Determine last bucket start by truncating the API query end to the interval duration.
		var endPtr *time.Time
		bucket := time.Duration(interval) * time.Second
		lastBucketStart := end.Truncate(bucket)
		if w.End.After(lastBucketStart) {
			endPtr = nil
		} else {
			endCopy := w.End // create a copy to take address safely
			endPtr = &endCopy
		}
		res = append(res, AlertWindow{
			Service:      w.Service,
			Metric:       w.Metric,
			Start:        w.Start,
			End:          endPtr,
			ExpectedCost: w.Threshold * float64(w.Hours),
			RealCost:     w.RealCost,
		})
	}
	return res, nil
}

// AlertWindows returns contiguous windows where hourly cost exceeded thresholds.
func (a *API) AlertWindows(ctx context.Context, _ *http.Request, _ model.Nil) (res *QueryResult[AlertWindow], err error) {
	end := time.Now().UTC()
	start := end.Add(-28 * 24 * time.Hour)
	interval := 3600

	windows, err := a.computeAlertWindows(ctx, start, end, interval)
	if err != nil {
		return nil, err
	}

	res = &QueryResult[AlertWindow]{
		Items:    windows,
		FromDate: start,
		ToDate:   end,
		Interval: interval,
	}
	return res, nil
}
