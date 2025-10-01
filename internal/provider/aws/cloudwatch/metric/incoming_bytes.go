package metric

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/tailbits/costwatch/internal/costwatch"
)

const (
	IncomingBytesPrice         = 50
	IncomingBytesUnitsPerPrice = 1e9
)

var _ costwatch.Metric = (*IncomingBytes)(nil)

type IncomingBytes struct {
	log    *slog.Logger
	client *cloudwatch.Client
}

func NewIncomingBytes(log *slog.Logger, client *cloudwatch.Client) *IncomingBytes {
	return &IncomingBytes{
		log:    log,
		client: client,
	}
}

// UnitsPerPrice implements costwatch.Metric.
func (m *IncomingBytes) UnitsPerPrice() float64 {
	return IncomingBytesUnitsPerPrice
}

func (m *IncomingBytes) Datapoints(ctx context.Context, label string, start time.Time, end time.Time) ([]costwatch.Datapoint, error) {
	data, err := m.client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
		MetricName: aws.String(m.Label()),
		Namespace:  aws.String("AWS/Logs"),
		Period:     aws.Int32(900),
		StartTime:  &start,
		EndTime:    &end,
		Statistics: []types.Statistic{types.StatisticSum},
	})
	if err != nil {
		return nil, fmt.Errorf("cloudwatch.GetMetricStatistics: %w", err)
	}
	m.log.Debug("fetched metric data", "points", len(data.Datapoints))

	points := make([]costwatch.Datapoint, 0, len(data.Datapoints))
	for _, result := range data.Datapoints {
		points = append(points, costwatch.Datapoint{
			Timestamp: *result.Timestamp,
			Value:     *result.Sum,
		})
	}

	return points, nil
}

func (m *IncomingBytes) Label() string {
	return "IncomingBytes"
}

func (m *IncomingBytes) Price() float64 {
	return IncomingBytesPrice
}
