package api2c2p

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a 2C2P API client
type Client struct {
	// SecretKey is the merchant's secret key for signing tokens
	SecretKey string

	// MerchantID is the merchant's unique identifier
	MerchantID string

	// httpClient is the HTTP client used for making requests
	httpClient *http.Client

	// BaseURL is the base URL for API requests
	BaseURL string
}

// NewClient creates a new 2C2P API client
func NewClient(secretKey, merchantID string, baseURL ...string) *Client {
	url := "https://sandbox-pgw.2c2p.com"
	if len(baseURL) > 0 {
		url = baseURL[0]
	}
	return &Client{
		SecretKey:  secretKey,
		MerchantID: merchantID,
		httpClient: &http.Client{},
		BaseURL:    url,
	}
}

func (c *Client) endpoint(path string) string {
	return fmt.Sprintf("%s/payment/4.3/%s", c.BaseURL, path)
}

// generateJWTToken generates a JWT token for the given payload
func (c *Client) generateJWTToken(payload []byte) (string, error) {
	// Create header
	header := map[string]string{
		"typ": "JWT",
		"alg": "HS256",
	}

	// Encode header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}
	headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode payload
	payloadBase64 := base64.RawURLEncoding.EncodeToString(payload)

	// Create signature
	signatureInput := headerBase64 + "." + payloadBase64
	h := hmac.New(sha256.New, []byte(c.SecretKey))
	h.Write([]byte(signatureInput))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	// Combine all parts
	token := headerBase64 + "." + payloadBase64 + "." + signature
	return token, nil
}

// decodeJWTToken decodes a JWT token and verifies its signature
func (c *Client) decodeJWTToken(token string, v interface{}) error {
	parts := bytes.Split([]byte(token), []byte{'.'})
	if len(parts) != 3 {
		return fmt.Errorf("invalid token format")
	}

	// Verify signature
	h := hmac.New(sha256.New, []byte(c.SecretKey))
	h.Write([]byte(string(parts[0]) + "." + string(parts[1])))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	if signature != string(parts[2]) {
		return fmt.Errorf("invalid token")
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(string(parts[1]))
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	// Unmarshal payload
	if err := json.Unmarshal(payload, v); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return nil
}

type debugRequest struct {
	URL     string
	Headers http.Header
	Body    string
}

type debugResponse struct {
	Status      string
	Headers     http.Header
	Body        string
	ElapsedTime time.Duration
}

type debugInfo struct {
	Request  debugRequest
	Response debugResponse
}

func (c *Client) doRequestWithDebug(req *http.Request) (*http.Response, *debugInfo, error) {
	// Create debug info
	debug := &debugInfo{
		Request: debugRequest{
			URL:     req.URL.String(),
			Headers: req.Header,
		},
	}

	// Read and restore request body
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, debug, fmt.Errorf("read request body: %w", err)
		}
		debug.Request.Body = string(body)
		req.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	// Make request
	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, debug, err
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, debug, fmt.Errorf("read response body: %w", err)
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	// Update debug info
	debug.Response = debugResponse{
		Status:      resp.Status,
		Headers:     resp.Header,
		Body:        string(body),
		ElapsedTime: time.Since(start),
	}

	return resp, debug, nil
}

func (c *Client) formatErrorWithDebug(err error, debug *debugInfo) error {
	if debug == nil {
		return err
	}

	// Try to extract response code from body
	var response struct {
		RespCode string `json:"respCode"`
		RespDesc string `json:"respDesc"`
	}
	if unmarshalError := json.Unmarshal([]byte(debug.Response.Body), &response); unmarshalError == nil {
		respCode := PaymentResponseCode(response.RespCode)
		return fmt.Errorf("%w\nRequest URL: %s\nRequest Headers: %v\nRequest Body: %s\nResponse Status: %s\nResponse Headers: %v\nResponse Body: %s\nResponse Time: %v\nResponse Code: %s (%s)",
			err,
			debug.Request.URL,
			debug.Request.Headers,
			debug.Request.Body,
			debug.Response.Status,
			debug.Response.Headers,
			debug.Response.Body,
			debug.Response.ElapsedTime,
			respCode,
			respCode.Description(),
		)
	}

	return fmt.Errorf("%w\nRequest URL: %s\nRequest Headers: %v\nRequest Body: %s\nResponse Status: %s\nResponse Headers: %v\nResponse Body: %s\nResponse Time: %v",
		err,
		debug.Request.URL,
		debug.Request.Headers,
		debug.Request.Body,
		debug.Response.Status,
		debug.Response.Headers,
		debug.Response.Body,
		debug.Response.ElapsedTime,
	)
}

func (c *Client) newRequest(method, path string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, c.endpoint(path), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
