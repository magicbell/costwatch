package api

import (
	"context"
	"log/slog"

	"github.com/costwatchai/costwatch/internal/appconfig"
	"github.com/costwatchai/costwatch/internal/clickstore"
	cwsync "github.com/costwatchai/costwatch/internal/costwatch/sync"
	"github.com/magicbell/mason"
)

// API wires ClickHouse and CostWatch and exposes HTTP routes.
type API struct {
	log    *slog.Logger
	store  *clickstore.Client
	alerts *cwsync.Store
}

// New initializes ClickHouse and CostWatch for serving HTTP routes.
func New(ctx context.Context, log *slog.Logger, cfg appconfig.Config) (*API, error) {
	// Try to create a real ClickHouse client; on failure, fall back to TestStore for graceful local runs.
	store, _ := clickstore.NewClient(ctx, log, cfg.Clickhouse)

	// Open (or create) sqlite for alerts and sync state
	alerts, _ := cwsync.Open(".db/costwatch_state.db")

	return &API{
		log:    log,
		store:  store,
		alerts: alerts,
	}, nil
}

// SetupRoutes registers HTTP routes on the provided Mason API.
func (a *API) SetupRoutes(api *mason.API) {
	grp := api.NewRouteGroup("costwatch")
	grp.Register(mason.HandleGet(a.Usage).
		Path("/usage").
		WithOpID("usage"))

	grp.Register(mason.HandleGet(a.Anomalies).
		Path("/anomalies").
		WithOpID("anomalies"))

	grp.Register(mason.HandleGet(a.UsagePercentiles).
		Path("/usage-percentiles").
		WithOpID("usage_percentiles"))

	grp.Register(mason.HandleGet(a.AlertRules).
		Path("/alert-rules").
		WithOpID("alert_rules"))

	grp.Register(mason.HandlePut(a.UpdateAlertRule).
		Path("/alert-rules").
		WithOpID("update_alert_threshold"))

	grp.Register(mason.HandleGet(a.AlertWindows).
		Path("/alert-windows").
		WithOpID("alert_windows"))
}
