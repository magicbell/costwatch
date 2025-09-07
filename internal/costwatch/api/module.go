package api

import (
	"context"
	"log/slog"

	"github.com/costwatchai/costwatch/internal/clickstore"
	appsvc "github.com/costwatchai/costwatch/internal/costwatch/app"
	ctlinfra "github.com/costwatchai/costwatch/internal/costwatch/infra/catalog"
	chinfra "github.com/costwatchai/costwatch/internal/costwatch/infra/clickhouse"
	sqlinfra "github.com/costwatchai/costwatch/internal/costwatch/infra/sqlite"
	"github.com/costwatchai/costwatch/internal/sqlstore"
	"github.com/magicbell/mason"
)

// API wires ClickHouse and CostWatch and exposes HTTP routes.
type API struct {
	log    *slog.Logger
	alerts *appsvc.AlertService
	usage  *appsvc.UsageService
}

// New constructs the API with a pre-initialized ClickHouse client.
func New(_ context.Context, log *slog.Logger, store *clickstore.Client) (*API, error) {
	alertsDB, _ := sqlstore.Open()
	m := chinfra.NewMetricsRepo(store)
	a := sqlinfra.NewAlertsRepos(alertsDB)
	c := ctlinfra.GlobalRegistryCatalog{}
	var n nilNotifier // API does not send alerts
	alerts := appsvc.NewAlertService(m, a, n, c)
	usage := appsvc.NewUsageService(m, c)

	return &API{
		log:    log,
		alerts: alerts,
		usage:  usage,
	}, nil
}

type nilNotifier struct{}

func (nilNotifier) Send(ctx context.Context, text string) error { return nil }

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
		WithOpID("update_alert_rule"))

	grp.Register(mason.HandleGet(a.AlertWindows).
		Path("/alert-windows").
		WithOpID("alert_windows"))
}
