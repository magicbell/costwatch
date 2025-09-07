// Package web contains a small web framework extension.
package web

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/costwatchai/costwatch/internal/webctx"
	"github.com/go-chi/chi/v5"
	"github.com/magicbell/mason"
	"github.com/magicbell/mason/model"
)

type APISchema struct {
	InputDoc  model.Entity
	OutputDoc model.Entity
	Fn        any
}

func (sch APISchema) Error() string {
	return ""
}

// Server is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this Server struct.
type Server struct {
	mux  *chi.Mux
	mw   []Middleware
	log  *slog.Logger
	host string
}

func (s *Server) SetHost(host string) {
	s.host = host
}

// NewServer creates an App value that handle a set of routes for the application.
func NewServer(log *slog.Logger, mw ...Middleware) *Server {
	mux := chi.NewMux()

	return &Server{
		mux: mux,
		mw:  mw,
		log: log,
	}
}

func (a *Server) Mux() *chi.Mux {
	return a.mux
}

// ServeHTTP implements the http.Handler interface. It's the entry point for
// all http traffic and allows the opentelemetry mux to run first to handle
// tracing. The opentelemetry mux then calls the application mux to handle
// application traffic. This was set up on line 44 in the NewApp function.
func (a *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	replaceDoubleSlash(r)
	a.mux.ServeHTTP(w, r)
}

// replaceDoubleSlash replaces a leading "//" with "/" in the request
// path. This is needed because some clients erroneously make requests
// to "//notifcations" rather than "/notifications".
func replaceDoubleSlash(r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "//") {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/")
	}
}

// NewRoute sets a handler function for a given HTTP method and path pair
// to the application server mux.
func (a *Server) NewRoute(method string, group string, path string, handler mason.WebHandler, mw ...Middleware) {
	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		v := webctx.Values{
			Now:  time.Now().UTC(),
			Host: a.host,
		}
		ctx := webctx.SetValues(r.Context(), &v)

		if err := handler(ctx, w, r); err != nil {
			// Format validation Error
			var fe model.ValidationError
			if errors.As(err, &fe) {
				// Return well-formatted validation errors
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(w).Encode(fe)
				return
			}

			a.log.Error("handler error", "error", err)
			// You might want to handle the error here, e.g., write an error response
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	a.mux.MethodFunc(method, finalPath, h)
}

func (a *Server) ServeStatic(path string, root http.FileSystem) {
	a.mux.Handle(path, http.FileServer(root))
}
