package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/magicbell/mason/model"
)

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
	return "PingResponse"
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

func PingHandler(ctx context.Context, r *http.Request, _ model.Nil) (rsp *Response, err error) {
	ver, ok := os.LookupEnv("APP_VERSION")
	if !ok {
		ver = time.Now().String()
	}

	return &Response{
		Version: ver,
	}, nil
}
