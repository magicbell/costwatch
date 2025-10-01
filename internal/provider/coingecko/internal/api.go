package cgapi

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/tailbits/costwatch/internal/costwatch"
)

// MarketChart represents the subset of CoinGecko market_chart response we care about.
type MarketChart struct {
	Prices [][]float64 `json:"prices"`
}

// DefaultHTTPClient returns a client with sane defaults if nil is supplied.
func DefaultHTTPClient(client *http.Client) *http.Client {
	if client != nil {
		return client
	}
	return &http.Client{Timeout: 10 * time.Second}
}

// FetchMarketChart retrieves the market_chart data for the given vsCurrency over the [start,end] window.
// It computes the fractional days parameter (rounded to 2 decimals) as required by CoinGecko.
func FetchMarketChart(ctx context.Context, client *http.Client, vsCurrency string, start, end time.Time) (out MarketChart, err error) {
	client = DefaultHTTPClient(client)

	days := math.Ceil(end.Sub(start).Hours() / 24.0)
	if days <= 0 {
		return out, nil
	}

	base := "https://api.coingecko.com/api/v3/coins/bitcoin/market_chart"
	q := url.Values{}
	q.Set("vs_currency", vsCurrency)
	q.Set("days", fmt.Sprintf("%d", int(days)))
	u := base + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return out, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return out, fmt.Errorf("coingecko request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return out, fmt.Errorf("coingecko response status: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return out, fmt.Errorf("decode response: %w", err)
	}
	return out, nil
}

// PricesToDatapoints converts market chart price pairs to datapoints using the provided
// bucket duration for timestamp truncation and filtering to the (start, end) window.
// valueFn allows transforming the numeric price into a demo-friendly value.
func PricesToDatapoints(mc MarketChart, start, end time.Time, bucket time.Duration) []costwatch.Datapoint {
	if bucket <= 0 {
		bucket = time.Hour
	}

	points := make([]costwatch.Datapoint, 0, len(mc.Prices))
	for _, pair := range mc.Prices {
		if len(pair) != 2 {
			continue
		}
		tsMillis := int64(pair[0])
		val := pair[1]
		ts := time.UnixMilli(tsMillis).UTC().Truncate(bucket)
		if ts.After(start) && ts.Before(end) {
			found := false
			for _, p := range points {
				if p.Timestamp.Equal(ts) {
					found = true
					break
				}
			}
			if !found {
				points = append(points, costwatch.Datapoint{Timestamp: ts, Value: val})
			}
		}
	}
	return points
}
