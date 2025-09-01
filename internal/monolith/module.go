package monolith

import (
	"context"

	"github.com/magicbell/mason"
)

type Module interface {
	Name() string
	Startup(context.Context, Monolith) error
}

type APIModule interface {
	SetupAPI(ctx context.Context, grp *mason.RouteGroup) error
}
