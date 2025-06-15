package server

import (
	"fmt"
	"io"
	"net/http"
)

// proxyHandler adds headers to overcome the CORS errors for the Jumble Nostr client.
func proxyHandler() func(w http.ResponseWriter, r *http.Request) {
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
				w.Header().Add(header, value)
			}
		}

		// Set the status code and write the response body.
		w.WriteHeader(resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		w.Write(body)
	}
}
