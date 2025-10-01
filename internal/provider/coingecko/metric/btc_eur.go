package metric

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/tailbits/costwatch/internal/costwatch"
	cgapi "github.com/tailbits/costwatch/internal/provider/coingecko/internal"
)

const (
	IncomingBytesPrice = 1 * 100 // map to cents
	// scale it down for demo purpose
	IncomingBytesUnitsPerPrice = 1
)

// Ensure PriceMetric implements costwatch.Metric
var _ costwatch.Metric = (*PriceMetric)(nil)

type PriceMetric struct {
	log    *slog.Logger
	client *http.Client
}

func NewPriceMetric(log *slog.Logger, client *http.Client) *PriceMetric {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &PriceMetric{log: log, client: client}
}

func (p *PriceMetric) Label() string { return "btc_usd" }

// Price is set to 1 so reported cost equals the units (price points) for demo purposes.
func (p *PriceMetric) Price() float64 { return IncomingBytesPrice }

// UnitsPerPrice kept at 1 to avoid scaling in demo.
func (p *PriceMetric) UnitsPerPrice() float64 { return IncomingBytesUnitsPerPrice }

func (p *PriceMetric) Datapoints(ctx context.Context, label string, start time.Time, end time.Time) ([]costwatch.Datapoint, error) {
	if !start.Before(end) {
		return nil, nil
	}
	mc, err := cgapi.FetchMarketChart(ctx, p.client, "eur", start, end)
	if err != nil {
		return nil, err
	}
	// Transform price into demo value consistent with previous behavior.
	points := cgapi.PricesToDatapoints(mc, start, end, time.Hour)
	return points, nil
}
