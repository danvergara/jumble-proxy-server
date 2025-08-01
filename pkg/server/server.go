package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/danvergara/jumble-proxy-server/pkg/config"
)

// NewServer constructor returns an http.Handler if possible, which can be a dedicated type for more complex situations.
// It configures its own muxer and calls out to routes.go
func NewServer(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, cfg)
	var handler http.Handler = mux
	return handler
}

// Run the proxy server and will help the server to gracefully shut down.
func Run(ctx context.Context, cfg *config.Config) error {
	// Creates a context and it's cancelled if there's Interrupt signal.
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Creates a new http.Server based on the Server struct.
	srv := NewServer(cfg)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
		Handler: srv,
	}

	// Runs the server in a separate go routine.
	go func() {
		cfg.Logger.Info(fmt.Sprintf("Listening on %s", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cfg.Logger.Error(fmt.Sprintf("Error listening and serving: %s", err))
		}
	}()

	// Create a WaitGroup to wait on different goroutine that gracefully shuts down the server.
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		// Read from Done channel and it blocks until the context created by the NotifyContext function gets cancelled.
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		// This goroutine gracefully shuts down the server, but it only gives the process 10 seconds before for wrapping up and return a context's Error.
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			cfg.Logger.Error(fmt.Sprintf("Error shutting down http server: %s", err))
		}
	}()

	wg.Wait()

	return nil
}
