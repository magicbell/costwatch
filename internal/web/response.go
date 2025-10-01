package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tailbits/costwatch/internal/webctx"
)

// Respond converts a Go value to JSON and sends it to the client.
func Respond(ctx context.Context, w http.ResponseWriter, data any, statusCode int) error {
	webctx.SetStatusCode(ctx, statusCode)

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encoder.Encode: %w", err)
	}

	return nil
}

func HTMLResponse(ctx context.Context, w http.ResponseWriter, data []byte, statusCode int) error {
	webctx.SetStatusCode(ctx, statusCode)

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	w.Header().Set("Content-Type", "text/html")

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write HTML response: %w", err)
	}

	return nil
}

func FileResponse(ctx context.Context, w http.ResponseWriter, data []byte, statusCode int) error {
	webctx.SetStatusCode(ctx, statusCode)

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write file response: %w", err)
	}

	return nil
}
