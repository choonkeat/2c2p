package api2c2p

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Client represents a 2C2P API client
type Client struct {
	MerchantID  string
	SecretKey   string
	BaseURL     string
	HTTPClient  *http.Client
}

// NewClient creates a new 2C2P API client
func NewClient(merchantID, secretKey, baseURL string) *Client {
	return &Client{
		MerchantID: merchantID,
		SecretKey:  secretKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

// GenerateJWTToken generates a JWT token for the given payload
func (c *Client) GenerateJWTToken(payload []byte) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return "", fmt.Errorf("error unmarshaling payload: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(data))
	signedToken, err := token.SignedString([]byte(c.SecretKey))
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}

	return signedToken, nil
}

// DecodeJWTToken decodes a JWT token into the provided response struct
func (c *Client) DecodeJWTToken(tokenString string, response interface{}) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(c.SecretKey), nil
	})
	if err != nil {
		return fmt.Errorf("error parsing token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		jsonData, err := json.Marshal(claims)
		if err != nil {
			return fmt.Errorf("error marshaling claims: %v", err)
		}

		if err := json.Unmarshal(jsonData, response); err != nil {
			return fmt.Errorf("error unmarshaling claims: %v", err)
		}

		return nil
	}

	return fmt.Errorf("invalid token")
}
