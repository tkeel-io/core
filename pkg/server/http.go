package server

import (
	"github.com/tkeel-io/kit/transport/http"
)

// NewHTTPServer new a HTTP server.
func NewHTTPServer(addr string) *http.Server {
	srv := http.NewServer(addr)
	return srv
}
