package server

import (
	"net/http"

	"github.com/danvergara/jumble-proxy-server/pkg/config"
)

// addRoutes function adds the handler to the server mux.
func addRoutes(mux *http.ServeMux, config *config.Config) {
	mux.HandleFunc("GET /sites/{site}", proxyHandler(config))
}
