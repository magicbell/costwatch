package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/smithy-go/ptr"
	"github.com/costwatchai/costwatch/internal/appconfig"
	"github.com/costwatchai/costwatch/internal/monolith"
	"github.com/magicbell/magicbell/src/gofoundation/web"
	"github.com/magicbell/mason"
	"github.com/magicbell/mason/openapi"
	"github.com/swaggest/openapi-go/openapi31"
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
	g.Register(mason.HandleGet(PingHandler).
		Path("/healthcheck").
		WithOpID("healthcheck").
		SkipIf(true))

	// ===========================================================================
	srv.Handle(http.MethodGet, "/openapi.json", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		sg, err := openapi.NewGenerator(api)
		if err != nil {
			return fmt.Errorf("openapi.NewGenerator: %w", err)
		}
		addSpecInfo(sg)

		sch, err := sg.Schema()
		if err != nil {
			return fmt.Errorf("sg.Schema: %w", err)
		}

		w.Header().Set("Content-Type", "application/json")
		web.FileResponse(ctx, w, sch, http.StatusOK)
		return nil
	})

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

func addSpecInfo(gen *openapi.Generator) {
	gen.Spec.
		WithServers(openapi31.Server{
			URL:         "https://api.costwatch.ai/v1",
			Description: ptr.String("CostWatch API (v1) Base URL"),
		}).
		WithJSONSchemaDialect("http://json-schema.org/draft-07/schema#")

	gen.Spec.Info.
		WithTitle("CostWatch API").
		WithDescription("OpenAPI 3.1.0 Specification for CostWatch API.").
		WithVersion("2.0.0").
		WithContact(openapi31.Contact{
			Name: ptr.String("CostWatch"),
			URL:  ptr.String("https://costwatch.ai"),
		})
}
