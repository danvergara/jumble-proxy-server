package server

import (
	"log"
	"net/http"
)

func loggingMiddlware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		site := r.PathValue("site")
		log.Printf("fetching site: %s\n", site)
		next.ServeHTTP(w, r)
	})
}
