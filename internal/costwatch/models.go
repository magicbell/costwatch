package costwatch

import "time"

type MetricRecord struct {
	Service   string    `ch:"service"`
	Metric    string    `ch:"metric"`
	Value     float64   `ch:"value"`
	Timestamp time.Time `ch:"timestamp"`
}

type AnomalyRecord struct {
	Service   string    `ch:"service"`
	Metric    string    `ch:"metric"`
	Timestamp time.Time `ch:"timestamp"`
	Sum       float64   `ch:"sum"`
	Diff      float64   `ch:"diff"`
	ZScore    float64   `ch:"z_score"`
}

type AnomalyWindow struct {
	Service   string     `ch:"service"`
	Metric    string     `ch:"metric"`
	StartTime time.Time  `ch:"start"`
	EndTime   *time.Time `ch:"end"`
	Value     float64    `ch:"sum"`
}

type PercentileRecord struct {
	Service string  `ch:"service"`
	Metric  string  `ch:"metric"`
	P50     float64 `ch:"p50"`
	P90     float64 `ch:"p90"`
	P95     float64 `ch:"p95"`
	PMax    float64 `ch:"pmax"`
}
