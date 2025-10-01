package api

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/magicbell/mason/model"
	"github.com/tailbits/costwatch/internal/costwatch/port"
)

type AlertRuleListResponse struct {
	Items    []AlertRule `json:"items"`
	Readonly bool        `json:"readonly"`
}

var _ model.Entity = (*AlertRuleListResponse)(nil)

//go:embed schemas/alert_rules_response.schema.json
var alertRulesResponseSchema []byte

//go:embed schemas/alert_rules_response.example.json
var alertRulesResponseExample []byte

func (r *AlertRuleListResponse) Example() []byte                   { return alertRulesResponseExample }
func (r *AlertRuleListResponse) Marshal() (json.RawMessage, error) { return json.Marshal(r) }
func (r *AlertRuleListResponse) Name() string                      { return "AlertRuleListResponse" }
func (r *AlertRuleListResponse) Schema() []byte                    { return alertRulesResponseSchema }
func (r *AlertRuleListResponse) Unmarshal(data json.RawMessage) error {
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
	if err := a.alert.Alerts.UpsertRule(ctx, port.AlertRule{Service: ent.Service, Metric: ent.Metric, Threshold: ent.Threshold}); err != nil {
		return nil, fmt.Errorf("rules.Upsert: %w", err)
	}
	return ent, nil
}

func (a *API) AlertRules(ctx context.Context, _ *http.Request, _ model.Nil) (res *AlertRuleListResponse, err error) {
	recs, err := a.alert.Alerts.ListRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("rules.List: %w", err)
	}
	items := make([]AlertRule, 0, len(recs))
	for _, rec := range recs {
		items = append(items, AlertRule{Service: rec.Service, Metric: rec.Metric, Threshold: rec.Threshold})
	}

	_, readonly := os.LookupEnv("ALERT_RULES")
	res = &AlertRuleListResponse{Items: items, Readonly: readonly}
	return res, nil
}

func (a *API) computeAlertWindows(ctx context.Context, start, end time.Time, interval int) ([]AlertWindow, error) {
	wins, err := a.alert.ComputeWindows(ctx, start, end, time.Duration(interval)*time.Second)
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

type AlertWindowsQueryResponse QueryResult[AlertWindow]

var _ model.Entity = (*AlertWindowsQueryResponse)(nil)

//go:embed schemas/alert_windows_response.schema.json
var alertWindowsResponseSchema []byte

//go:embed schemas/alert_windows_response.example.json
var alertWindowsResponseExample []byte

func (r *AlertWindowsQueryResponse) Example() []byte { return alertWindowsResponseExample }
func (r *AlertWindowsQueryResponse) Marshal() (json.RawMessage, error) {
	return json.Marshal(r)
}
func (r *AlertWindowsQueryResponse) Name() string {
	return "AlertWindowsQueryResponse"
}
func (r *AlertWindowsQueryResponse) Schema() []byte {
	return alertWindowsResponseSchema
}
func (r *AlertWindowsQueryResponse) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, r)
}

// AlertWindows returns contiguous windows where hourly cost exceeded thresholds.
func (a *API) AlertWindows(ctx context.Context, _ *http.Request, _ model.Nil) (res *AlertWindowsQueryResponse, err error) {
	end := time.Now().UTC()
	start := end.Add(-7 * 24 * time.Hour)
	interval := 3600

	windows, err := a.computeAlertWindows(ctx, start, end, interval)
	if err != nil {
		return nil, err
	}

	res = &AlertWindowsQueryResponse{
		Items:    windows,
		FromDate: start,
		ToDate:   end,
		Interval: interval,
	}
	return res, nil
}
