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

type AnomalyRecord struct {
	Service   string    `json:"service"`
	Metric    string    `json:"metric"`
	Timestamp time.Time `json:"timestamp"`
	Sum       float64   `json:"sum"`
	Diff      float64   `json:"diff"`
	ZScore    float64   `json:"z_score"`
	Cost      float64   `json:"cost"`
}

type AnomalyWindow struct {
	Service   string     `json:"service"`
	Metric    string     `json:"metric"`
	StartTime time.Time  `json:"start"`
	EndTime   *time.Time `json:"end"`
	Value     float64    `json:"value"`
	Cost      float64    `json:"cost"`
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
