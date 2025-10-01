// Package monolith provides the monolith interface.
package monolith

import (
	"log/slog"

	"github.com/magicbell/mason"
	"github.com/tailbits/costwatch/internal/appconfig"
)

type Monolith interface {
	Config() appconfig.Config

	// Returns a named logger
	Logger(name string) *slog.Logger

	// v2
	API() *mason.API
}
