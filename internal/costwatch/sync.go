package costwatch

import (
	"context"
	"fmt"
	"time"
)

// Sync performs a full sync across services/metrics using the sync-state to compute windows.
// Rules:
// - End is always "now" (UTC).
// - Start is the oldest (earliest) of [last-synced, now-15m] to ensure we fetch at least 15 minutes.
// - If no last-synced exists, default to a first-time lookback (currently 48 hours) to backfill history.
func (cw *CostWatch) Sync(ctx context.Context) error {
	st, err := cw.getSyncStore()
	if err != nil {
		return fmt.Errorf("open syncstate: %w", err)
	}
	now := time.Now().UTC()
	fifteenAgo := now.Add(-15 * time.Minute)
	for _, s := range ListServices() {
		for _, m := range s.Metrics() {
			last, ok, err := st.GetLastSync(ctx, s.Label(), m.Label())
			if err != nil {
				cw.log.Error("syncstate.GetLastSync error", "service", s.Label(), "metric", m.Label(), "error", err)
				continue
			}
			start := last
			if !ok || last.IsZero() {
				start = now.Add(-48 * time.Hour)
			} else if last.After(fifteenAgo) {
				start = fifteenAgo
			}
			end := now
			if !start.Before(end) {
				continue
			}
			cw.log.Info("fetching metric", "service", s.Label(), "metric", m.Label(), "start", start, "end", end)
			if err := cw.FetchMetricForService(ctx, s, m, start, end); err != nil {
				cw.log.Error("FetchMetricForService failed", "service", s.Label(), "metric", m.Label(), "error", err)
				continue
			}
			if err := st.SetLastSync(ctx, s.Label(), m.Label(), end); err != nil {
				cw.log.Error("syncstate.SetLastSync error", "service", s.Label(), "metric", m.Label(), "error", err)
			}
		}
	}
	// After successful sync, attempt to send alerts.
	if err := cw.sendAlerts(ctx); err != nil {
		cw.log.Error("sendAlerts failed", "error", err)
	}
	return nil
}
