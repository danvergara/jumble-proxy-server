package server

import (
	"fmt"
	"log/slog"
	"net/http"
)

func loggingMiddlware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		site := r.PathValue("site")
		logger.Info(fmt.Sprintf("Fetching site: %s", site))
		next.ServeHTTP(w, r)
	})
}
