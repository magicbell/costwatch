package api

import (
	"context"
	"log/slog"

	"github.com/costwatchai/costwatch/internal/clickstore"
	"github.com/costwatchai/costwatch/internal/sqlstore"
	"github.com/magicbell/mason"
)

// API wires ClickHouse and CostWatch and exposes HTTP routes.
type API struct {
	log   *slog.Logger
	store *clickstore.Client
	db    *sqlstore.Store
}

// New constructs the API with a pre-initialized ClickHouse client.
func New(_ context.Context, log *slog.Logger, store *clickstore.Client) (*API, error) {
	alerts, _ := sqlstore.Open()

	return &API{
		log:   log,
		store: store,
		db:    alerts,
	}, nil
}

// SetupRoutes registers HTTP routes on the provided Mason API.
func (a *API) SetupRoutes(api *mason.API) {
	grp := api.NewRouteGroup("costwatch")
	grp.Register(mason.HandleGet(a.Usage).
		Path("/usage").
		WithOpID("usage"))

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
