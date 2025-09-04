package api

import (
	"context"
	"log/slog"

	"github.com/costwatchai/costwatch/apps/api/handler"
	"github.com/costwatchai/costwatch/internal/appconfig"
	ca "github.com/costwatchai/costwatch/internal/costwatch/api"
	"github.com/costwatchai/costwatch/internal/monolith"
	"github.com/magicbell/mason"
)

var _ monolith.Monolith = (*Monolith)(nil)

type Monolith struct {
	*Server
	api *mason.API
	cfg appconfig.Config
	log *slog.Logger
}

func NewMonolith(log *slog.Logger, cfg appconfig.Config) *Monolith {
	// ===========================================================================
	// New Server
	srv := NewServer(log)

	// ===========================================================================
	// Mason API
	api := mason.NewAPI(srv)

	// healthcheck
	g := api.NewRouteGroup("health")
	g.Register(mason.HandleGet(handler.HealthCheck).
		Path("/healthcheck").
		WithOpID("healthcheck").
		SkipIf(true))

	// CostWatch API routes
	cw, _ := ca.New(context.Background(), log, cfg)
	cw.SetupRoutes(api)

	// ===========================================================================
	spec := handler.NewSpecFile(api)
	spec.SetupRoutes()

	return &Monolith{
		Server: srv,
		api:    api,
		cfg:    cfg,
		log:    log,
	}
}

// API implements monolith.Monolith.
func (m *Monolith) API() *mason.API {
	return m.api
}

// Config implements monolith.Monolith.
func (m *Monolith) Config() appconfig.Config {
	return m.cfg
}

// Logger implements monolith.Monolith.
func (m *Monolith) Logger(name string) *slog.Logger {
	return m.log.WithGroup(name)
}
