package monolith

import (
	"net/http"
	"testing"
)

type TestMonolith interface {
	Monolith
	http.Handler
	ResetAndSeed(t *testing.T, data []byte)
}
