/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
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

		http.HandleFunc("GET /sites/{site}", proxyHandler)
		if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		log.Printf("server listening on port %s\n", port)

		return nil
	},
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	site := r.PathValue("site")

	// Send request to the target site.
	req, err := http.NewRequest(r.Method, site, r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy headers from original request.
	for header, values := range r.Header {
		for _, value := range values {
			req.Header.Add(header, value)
		}
	}

	// Perform the proxy request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Proxy request failed %v", err)
		http.Error(w, "Proxy request failed", http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	// Copy the response headers.
	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	// Set the status code and write the response body.
	w.WriteHeader(resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	w.Write(body)
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().
		StringVarP(&allowedOrigin, "allowed-origin", "a", "", "Restrict access to a specific allowed domain")
	serverCmd.Flags().
		StringVarP(&port, "port", "p", "", "Server Port")
}
