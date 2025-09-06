package monolith

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	lambdarouter "github.com/code-inbox/mason-go/lambda"
	"github.com/magicbell/magicbell/src/gofoundation/web"
	"github.com/magicbell/mason"
)

// ServerOptions configures the shared Server behavior.
type ServerOptions struct {
	// VersionPrefix adds a prefix to all registered routes (e.g., "v1").
	VersionPrefix string
	// EnableCORS applies permissive CORS headers in ServeHTTP when true.
	EnableCORS bool
	// LambdaAware enables AWS Lambda router when AWS_LAMBDA_FUNCTION_NAME is set.
	LambdaAware bool
	// DefaultPort use this port when env.PORT is not set
	DefaultPort string
}

// Server is a shared Mason runtime server with optional CORS and Lambda support.
type Server struct {
	*web.Server
	log     *slog.Logger
	options ServerOptions
	API     *mason.API
}

type SetupRoutesFunc func(*mason.API)

// NewServer constructs a Server with the provided options.
func NewServer(log *slog.Logger, opts ServerOptions, handlers ...SetupRoutesFunc) *Server {
	srv := &Server{
		Server:  web.NewServer(log),
		log:     log,
		options: opts,
	}

	srv.API = mason.NewAPI(srv)

	for _, h := range handlers {
		h(srv.API)
	}

	return srv
}

// Run starts the HTTP server locally or as a Lambda handler when enabled.
func (s *Server) Run() error {
	if s.options.LambdaAware && os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		l := lambdarouter.HTTP{App: s}
		lambda.Start(l.APIGWHandler)
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = s.options.DefaultPort
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: s,
	}

	fmt.Printf("listening on :%s\n", port)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("server.ListenAndServe: %w", err)
	}
	return nil
}

// Ensure Server implements mason.Runtime.
var _ mason.Runtime = (*Server)(nil)

// Handle registers a route using the configured version prefix.
func (s *Server) Handle(method string, path string, handler mason.WebHandler, mws ...func(mason.WebHandler) mason.WebHandler) {
	mids := make([]web.Middleware, 0, len(mws))
	for _, mw := range mws {
		mids = append(mids, web.Middleware{Handler: mw})
	}
	s.NewRoute(method, s.options.VersionPrefix, path, handler, mids...)
}

// Respond writes a JSON response using the shared web.Respond helper.
func (s *Server) Respond(ctx context.Context, w http.ResponseWriter, data any, status int) error {
	return web.Respond(ctx, w, data, status)
}

// ServeHTTP applies optional CORS and delegates to the embedded router.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.options.EnableCORS {
		h := w.Header()
		h.Set("Vary", "Origin")
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Requested-With")
		h.Set("Access-Control-Expose-Headers", "Content-Length,Content-Type")
		h.Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	s.Server.ServeHTTP(w, r)
}
