package api2c2p

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewRefundRequest(t *testing.T) {
	client := NewClient(Config{
		SecretKey:           "test_secret",
		MerchantID:          "JT01",
		BaseURL:             "https://demo2.2c2p.com/2C2PFrontend",
		PrivateKeyFile:      "testdata/combined_private_public.pem",
		ServerPublicKeyFile: "testdata/server.public_cert.pem",
	})

	// Create refund request
	req := &PaymentProcessRequest{
		Version:      "4.3",
		MerchantID:   client.MerchantID,
		InvoiceNo:    "260121085327",
		ActionAmount: "25.00",
		ProcessType:  "R",
	}

	// Convert request to XML
	xmlData, err := xml.MarshalIndent(req, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// First encrypt with JWE
	jweToken, err := client.encryptWithJWE(xmlData)
	if err != nil {
		t.Fatalf("Failed to encrypt token: %v", err)
	}

	// Then sign with JWS PS256
	token := jwt.NewWithClaims(jwt.SigningMethodPS256, jwt.MapClaims{
		"request": jweToken,
	})

	// Load private key for signing
	privateKey, _, err := client.loadPrivateKeyAndCert()
	if err != nil {
		t.Fatalf("Failed to load private key: %v", err)
	}

	// Sign the token
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Create request
	httpReq, err := http.NewRequestWithContext(context.Background(), "POST", "https://demo2.2c2p.com/2C2PFrontend/PaymentAction/2.0/action", strings.NewReader(signedToken))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "text/plain")

	// Read request body
	body, err := io.ReadAll(httpReq.Body)
	if err != nil {
		t.Fatalf("Failed to read request body: %v", err)
	}

	// Re-create request body for subsequent reads
	httpReq.Body = io.NopCloser(bytes.NewBuffer(body))

	// Split JWS token into parts
	parts := strings.Split(string(body), ".")
	if len(parts) != 3 {
		t.Fatalf("Invalid JWS token format: expected 3 parts (header.payload.signature), got %d parts", len(parts))
	}

	// Verify JWS header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("Failed to decode JWS header: %v", err)
	}
	var header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		t.Fatalf("Failed to unmarshal JWS header: %v", err)
	}
	if header.Alg != "PS256" {
		t.Errorf("Expected alg PS256, got %s", header.Alg)
	}
	if header.Typ != "JWT" {
		t.Errorf("Expected typ JWT, got %s", header.Typ)
	}

	// Verify JWS payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("Failed to decode JWS payload: %v", err)
	}
	var payload struct {
		Request string `json:"request"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatalf("Failed to unmarshal JWS payload: %v", err)
	}

	// Split JWE token into parts
	jweParts := strings.Split(payload.Request, ".")
	if len(jweParts) != 5 {
		t.Fatalf("Invalid JWE token format: expected 5 parts (header.key.iv.ciphertext.tag), got %d parts", len(jweParts))
	}

	// Verify JWE header
	jweHeaderBytes, err := base64.RawURLEncoding.DecodeString(jweParts[0])
	if err != nil {
		t.Fatalf("Failed to decode JWE header: %v", err)
	}
	var jweHeader struct {
		Alg string `json:"alg"`
		Enc string `json:"enc"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(jweHeaderBytes, &jweHeader); err != nil {
		t.Fatalf("Failed to unmarshal JWE header: %v", err)
	}
	if jweHeader.Alg != "RSA-OAEP" {
		t.Errorf("Expected alg RSA-OAEP, got %s", jweHeader.Alg)
	}
	if jweHeader.Enc != "A256GCM" {
		t.Errorf("Expected enc A256GCM, got %s", jweHeader.Enc)
	}
	if jweHeader.Typ != "JWE" {
		t.Errorf("Expected typ JWE, got %s", jweHeader.Typ)
	}

	// Note: We can't decrypt the JWE token in the test because we don't have access to the private key
	// Instead, we'll verify that the other components are present and base64url encoded
	if _, err := base64.RawURLEncoding.DecodeString(jweParts[1]); err != nil {
		t.Errorf("Invalid encrypted key encoding: %v", err)
	}
	if _, err := base64.RawURLEncoding.DecodeString(jweParts[2]); err != nil {
		t.Errorf("Invalid IV encoding: %v", err)
	}
	if _, err := base64.RawURLEncoding.DecodeString(jweParts[3]); err != nil {
		t.Errorf("Invalid ciphertext encoding: %v", err)
	}
	if _, err := base64.RawURLEncoding.DecodeString(jweParts[4]); err != nil {
		t.Errorf("Invalid authentication tag encoding: %v", err)
	}
}

type mockRoundTripper struct {
	response []byte
	err      error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(m.response)),
	}, nil
}

func TestRefund(t *testing.T) {
	client := NewClient(Config{
		SecretKey:           "test_secret",
		MerchantID:          "JT01",
		BaseURL:             "https://demo2.2c2p.com/2C2PFrontend",
		PrivateKeyFile:      "testdata/combined_private_public.pem",
		ServerPublicKeyFile: "testdata/server.public_cert.pem",
	})

	// Mock successful response
	mockResp := `<?xml version="1.0" encoding="UTF-8"?>
	<PaymentProcessResponse>
		<version>4.3</version>
		<timeStamp>2021-01-26 08:53:27</timeStamp>
		<merchantID>JT01</merchantID>
		<invoiceNo>260121085327</invoiceNo>
		<actionAmount>25.00</actionAmount>
		<processType>R</processType>
		<respCode>0000</respCode>
		<respDesc>Success</respDesc>
		<approvalCode>123456</approvalCode>
		<referenceNo>REF123</referenceNo>
		<transactionID>T123</transactionID>
		<transactionRef>TREF123</transactionRef>
	</PaymentProcessResponse>`

	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			response: []byte(mockResp),
		},
	}
	client.httpClient = NewLoggingClient(mockClient, nil, false)

	resp, err := client.Refund(context.Background(), "260121085327", 25.00)
	if err != nil {
		t.Fatalf("Failed to process refund: %v", err)
	}

	// Verify response fields
	if resp.RespCode != "0000" {
		t.Errorf("Expected response code 0000, got %s", resp.RespCode)
	}
	if resp.RespDesc != "Success" {
		t.Errorf("Expected response description Success, got %s", resp.RespDesc)
	}
	if resp.ProcessType != "R" {
		t.Errorf("Expected process type R, got %s", resp.ProcessType)
	}
}
