/*
Copyright © 2025 Daniel Vergara  daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/danvergara/jumble-proxy-server/pkg/config"
	"github.com/danvergara/jumble-proxy-server/pkg/server"
)

var (
	port          string
	allowedOrigin string
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

		cfg := config.Config{
			Port:          port,
			AllowedOrigin: allowedOrigin,
		}

		log.Printf("server listening on port %s\n", port)

		ctx := context.Background()
		if err := server.Run(ctx, &cfg); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	port = os.Getenv("PORT")
	allowedOrigin = os.Getenv("ALLOW_ORIGIN")
}
