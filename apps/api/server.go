package api

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

type Server struct {
	*web.Server
	log *slog.Logger
}

func NewServer(log *slog.Logger) *Server {
	server := web.NewServer(
		log,
	)

	return &Server{
		Server: server,
		log:    log,
	}
}

func (s *Server) Run(port string) error {
	// ===========================================================================
	// Lambda
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		l := lambdarouter.HTTP{
			App: s,
		}

		lambda.Start(l.APIGWHandler)
	}

	// Local
	server := &http.Server{
		Addr:    ":" + port,
		Handler: s,
	}

	lp := port
	if pp, ok := os.LookupEnv("PROXY_PORT"); ok {
		lp = pp
	}
	fmt.Printf("listening on :%s\n", lp)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("server.ListenAndServe: %w", err)
	}

	return nil
}

// ==========================================================================
var _ mason.Runtime = (*Server)(nil)

func (s *Server) Handle(method string, path string, handler mason.WebHandler, mws ...func(mason.WebHandler) mason.WebHandler) {
	mids := make([]web.Middleware, 0, len(mws))
	for _, mw := range mws {
		mids = append(mids, web.Middleware{
			Handler: mw,
		})
	}

	s.NewRoute(method, "v1", path, handler, mids...)
}

func (s *Server) Respond(ctx context.Context, w http.ResponseWriter, data any, status int) error {
	return web.Respond(ctx, w, data, status)
}
