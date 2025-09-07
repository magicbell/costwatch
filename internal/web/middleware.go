package web

import (
	"embed"

	"github.com/magicbell/mason"
)

// Middleware is a function designed to run some code before and/or after
// another Handler. It is designed to remove boilerplate or other concerns not
// direct to any given Handler.
type Middleware struct {
	Name    string
	Handler func(mason.WebHandler) mason.WebHandler
}

// APIv2 middleware
func (m Middleware) GetHandler(_ mason.Builder) func(mason.WebHandler) mason.WebHandler {
	return m.Handler
}

type JSONSchemaMiddleware struct {
	Handler func(mason.WebHandler) mason.WebHandler
	Schemas embed.FS
	Base    string
}

func (j JSONSchemaMiddleware) GetSchemas() embed.FS {
	return j.Schemas
}
func (j JSONSchemaMiddleware) GetBase() string {
	return j.Base
}

// wrapMiddleware creates a new handler by wrapping middleware around a final
// handler. The middlewares' Handlers will be executed by requests in the order
// they are provided.
func wrapMiddleware(mw []Middleware, handler mason.WebHandler) mason.WebHandler {

	// Loop backwards through the middleware invoking each one. Replace the
	// handler with the new wrapped handler. Looping backwards ensures that the
	// first middleware of the slice is the first to be executed by requests.
	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		handler = h.Handler(handler)
	}

	return handler
}
