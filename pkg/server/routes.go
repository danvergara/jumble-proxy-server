package server

import (
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/danvergara/jumble-proxy-server/pkg/config"
)

// addRoutes function adds the handler to the server mux.
func addRoutes(mux *http.ServeMux, cfg *config.Config) {
	proxy := http.HandlerFunc(proxyHandler(cfg.Logger))
	mux.Handle("GET /sites/{site}", loggingMiddlware(proxy, cfg.Logger))

	// Add pprof routes only if enabled
	if os.Getenv("ENABLE_PPROF") == "true" {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
		mux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
		mux.Handle("/debug/pprof/block", pprof.Handler("block"))
		mux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))

		cfg.Logger.Info("pprof endpoints enabled at /debug/pprof/")
	}
}
