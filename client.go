package api2c2p

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"

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

	// PaymentGatewayURL is the base URL for payment gateway API requests (e.g. payment inquiry)
	// Default: https://sandbox-pgw.2c2p.com
	PaymentGatewayURL string

	// FrontendURL is the base URL for frontend-related API requests (e.g. secure fields, refunds)
	// Default: https://demo2.2c2p.com
	FrontendURL string

	// PrivateKeyFile is the path to the combined private key and certificate PEM file
	PrivateKey *rsa.PrivateKey
	PublicCert *x509.Certificate

	// ServerJWTPublicKeyFile is the path to the 2C2P's public key certificate (.cer file) for JWT
	ServerJWTPublicCert *x509.Certificate

	// ServerPKCS7PublicKeyFile is the path to the 2C2P's public key certificate (.cer file) for PKCS7
	ServerPKCS7PublicCert *x509.Certificate
}

// Config holds the configuration for creating a new 2C2P client
type Config struct {
	SecretKey                string
	MerchantID               string
	HttpClient               *http.Client
	PaymentGatewayURL        string // URL for payment gateway APIs
	FrontendURL              string // URL for frontend-related APIs
	CombinedPEM              string
	ServerJWTPublicKeyFile   string
	ServerPKCS7PublicKeyFile string
}

// NewClient creates a new 2C2P API client
func NewClient(cfg Config) (*Client, error) {
	privateKey, publicCert, err := loadPrivateKeyAndCert(cfg.CombinedPEM)
	if err != nil {
		return nil, err
	}
	serverJWTPublicKey, err := serverPublicCert(cfg.ServerJWTPublicKeyFile)
	if err != nil {
		return nil, err
	}
	serverPKCS7PublicKey, err := serverPublicCert(cfg.ServerPKCS7PublicKeyFile)
	if err != nil {
		return nil, err
	}

	if cfg.PaymentGatewayURL == "" {
		cfg.PaymentGatewayURL = "https://sandbox-pgw.2c2p.com"
	}
	if cfg.FrontendURL == "" {
		cfg.FrontendURL = "https://demo2.2c2p.com"
	}
	if cfg.HttpClient == nil {
		cfg.HttpClient = &http.Client{}
	}
	loggingClient := NewLoggingClient(cfg.HttpClient, nil, true)
	return &Client{
		SecretKey:             cfg.SecretKey,
		MerchantID:            cfg.MerchantID,
		httpClient:            loggingClient,
		PaymentGatewayURL:     cfg.PaymentGatewayURL,
		FrontendURL:           cfg.FrontendURL,
		PrivateKey:            privateKey,
		PublicCert:            publicCert,
		ServerJWTPublicCert:   serverJWTPublicKey,
		ServerPKCS7PublicCert: serverPKCS7PublicKey,
	}, nil
}

func (c *Client) paymentGatewayEndpoint(path string) string {
	return fmt.Sprintf("%s/payment/4.3/%s", c.PaymentGatewayURL, path)
}

func (c *Client) frontendEndpoint(path string) string {
	return fmt.Sprintf("%s/%s", c.FrontendURL, path)
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

//

func serverPublicCert(serverPublicKeyFile string) (*x509.Certificate, error) {
	// Read and parse 2C2P's public key certificate
	certPEM, err := os.ReadFile(serverPublicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read server public key file: %#v: %w", serverPublicKeyFile, err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode server public key PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse server certificate: %w", err)
	}
	return cert, nil
}

func loadPrivateKeyAndCert(combinedPEMFile string) (*rsa.PrivateKey, *x509.Certificate, error) {
	// Read the combined PEM file
	pemData, err := os.ReadFile(combinedPEMFile)
	if err != nil {
		return nil, nil, fmt.Errorf("read private key file: %w", err)
	}

	// Parse private key
	var privateKey *rsa.PrivateKey
	var cert *x509.Certificate
	for {
		block, rest := pem.Decode(pemData)
		if block == nil {
			break
		}
		switch block.Type {
		case "RSA PRIVATE KEY", "PRIVATE KEY":
			if privateKey == nil {
				if block.Type == "RSA PRIVATE KEY" {
					privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
				} else {
					var key interface{}
					key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
					if err == nil {
						var ok bool
						privateKey, ok = key.(*rsa.PrivateKey)
						if !ok {
							err = fmt.Errorf("not an RSA private key")
						}
					}
				}
				if err != nil {
					return nil, nil, fmt.Errorf("parse private key: %w", err)
				}
			}
		case "CERTIFICATE":
			if cert == nil {
				cert, err = x509.ParseCertificate(block.Bytes)
				if err != nil {
					return nil, nil, fmt.Errorf("parse certificate: %w", err)
				}
			}
		}
		pemData = rest
	}

	if privateKey == nil {
		return nil, nil, fmt.Errorf("no private key found")
	}
	if cert == nil {
		return nil, nil, fmt.Errorf("no certificate found")
	}

	return privateKey, cert, nil
}
