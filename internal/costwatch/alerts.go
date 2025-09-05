package costwatch

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/costwatchai/costwatch/internal/clickstore"
	cwsync "github.com/costwatchai/costwatch/internal/costwatch/sync"
)

// AlertWindow is a shared representation of an alert window computed from metrics
// exceeding a configured per-hour threshold.
// Hours indicates the number of contiguous buckets in the window.
// RealCost is the sum of costs over the window; Threshold is the per-hour threshold.
// ExpectedCost can be derived as Threshold * float64(Hours).
type AlertWindow struct {
	Service   string
	Metric    string
	Start     time.Time
	End       time.Time
	Hours     int
	RealCost  float64
	Threshold float64
}

// ComputeAlertWindows aggregates usage into buckets and returns contiguous windows
// where the cost exceeded configured thresholds. This is the shared implementation
// used by both the API and the worker (sendAlerts).
func ComputeAlertWindows(ctx context.Context, ch *clickstore.Client, ruleStore *cwsync.Store, start, end time.Time, interval int) ([]AlertWindow, error) {
	// Load thresholds
	rules, err := ruleStore.GetAlertRules(ctx)
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

	// Fetch aggregated metrics
	qry := `
		SELECT service, metric, toStartOfInterval(timestamp, toIntervalSecond(?)) AS timestamp, sum(value) AS value
		FROM metrics FINAL
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY service, metric, timestamp
		ORDER BY service, metric, timestamp;
	`
	var recs []MetricRecord
	if err := ch.Select(ctx, &recs, qry, interval, start, end); err != nil {
		return nil, fmt.Errorf("click.Select: %w", err)
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
		bucket     = time.Duration(interval) * time.Second
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
		thr, ok := th[key]
		if !ok {
			if cur != nil && key != curKey {
				flush()
			}
			cur = nil
			curKey = key
			curService, curMetric = r.Service, r.Metric
			continue
		}

		cost, _ := ComputeCost(r.Service, r.Metric, r.Value)

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
				flush()
				cur = &curWin{start: r.Timestamp, end: r.Timestamp.Add(bucket), hours: 1, real: cost, last: r.Timestamp, thr: thr}
			}
		} else {
			if cur != nil {
				flush()
			}
		}
	}
	flush()

	// For UI friendliness
	sort.Slice(windows, func(i, j int) bool { return windows[i].Start.After(windows[j].Start) })

	return windows, nil
}
