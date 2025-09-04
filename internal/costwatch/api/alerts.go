package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

// UpdateAlertRule upserts a threshold in sqlite keyed by service+metric.
func (a *API) UpdateAlertRule(ctx context.Context, _ *http.Request, ent *AlertRule, _ model.Nil) (res *AlertRule, err error) {
	if err := a.alerts.SetAlertThreshold(ctx, ent.Service, ent.Metric, ent.Threshold); err != nil {
		return nil, fmt.Errorf("alerts.SetAlertThreshold: %w", err)
	}

	return ent, nil
}

func (a *API) AlertRules(ctx context.Context, _ *http.Request, _ model.Nil) (res *ListResult[AlertRule], err error) {
	recs, err := a.alerts.GetAlertRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAlertRules: %w", err)
	}

	items := make([]AlertRule, 0, len(recs))
	for _, rec := range recs {
		items = append(items, AlertRule{
			Service:   rec.Service,
			Metric:    rec.Metric,
			Threshold: rec.Threshold,
		})
	}

	res = &ListResult[AlertRule]{
		Items: items,
	}
	
	return res, nil
}
