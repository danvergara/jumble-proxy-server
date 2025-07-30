package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/danvergara/jumble-proxy-server/pkg/config"
)

const htmlContent = `
<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
</head>
<body>
    <h1>Hello from Test Server!</h1>
    <p>This is a test HTML page served by httptest.</p>
</body>
</html>
`

func htmlHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(htmlContent))
}

func TestServer(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	cfg := config.Config{
		Port: "8080",
	}

	site := httptest.NewServer(http.HandlerFunc(htmlHandler))
	defer site.Close()

	go Run(ctx, &cfg)

	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/sites/%s", url.QueryEscape(site.URL)))
	if err != nil {
		t.Fatalf("Failed to make request to the site through the proxy server: %v", err)
	}
	defer resp.Body.Close()

	allowOrigin := resp.Header["Access-Control-Allow-Origin"]
	if len(allowOrigin) > 1 {
		t.Fatalf("Access-Control-Allow-Origin header contains undesired values %v\n", allowOrigin)
	}

	if allowOrigin[0] != "*" {
		t.Fatalf("missing Access-Control-Allow-Origin header")
	}

	// Read the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Validate exact HTML content.
	if string(body) != htmlContent {
		t.Errorf("HTML content mismatch.\nExpected:\n%s\nGot:\n%s",
			htmlContent, string(body))
	}
}

func TestProxyHandlerGitHub(t *testing.T) {
	// Set up GitHub token for testing.
	originalToken := os.Getenv("JUMBLE_PROXY_GITHUB_TOKEN")
	os.Setenv("JUMBLE_PROXY_GITHUB_TOKEN", "test-github-token")
	defer func() {
		if originalToken == "" {
			os.Unsetenv("JUMBLE_PROXY_GITHUB_TOKEN")
		} else {
			os.Setenv("JUMBLE_PROXY_GITHUB_TOKEN", originalToken)
		}
	}()

	// Create a mock GitHub server.
	mockGitHubServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify GitHub-specific headers were added by the proxy.
			if auth := r.Header.Get("Authorization"); auth != "Bearer test-github-token" {
				t.Errorf("Expected Authorization header 'Bearer test-github-token', got '%s'", auth)
			}
			if userAgent := r.Header.Get("User-Agent"); userAgent != "Mozilla/5.0 (compatible; JumbleProxy/0.1)" {
				t.Errorf(
					"Expected User-Agent 'Mozilla/5.0 (compatible; JumbleProxy/0.1)', got '%s'",
					userAgent,
				)
			}
			if accept := r.Header.Get("Accept"); accept != "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8" {
				t.Errorf("Expected specific Accept header, got '%s'", accept)
			}

			// Send back a mock GitHub response.
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-GitHub-Media-Type", "github.v3")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "Hello from GitHub!"}`))
		}),
	)
	defer mockGitHubServer.Close()

	// Test cases for different GitHub URL formats.
	testCases := []struct {
		name string
		url  string
	}{
		{"github.com", mockGitHubServer.URL + "/github.com/user/repo"},
		{"api.github.com", mockGitHubServer.URL + "/api.github.com/repos/user/repo"},
		{"gist.github.com", mockGitHubServer.URL + "/gist.github.com/user/gist-id"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a request to the proxy handler.
			req := httptest.NewRequest("GET", "/sites/"+url.QueryEscape(tc.url), nil)
			req.SetPathValue("site", tc.url)

			// Create a response recorder.
			w := httptest.NewRecorder()

			// Call the proxy handler.
			handler := proxyHandler()
			handler(w, req)

			// Check response status.
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			// Validate CORS headers.
			expectedHeaders := map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			}

			for header, expectedValue := range expectedHeaders {
				if actualValue := w.Header().Get(header); actualValue != expectedValue {
					t.Errorf(
						"Expected %s header '%s', got '%s'",
						header,
						expectedValue,
						actualValue,
					)
				}
			}

			// Validate that GitHub response headers are proxied.
			if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}

			if githubHeader := w.Header().Get("X-GitHub-Media-Type"); githubHeader != "github.v3" {
				t.Errorf("Expected X-GitHub-Media-Type 'github.v3', got '%s'", githubHeader)
			}

			// Validate response body.
			expectedBody := `{"message": "Hello from GitHub!"}`
			if actualBody := w.Body.String(); actualBody != expectedBody {
				t.Errorf("Expected body '%s', got '%s'", expectedBody, actualBody)
			}
		})
	}
}

func TestIsGitHubURL(t *testing.T) {
	testCases := []struct {
		url      string
		expected bool
	}{
		{"https://github.com/user/repo", true},
		{"https://api.github.com/repos/user/repo", true},
		{"https://gist.github.com/user/gist-id", true},
		{"https://GITHUB.COM/user/repo", true}, // case insensitive
		{"https://example.com", false},
		{"https://gitlab.com/user/repo", false},
	}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			result := isGitHubURL(tc.url)
			if result != tc.expected {
				t.Errorf("isGitHubURL(%s) = %v, expected %v", tc.url, result, tc.expected)
			}
		})
	}
}
