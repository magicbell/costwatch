package api

import "time"

type QueryResult[T any] struct {
	FromDate time.Time `json:"from_date"`
	ToDate   time.Time `json:"to_date"`
	Interval int       `json:"interval"`
	Items    []T       `json:"items"`
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
