package server

import (
	"net/http"

	"github.com/danvergara/jumble-proxy-server/pkg/config"
)

// addRoutes function adds the handler to the server mux.
func addRoutes(mux *http.ServeMux, _ *config.Config) {
	proxy := http.HandlerFunc(proxyHandler())
	mux.Handle("GET /sites/{site}", loggingMiddlware(proxy))
}
