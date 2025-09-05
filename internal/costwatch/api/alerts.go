package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/costwatchai/costwatch/internal/costwatch"
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

func (a *API) computeAlertWindows(ctx context.Context, start, end time.Time, interval int) ([]AlertWindow, error) {
	// Load thresholds from sqlite
	rules, err := a.alerts.GetAlertRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAlertRules: %w", err)
	}
	if len(rules) == 0 {
		return nil, nil
	}
	th := make(map[string]float64, len(rules))
	for _, r := range rules {
		key := r.Service + "\x00" + r.Metric
		th[key] = r.Threshold
	}

	// Query hourly usage from ClickHouse
	recs, err := a.SelectMetrics(ctx, start, end, interval)
	if err != nil {
		return nil, fmt.Errorf("SelectMetrics: %w", err)
	}

	type curWin struct {
		start time.Time
		end   time.Time
		hours int
		real  float64
		last  time.Time
		thr   float64
	}

	windows := make([]AlertWindow, 0)
	// Iterate sorted by service, metric, timestamp (SelectMetrics orders by that)
	var curService, curMetric string
	var cur *curWin
	var curKey string

	flush := func() {
		if cur == nil || cur.hours == 0 {
			return
		}
		expected := cur.thr * float64(cur.hours)

		windows = append(windows, AlertWindow{
			Service:      curService,
			Metric:       curMetric,
			Start:        cur.start,
			End:          cur.end,
			ExpectedCost: expected,
			RealCost:     cur.real,
		})
		cur = nil
	}

	bucket := time.Duration(interval) * time.Second
	for _, r := range recs {
		key := r.Service + "\x00" + r.Metric
		thr, ok := th[key]
		if !ok {
			// We only consider series that have thresholds configured
			if cur != nil && key != curKey {
				flush()
			}
			cur = nil
			curKey = key
			curService, curMetric = r.Service, r.Metric
			continue
		}

		cost, _ := costwatch.ComputeCost(r.Service, r.Metric, r.Value)

		// Change of series: flush previous window
		if key != curKey {
			flush()
			curKey = key
			curService, curMetric = r.Service, r.Metric
		}

		if cost > thr {
			if cur == nil {
				cur = &curWin{start: r.Timestamp, end: r.Timestamp.Add(bucket), hours: 1, real: cost, last: r.Timestamp, thr: thr}
			} else if r.Timestamp.Sub(cur.last) == bucket {
				cur.end = r.Timestamp.Add(bucket)
				cur.hours++
				cur.real += cost
				cur.last = r.Timestamp
			} else {
				// gap: close previous and start new
				flush()
				cur = &curWin{start: r.Timestamp, end: r.Timestamp.Add(bucket), hours: 1, real: cost, last: r.Timestamp, thr: thr}
			}
		} else {
			// below threshold: close if open
			if cur != nil {
				flush()
			}
		}
	}
	// flush at end
	flush()

	// Optional: sort windows by start desc for table readability
	sort.Slice(windows, func(i, j int) bool { return windows[i].Start.After(windows[j].Start) })

	return windows, nil
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
