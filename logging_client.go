package api2c2p

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

// LoggingClient wraps an http.Client to provide logging capabilities
type LoggingClient struct {
	client  *http.Client
	logger  *log.Logger
	verbose bool
}

// NewLoggingClient creates a new LoggingClient
func NewLoggingClient(client *http.Client, logger *log.Logger, verbose bool) *LoggingClient {
	if client == nil {
		client = http.DefaultClient
	}
	if logger == nil {
		logger = log.Default()
	}
	return &LoggingClient{
		client:  client,
		logger:  logger,
		verbose: verbose,
	}
}

func (c *LoggingClient) Do(req *http.Request) (*http.Response, error) {
	// Log request
	if c.verbose {
		c.logRequest(req)
	}

	start := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		c.logger.Printf("[ERROR] Request failed: %v", err)
		return resp, err
	}

	// Log response
	if c.verbose {
		c.logResponse(resp, duration)
	} else {
		c.logger.Printf("[INFO] %s %s -> %s (%.3fs)", req.Method, req.URL, resp.Status, duration.Seconds())
	}

	return resp, nil
}

func (c *LoggingClient) logRequest(req *http.Request) {
	c.logger.Printf("[REQUEST] %s %s", req.Method, req.URL)
	c.logger.Printf("[REQUEST HEADERS] %v", req.Header)

	if req.Body != nil {
		if body, err := req.GetBody(); err == nil {
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, body); err == nil {
				c.logger.Printf("[REQUEST BODY] %s", buf.String())
				// Reset the body for the actual request
				req.Body = io.NopCloser(bytes.NewBuffer(buf.Bytes()))
			}
		}
	}
}

func (c *LoggingClient) logResponse(resp *http.Response, duration time.Duration) {
	c.logger.Printf("[RESPONSE] %s (%.3fs)", resp.Status, duration.Seconds())
	c.logger.Printf("[RESPONSE HEADERS] %v", resp.Header)

	if resp.Body != nil {
		// Create a copy of the body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.Printf("[ERROR] Failed to read response body: %v", err)
			return
		}
		// Replace the body for downstream consumers
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		// Log the body
		c.logger.Printf("[RESPONSE BODY] %s", string(bodyBytes))
	}
}
