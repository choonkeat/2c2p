package testutil

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// AssertRequest verifies that an http.Request matches expected values
func AssertRequest(t *testing.T, req *http.Request, want struct {
	Method      string            // Expected HTTP method (e.g., "POST", "GET")
	URL         string            // Expected URL
	ContentType string            // Expected Content-Type header
	Headers     map[string]string // Expected headers (optional)
	Body        any               // Expected body (will be compared as JSON)
}) {
	t.Helper()

	// Check method
	if req.Method != want.Method {
		t.Errorf("Method = %q, want %q", req.Method, want.Method)
	}

	// Check URL
	if req.URL.String() != want.URL {
		t.Errorf("URL = %q, want %q", req.URL.String(), want.URL)
	}

	// Check Content-Type
	if got := req.Header.Get("Content-Type"); got != want.ContentType {
		t.Errorf("Content-Type = %q, want %q", got, want.ContentType)
	}

	// Check other headers
	for k, v := range want.Headers {
		if got := req.Header.Get(k); got != v {
			t.Errorf("Header[%q] = %q, want %q", k, got, v)
		}
	}

	// Check body if provided
	if want.Body != nil {
		// Read and restore the body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(body))

		// For JWT tokens, decode and compare the payload
		var reqBody struct {
			Payload string `json:"payload"`
		}
		if err := json.Unmarshal(body, &reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// If this is a JWT token (has 2 dots), decode and compare the payload
		if strings.Count(reqBody.Payload, ".") == 2 {
			parts := strings.Split(reqBody.Payload, ".")
			if len(parts) != 3 {
				t.Fatalf("Invalid JWT token format")
			}

			// Compare only the payload (middle part)
			payload, err := decodeJWTPayload(parts[1])
			if err != nil {
				t.Fatalf("Failed to decode JWT payload: %v", err)
			}

			// Compare payload with expected body
			var gotMap, wantMap map[string]any
			if err := json.Unmarshal(payload, &gotMap); err != nil {
				t.Fatalf("Failed to decode JWT payload: %v", err)
			}
			wantBytes, err := json.Marshal(want.Body)
			if err != nil {
				t.Fatalf("Failed to marshal expected body: %v", err)
			}
			if err := json.Unmarshal(wantBytes, &wantMap); err != nil {
				t.Fatalf("Failed to decode expected body: %v", err)
			}

			if !reflect.DeepEqual(gotMap, wantMap) {
				t.Errorf("JWT payload = %#v, want %#v", gotMap, wantMap)
			}
		} else {
			// Regular JSON comparison for non-JWT bodies
			var gotMap, wantMap map[string]any
			if err := json.Unmarshal(body, &gotMap); err != nil {
				t.Fatalf("Failed to decode actual body: %v", err)
			}
			wantBytes, err := json.Marshal(want.Body)
			if err != nil {
				t.Fatalf("Failed to marshal expected body: %v", err)
			}
			if err := json.Unmarshal(wantBytes, &wantMap); err != nil {
				t.Fatalf("Failed to decode expected body: %v", err)
			}

			if !reflect.DeepEqual(gotMap, wantMap) {
				t.Errorf("Body = %#v, want %#v", gotMap, wantMap)
			}
		}
	}
}

// decodeJWTPayload decodes the base64-encoded payload of a JWT token
func decodeJWTPayload(payload string) ([]byte, error) {
	// Add padding if needed
	if l := len(payload) % 4; l > 0 {
		payload += strings.Repeat("=", 4-l)
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

func mustMarshalJSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	return string(data)
}
