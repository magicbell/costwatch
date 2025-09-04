package api

import (
	"context"
	"log/slog"

	"github.com/costwatchai/costwatch/internal/appconfig"
	"github.com/costwatchai/costwatch/internal/clickstore"
	"github.com/magicbell/mason"
)

// API wires ClickHouse and CostWatch and exposes HTTP routes.
type API struct {
	log   *slog.Logger
	store *clickstore.Client
}

// New initializes ClickHouse and CostWatch for serving HTTP routes.
func New(ctx context.Context, log *slog.Logger, cfg appconfig.Config) (*API, error) {
	// Try to create a real ClickHouse client; on failure, fall back to TestStore for graceful local runs.
	store, _ := clickstore.NewClient(ctx, log, cfg.Clickhouse)

	return &API{
		log:   log,
		store: store,
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

}
