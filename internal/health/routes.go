package health

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/magicbell/mason"
	"github.com/magicbell/mason/model"
)

// SetupRoutes registers the healthcheck route on the provided mason API.
func SetupRoutes(api *mason.API) {
	grp := api.NewRouteGroup("health")
	grp.Register(mason.HandleGet(HealthCheck).
		Path("/healthcheck").
		WithOpID("healthcheck").
		SkipIf(true))
}

var _ model.Entity = (*Response)(nil)

type Response struct {
	Version string `json:"version"`
}

func (r *Response) Example() []byte {
	return []byte(`{
      "version": "1.0.0"
    }`)
}

func (r *Response) Marshal() (json.RawMessage, error) {
	return json.Marshal(r)
}

func (r *Response) Name() string {
	return "HealthCheckResponse"
}

func (r *Response) Schema() []byte {
	return []byte(`{
      "type": "object",
      "properties": {
        "version": {
          "type": "string",
        }
      },
      "required": ["version"]
    }`)
}

func (r *Response) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, r)
}

// HealthCheck is the HTTP handler returning current version information.
func HealthCheck(ctx context.Context, r *http.Request, _ model.Nil) (rsp *Response, err error) {
	return &Response{Version: GetVersion()}, nil
}
