package api

import (
	"encoding/json"
	"time"
)

type QueryResult[T any] struct {
	FromDate time.Time `json:"from_date"`
	ToDate   time.Time `json:"to_date"`
	Interval int       `json:"interval"`
	Items    []T       `json:"items"`
}

type ListResult[T any] struct {
	Items []T `json:"items"`
}

type UsageRecord struct {
	Service   string    `json:"service"`
	Metric    string    `json:"metric"`
	Cost      float64   `json:"cost"`
	Timestamp time.Time `json:"timestamp"`
}

// AlertWindow represents a contiguous time window where the hourly cost exceeded
// the configured threshold for a given service/metric. ExpectedCost is the
// (threshold * hours in window). RealCost is the sum of costs in the window.
// Diff is RealCost - ExpectedCost. DiffPercent is 100 * Diff / ExpectedCost
// when ExpectedCost > 0, otherwise 0.
type AlertWindow struct {
	Service      string     `json:"service"`
	Metric       string     `json:"metric"`
	Start        time.Time  `json:"start"`
	End          *time.Time `json:"end"`
	ExpectedCost float64    `json:"expected_cost"`
	RealCost     float64    `json:"real_cost"`
}

type PercentileRecord struct {
	Service string  `json:"service"`
	Metric  string  `json:"metric"`
	P50     float64 `json:"p50"`
	P90     float64 `json:"p90"`
	P95     float64 `json:"p95"`
	PMax    float64 `json:"pmax"`
}

type AlertRule struct {
	Service   string  `json:"service"`
	Metric    string  `json:"metric"`
	Threshold float64 `json:"threshold"`
}

func (u *AlertRule) Name() string {
	return "AlertRule"
}

func (u *AlertRule) Example() []byte {
	return []byte(`{
		"service": "aws.CloudWatch",
		"metric": "IncomingBytes",
		"threshold": 5.25
	}`)
}

func (u *AlertRule) Schema() []byte {
	return []byte(`{
      "type": "object",
      "properties": {
        "service": {
          "type": "string"
        },
        "metric": {
          "type": "string"
        },
		"threshold": {
          "type": "number"
        }
      },
      "required": ["service", "metric", "threshold"]
    }`)
}

func (u *AlertRule) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, u)
}

func (u *AlertRule) Marshal() (json.RawMessage, error) {
	return json.Marshal(u)
}
