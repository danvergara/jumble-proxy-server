package server

import (
	"net/http"

	"github.com/danvergara/jumble-proxy-server/pkg/config"
)

// addRoutes function adds the handler to the server mux.
func addRoutes(mux *http.ServeMux, cfg *config.Config) {
	proxy := http.HandlerFunc(proxyHandler(cfg.Logger))
	mux.Handle("GET /sites/{site}", loggingMiddlware(proxy, cfg.Logger))
}
