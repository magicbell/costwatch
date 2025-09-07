package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Param returns the path parameter with the given key.
func Param(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

// Decode reads the body of an HTTP request looking for a JSON document. The
// body is decoded into the provided value.
func Decode(r *http.Request, model any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read the body: %w", err)
	}

	if err := json.Unmarshal(body, model); err != nil {
		return fmt.Errorf("unable to unmarshal the data: %w", err)
	}

	return nil
}
