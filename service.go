package costwatch

import (
	"context"
	"time"
)

type Metric interface {
	Label() string
	Price() float64
	UnitsPerPrice() float64
	Datapoints(ctx context.Context, label string, start time.Time, end time.Time) ([]Datapoint, error)
}

type Datapoint struct {
	Value     float64
	Timestamp time.Time
}

type Service interface {
	Label() string
	Metrics() []Metric
	NewMetric(mtr Metric)
}
