package costwatch

import "time"

type MetricRecord struct {
	Service   string    `ch:"service"`
	Metric    string    `ch:"metric"`
	Value     float64   `ch:"value"`
	Timestamp time.Time `ch:"timestamp"`
}

type PercentileRecord struct {
	Service string  `ch:"service"`
	Metric  string  `ch:"metric"`
	P50     float64 `ch:"p50"`
	P90     float64 `ch:"p90"`
	P95     float64 `ch:"p95"`
	PMax    float64 `ch:"pmax"`
}
