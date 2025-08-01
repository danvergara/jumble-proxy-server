/*
Copyright © 2025 Daniel Vergara  daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/danvergara/jumble-proxy-server/pkg/config"
	"github.com/danvergara/jumble-proxy-server/pkg/server"
)

var (
	port string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Golang Backend Server as a Proxy to overcome CORS errors for the Jumble Nostr client",
	Long: `This application is a proxy server used by the Jumble Nostr client as a workaround to fix CORS erros,
so that the client can show the URL preview from links' Open Graph data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if port == "" {
			port = "8000"
		}

		jsonHandler := slog.NewJSONHandler(os.Stderr, nil)

		logger := slog.New(jsonHandler)

		cfg := config.Config{
			Port:   port,
			Logger: logger,
		}

		logger.Info(fmt.Sprintf("Server listening on port %s", port))

		ctx := context.Background()
		if err := server.Run(ctx, &cfg); err != nil {
			logger.Error(fmt.Sprintf("Error running the server: %s", err))
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	port = os.Getenv("PORT")
}
