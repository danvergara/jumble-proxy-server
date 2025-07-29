package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
)

// proxyHandler adds headers to overcome the CORS errors for the Jumble Nostr client.
func proxyHandler() func(w http.ResponseWriter, r *http.Request) {
	// Get token from environment variable
	githubToken := os.Getenv("JUMBLE_PROXY_GITHUB_TOKEN")
	return func(w http.ResponseWriter, r *http.Request) {
		// add the paraters to fix CORS issues.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// get the site of interest from the path parameters.
		site := r.PathValue("site")

		// Send request to the target site.
		req, err := http.NewRequest(r.Method, site, r.Body)
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}

		// Check if the target URL is GitHub.
		if isGitHubURL(site) {
			// Set GitHub-specific headers in the proxy.
			req.Header.Set("Authorization", "Bearer "+githubToken)
			req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; JumbleProxy/0.1)")
			req.Header.Set(
				"Accept",
				"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			)
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
			fmt.Printf("proxy request failed %v", err)
			http.Error(w, "proxy request failed", http.StatusBadGateway)
			return
		}

		defer resp.Body.Close()

		// Copy the response headers.
		for header, values := range resp.Header {
			for _, value := range values {
				// If w.Header contains the header and the value is already in it, continue
				if _, ok := w.Header()[header]; ok && slices.Contains(w.Header()[header], value) {
					continue
				}
				w.Header().Add(header, value)
			}
		}

		// Set the status code and write the response body.
		w.WriteHeader(resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		w.Write(body)
	}
}

func isGitHubURL(url string) bool {
	return strings.Contains(strings.ToLower(url), "github.com") ||
		strings.Contains(strings.ToLower(url), "api.github.com") ||
		strings.Contains(strings.ToLower(url), "gist.github.com")
}
