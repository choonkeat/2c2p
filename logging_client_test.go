package api2c2p

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingClient(t *testing.T) {
	// Create a buffer to capture log output
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "test-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response body"))
	}))
	defer server.Close()

	// Create our logging client
	client := NewLoggingClient(nil, logger, true)

	// Create a test request with a body
	reqBody := strings.NewReader("request body")
	req, err := http.NewRequest("POST", server.URL, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if string(body) != "response body" {
		t.Errorf("Unexpected response body: %s", string(body))
	}

	// Check log output
	logOutput := logBuf.String()

	// Define expected log entries
	expectedEntries := []string{
		"[REQUEST] POST " + server.URL,
		"[REQUEST HEADERS] map[Content-Type:[text/plain]]",
		"[REQUEST BODY] request body",
		"[RESPONSE] 200 OK",
		"[RESPONSE BODY] response body",
	}

	for _, expected := range expectedEntries {
		if !strings.Contains(logOutput, expected) {
			t.Errorf("Log output missing expected content: %q\nGot log output:\n%s", expected, logOutput)
		}
	}
}
