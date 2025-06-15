package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	if allowOrigin[0] != "*" {
		t.Fatalf("missing Access-Control-Allow-Origin headers")
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
