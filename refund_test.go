package api2c2p

import (
	"bytes"
	"context"
	"encoding/pem"
	"io"
	"net/http"
	"os"
	"testing"

	"crypto/x509"
	"encoding/xml"
)

func TestNewRefundRequest(t *testing.T) {
	client := NewClient(Config{
		SecretKey:           "test_secret",
		MerchantID:          "JT01",
		PaymentGatewayURL:   "https://pgw.example.com",
		FrontendURL:         "https://frontend.example.com",
		PrivateKeyFile:      "testdata/combined_private_public.pem",
		ServerPublicKeyFile: "testdata/server.public_cert.pem",
	})
	serverCombinedPrivatePublicKeyFile := "testdata/server.combined_private_public.pem"

	httpReq, err := client.NewRefundRequest(context.Background(), &PaymentProcessRequest{
		Version:      "4.3",
		MerchantID:   "JT01",
		InvoiceNo:    "260121085327",
		ActionAmount: Cents(2500).ToDollars(),
		ProcessType:  "R",
	})
	if err != nil {
		t.Fatalf("Failed to create refund request: %v", err)
	}

	// Read request body
	body, err := io.ReadAll(httpReq.Body)
	if err != nil {
		t.Fatalf("Failed to read request body: %v", err)
	}

	// Re-create request body for subsequent reads
	httpReq.Body = io.NopCloser(bytes.NewBuffer(body))

	// Load public key for JWS verification
	publicKeyPEM, err := os.ReadFile("testdata/public_cert.pem")
	if err != nil {
		t.Fatalf("Failed to read public key file: %v", err)
	}
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		t.Fatalf("Failed to decode public key PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Load private key for JWE decryption
	serverKeyData, err := os.ReadFile(serverCombinedPrivatePublicKeyFile)
	if err != nil {
		t.Fatalf("Failed to read server key file: %v", err)
	}
	block, _ = pem.Decode(serverKeyData)
	if block == nil {
		t.Fatalf("Failed to decode server key PEM")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	// Verify JWS and decrypt JWE
	decrypted, err := verifyAndDecryptJWSJWE(string(body), cert.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to verify and decrypt: %v", err)
	}

	// Parse XML payload
	var payload struct {
		XMLName      xml.Name `xml:"PaymentProcessRequest"`
		Version      string   `xml:"version"`
		MerchantID   string   `xml:"merchantID"`
		InvoiceNo    string   `xml:"invoiceNo"`
		ActionAmount string   `xml:"actionAmount"`
		ProcessType  string   `xml:"processType"`
	}
	if err := xml.Unmarshal(decrypted, &payload); err != nil {
		t.Fatalf("Failed to unmarshal decrypted payload: %v", err)
	}

	// Verify payload values
	if payload.Version != "4.3" {
		t.Errorf("Expected version 4.3, got %s", payload.Version)
	}
	if payload.MerchantID != "JT01" {
		t.Errorf("Expected merchantID JT01, got %s", payload.MerchantID)
	}
	if payload.InvoiceNo != "260121085327" {
		t.Errorf("Expected invoiceNo 260121085327, got %s", payload.InvoiceNo)
	}
	if payload.ActionAmount != "25.00" {
		t.Errorf("Expected actionAmount 25.00, got %s", payload.ActionAmount)
	}
	if payload.ProcessType != "R" {
		t.Errorf("Expected processType R, got %s", payload.ProcessType)
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
		PaymentGatewayURL:   "https://pgw.example.com",
		FrontendURL:         "https://frontend.example.com",
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
