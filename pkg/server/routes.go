package server

import (
	"net/http"

	"github.com/danvergara/jumble-proxy-server/pkg/config"
)

func addRoutes(mux *http.ServeMux, config *config.Config) {
	mux.HandleFunc("GET /sites/{site}", proxyHandler(config))
}
