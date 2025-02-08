package api2c2p

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// Client represents a 2C2P API client
type Client struct {
	// SecretKey is the merchant's secret key for signing tokens
	SecretKey string

	// MerchantID is the merchant's unique identifier
	MerchantID string

	// httpClient is the HTTP client used for making requests
	httpClient *LoggingClient

	// BaseURL is the base URL for API requests
	BaseURL string

	// PrivateKeyFile is the path to the combined private key and certificate PEM file
	PrivateKeyFile string

	// ServerPublicKeyFile is the path to the 2C2P's public key certificate (.cer file)
	ServerPublicKeyFile string
}

// Config holds the configuration for creating a new 2C2P client
type Config struct {
	SecretKey           string
	MerchantID          string
	HttpClient          *http.Client
	BaseURL             string
	PrivateKeyFile      string
	ServerPublicKeyFile string
}

// NewClient creates a new 2C2P API client
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://sandbox-pgw.2c2p.com"
	}
	if cfg.HttpClient == nil {
		cfg.HttpClient = &http.Client{}
	}
	loggingClient := NewLoggingClient(cfg.HttpClient, nil, true)
	return &Client{
		SecretKey:           cfg.SecretKey,
		MerchantID:          cfg.MerchantID,
		httpClient:          loggingClient,
		BaseURL:             cfg.BaseURL,
		PrivateKeyFile:      cfg.PrivateKeyFile,
		ServerPublicKeyFile: cfg.ServerPublicKeyFile,
	}
}

func (c *Client) endpoint(path string) string {
	return fmt.Sprintf("%s/payment/4.3/%s", c.BaseURL, path)
}

func (c *Client) generateJWTTokenForJSON(payload []byte) (string, error) {
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("error unmarshaling payload: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	return token.SignedString([]byte(c.SecretKey))
}

func (c *Client) decodeJWTTokenForJSON(token string, v interface{}) error {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(c.SecretKey), nil
	})
	if err != nil {
		return err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return fmt.Errorf("invalid token")
	}

	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return fmt.Errorf("error marshaling claims: %v", err)
	}

	return json.Unmarshal(claimsBytes, v)
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	return resp, nil
}
