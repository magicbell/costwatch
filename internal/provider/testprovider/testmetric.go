package testprovider

import (
	"context"
	_ "embed"
	"encoding/json"
	"sort"
	"testing"
	"time"

	"github.com/costwatchai/costwatch/internal/costwatch"
)

//go:embed testdata/testmetric.json
var testDataPoints []byte

type sortedDatapoints []costwatch.Datapoint

func (s sortedDatapoints) Len() int {
	return len(s)
}

func (s sortedDatapoints) Less(i, j int) bool {
	return s[i].Timestamp.Before(s[j].Timestamp)
}

func (s sortedDatapoints) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type testData struct {
	Label         string                `json:"Label"`
	Datapoints    []costwatch.Datapoint `json:"Datapoints"`
	Price         float64               `json:"Price"`
	UnitsPerPrice float64               `json:"UnitsPerPrice"`
}

type TestMetric struct {
	data testData
}

func NewTestMetric(t *testing.T) *TestMetric {
	var d testData
	if err := json.Unmarshal(testDataPoints, &d); err != nil {
		t.Fatal(err)
	}

	sort.Sort(sortedDatapoints(d.Datapoints))

	return &TestMetric{
		data: d,
	}
}

func (m *TestMetric) Datapoints(ctx context.Context, label string, start time.Time, end time.Time) ([]costwatch.Datapoint, error) {
	points := make([]costwatch.Datapoint, 0, len(m.data.Datapoints))
	for _, p := range m.data.Datapoints {
		if p.Timestamp.After(start) && p.Timestamp.Before(end) {
			points = append(points, p)
		}
	}

	return points, nil
}

// Label implements costwatch.IncomingBytes.
func (m *TestMetric) Label() string {
	return m.data.Label
}

// Cost implements costwatch.IncomingBytes.
func (m *TestMetric) Price() float64 {
	return m.data.Price
}

func (m *TestMetric) UnitsPerPrice() float64 {
	return m.data.UnitsPerPrice
}
