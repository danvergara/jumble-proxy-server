package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/danvergara/jumble-proxy-server/pkg/github"
)

// proxyHandler adds headers to overcome the CORS errors for the Jumble Nostr client.
func proxyHandler(logger *slog.Logger) func(w http.ResponseWriter, r *http.Request) {
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
			gc := github.New(githubToken)
			resp, err := gc.GenerateGithubOpenGraph(r.Context(), site)
			if err != nil {
				http.Error(
					w,
					"Failed to generate the GitHub Open Graph HTML response",
					http.StatusInternalServerError,
				)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(resp))
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
			// More detailed logging for debugging the 502 issue
			logger.Error(
				fmt.Sprintf("Proxy error - URL: %s, Error: %v, Error Type: %T", site, err, err),
			)
			http.Error(w, fmt.Sprintf("proxy request failed: %v", err), http.StatusBadGateway)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusBadRequest {
			switch resp.StatusCode {
			case http.StatusTooManyRequests:
				logger.Error(
					fmt.Sprintf(
						"Rate limit exceeded - URL: %s, Status: %d",
						site,
						resp.StatusCode,
					),
				)
				http.Error(
					w,
					fmt.Sprintf("Rate limit exceeded for site %s", site),
					http.StatusTooManyRequests,
				)
				return
			case http.StatusForbidden:
				logger.Error(
					fmt.Sprintf(
						"Access denied - URL: %s, Status: %d",
						site,
						resp.StatusCode,
					),
				)
				http.Error(
					w,
					fmt.Sprintf("Access forbidden for site %s", site),
					http.StatusForbidden,
				)
				return
			case http.StatusServiceUnavailable:
				logger.Error(
					fmt.Sprintf(
						"Service unavailable - URL: %s, Status: %d",
						site,
						resp.StatusCode,
					),
				)
				http.Error(
					w,
					fmt.Sprintf("Service temporarily unavailable for site %s", site),
					http.StatusServiceUnavailable,
				)
				return
			default:
				logger.Error(
					fmt.Sprintf(
						"Proxy error - URL: %s, Status: %d",
						site,
						resp.StatusCode,
					),
				)
				http.Error(w, fmt.Sprintf("Request failed for site %s", site), resp.StatusCode)
				return
			}
		}
		// Log successful requests too, to see what's working
		logger.Info(fmt.Sprintf("Proxy success - URL: %s, Status: %d", site, resp.StatusCode))
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
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("Error copying response body: %v", err))
		}
	}
}

func isGitHubURL(url string) bool {
	return strings.Contains(strings.ToLower(url), "github.com") ||
		strings.Contains(strings.ToLower(url), "api.github.com") ||
		strings.Contains(strings.ToLower(url), "gist.github.com")
}
