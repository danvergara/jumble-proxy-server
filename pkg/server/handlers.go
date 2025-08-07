package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/danvergara/jumble-proxy-server/pkg/config"
	"github.com/danvergara/jumble-proxy-server/pkg/github"
)

// proxyHandler adds headers to overcome the CORS errors for the Jumble Nostr client.
func proxyHandler(cfg *config.Config) func(w http.ResponseWriter, r *http.Request) {
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

			value, err := cfg.Cache.Get([]byte(site))
			// The Get method returns not found error when the key does not exist in the cache.
			if err != nil {
				resp, err := gc.GenerateGithubOpenGraph(r.Context(), site)
				if err != nil {
					http.Error(
						w,
						"Failed to generate the GitHub Open Graph HTML response",
						http.StatusInternalServerError,
					)
					return
				}

				cfg.Logger.Info(
					fmt.Sprintf("Fetch Open Graph data from the GitHub API for the site: %s", site),
				)

				// Stores the HTML file with the Open Graph data.
				// Expire in 1 hour.
				if err := cfg.Cache.Set([]byte(site), []byte(resp), 3600); err != nil {
					cfg.Logger.Error(
						fmt.Sprintf(
							"Failed to store the html file in the cache from the site: %s",
							site,
						),
					)
				}

				cfg.Logger.Info(
					fmt.Sprintf(
						"HTML document with Open Graph data from GitHub successfully stored in the cache for the %s site",
						site,
					),
				)

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(resp))
				return
			}

			cfg.Logger.Info(
				fmt.Sprintf(
					"HTML document with Open Graph data from Github found in cache for the %s site",
					site,
				),
			)

			// Return the stored HTML document.
			w.WriteHeader(http.StatusOK)
			w.Write(value)
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
			cfg.Logger.Error(
				fmt.Sprintf("Proxy error - URL: %s, Error: %v, Error Type: %T", site, err, err),
			)
			http.Error(w, fmt.Sprintf("proxy request failed: %v", err), http.StatusBadGateway)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusBadRequest {
			switch resp.StatusCode {
			case http.StatusTooManyRequests:
				cfg.Logger.Error(
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
				cfg.Logger.Error(
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
				cfg.Logger.Error(
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
				cfg.Logger.Error(
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
		cfg.Logger.Info(fmt.Sprintf("Proxy success - URL: %s, Status: %d", site, resp.StatusCode))
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
