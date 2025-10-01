package app

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/tailbits/costwatch/internal/costwatch/port"
)

type AlertService struct {
	Metrics port.MetricsRepo
	Alerts  port.AlertsRepo
	Notify  port.Notifier // optional for SendAlerts
	Catalog port.Catalog
}

type AlertWindow struct {
	Service   string
	Metric    string
	Start     time.Time
	End       time.Time
	Hours     int
	RealCost  float64
	Threshold float64
}

func NewAlertService(metrics port.MetricsRepo, alerts port.AlertsRepo, notifier port.Notifier, catalog port.Catalog) *AlertService {
	return &AlertService{Metrics: metrics, Alerts: alerts, Notify: notifier, Catalog: catalog}
}

// ComputeWindows aggregates usage into buckets and returns contiguous windows
// where the computed hourly cost exceeded configured thresholds.
func (s *AlertService) ComputeWindows(ctx context.Context, start, end time.Time, bucket time.Duration) ([]AlertWindow, error) {
	rules, err := s.Alerts.ListRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("rules.List: %w", err)
	}
	if len(rules) == 0 {
		return nil, nil
	}
	thr := make(map[string]float64, len(rules))
	for _, r := range rules {
		thr[r.Service+"\x00"+r.Metric] = r.Threshold
	}

	recs, err := s.Metrics.Aggregate(ctx, start, end, bucket)
	if err != nil {
		return nil, fmt.Errorf("q.Aggregate: %w", err)
	}

	type curWin struct {
		start time.Time
		end   time.Time
		hours int
		real  float64
		last  time.Time
		thr   float64
	}

	var (
		windows    = make([]AlertWindow, 0)
		curService string
		curMetric  string
		cur        *curWin
		curKey     string
	)

	flush := func() {
		if cur == nil || cur.hours == 0 {
			return
		}
		windows = append(windows, AlertWindow{
			Service:   curService,
			Metric:    curMetric,
			Start:     cur.start,
			End:       cur.end,
			Hours:     cur.hours,
			RealCost:  cur.real,
			Threshold: cur.thr,
		})
		cur = nil
	}

	for _, r := range recs {
		key := r.Service + "\x00" + r.Metric
		thrVal, ok := thr[key]
		if !ok {
			if cur != nil && key != curKey {
				flush()
			}
			cur = nil
			curKey = key
			curService, curMetric = r.Service, r.Metric
			continue
		}

		cost, _ := s.Catalog.ComputeCost(r.Service, r.Metric, r.Units)

		if key != curKey {
			flush()
			curKey = key
			curService, curMetric = r.Service, r.Metric
		}

		if cost > thrVal {
			if cur == nil {
				cur = &curWin{start: r.Timestamp, end: r.Timestamp.Add(bucket), hours: 1, real: cost, last: r.Timestamp, thr: thrVal}
			} else if r.Timestamp.Sub(cur.last) == bucket {
				cur.end = r.Timestamp.Add(bucket)
				cur.hours++
				cur.real += cost
				cur.last = r.Timestamp
			} else {
				flush()
				cur = &curWin{start: r.Timestamp, end: r.Timestamp.Add(bucket), hours: 1, real: cost, last: r.Timestamp, thr: thrVal}
			}
		} else if cur != nil {
			flush()
		}
	}
	flush()

	sort.Slice(windows, func(i, j int) bool { return windows[i].Start.After(windows[j].Start) })
	return windows, nil
}

// SendAlerts computes windows over a lookback and sends notifications using the injected Notifier.
func (s *AlertService) SendAlerts(ctx context.Context) error {
	if s.Notify == nil || s.Alerts == nil {
		return nil // nothing to do if not wired
	}
	now := time.Now().UTC()
	start := now.Add(-48 * time.Hour)
	end := now // use precise now to allow detecting ongoing windows in the current bucket
	bucket := time.Hour

	wins, err := s.ComputeWindows(ctx, start, end, bucket)
	if err != nil {
		return err
	}
	if len(wins) == 0 {
		return nil
	}
	recentCutoff := now.Add(-2 * time.Hour)
	lastBucketStart := now.Truncate(bucket)
	for _, w := range wins {
		if !w.End.After(recentCutoff) {
			continue
		}
		lastUnix, ok, err := s.Alerts.GetLastNotified(ctx, w.Service, w.Metric)
		if err != nil {
			continue
		}
		if ok {
			last := time.Unix(lastUnix, 0).UTC()
			if now.Sub(last) < time.Hour {
				continue
			}
		}
		expected := w.Threshold * float64(w.Hours)
		ongoing := w.End.After(lastBucketStart)
		var text string
		if ongoing {
			text = fmt.Sprintf("[CostWatch] Alert: %s/%s exceeded threshold for %dh (expected $%.2f, actual $%.2f) since %s UTC (ongoing)", w.Service, w.Metric, w.Hours, expected, w.RealCost, w.Start.Format(time.RFC3339))
		} else {
			text = fmt.Sprintf("[CostWatch] Alert: %s/%s exceeded threshold for %dh (expected $%.2f, actual $%.2f) from %s to %s UTC", w.Service, w.Metric, w.Hours, expected, w.RealCost, w.Start.Format(time.RFC3339), w.End.Format(time.RFC3339))
		}
		if err := s.Notify.Send(ctx, text); err != nil {
			continue
		}
		_ = s.Alerts.SetLastNotified(ctx, w.Service, w.Metric, now.Unix())
	}
	return nil
}
