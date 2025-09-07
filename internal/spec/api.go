package spec

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/smithy-go/ptr"
	"github.com/costwatchai/costwatch/internal/web"
	"github.com/magicbell/mason"
	"github.com/magicbell/mason/openapi"
	"github.com/swaggest/openapi-go/openapi31"
)

// SetupRoutes registers the GET /openapi.json endpoint on the provided mason API.
func SetupRoutes(api *mason.API) {
	api.Handle(http.MethodGet, "/openapi.json", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
