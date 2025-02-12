package api2c2p

import (
	"context"
	"encoding/xml"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVoidCancel(t *testing.T) {
	var ts *httptest.Server
	var client *Client
	var err error

	// Example request data
	request := &VoidCancelRequest{
		InvoiceNo:    "test_invoice_123",
		ActionAmount: Cents(100.00).ToDollars(),
	}

	// Example response data
	exampleResponse := VoidCancelResponse{
		Version:        "3.8",
		TimeStamp:      "20250212090235",
		MerchantID:     "JT01",
		InvoiceNo:      request.InvoiceNo,
		ActionAmount:   "100.00",
		ProcessType:    "V",
		RespCode:       "0000",
		RespDesc:       "Success",
		ApprovalCode:   "123456",
		ReferenceNo:    "REF123",
		TransactionID:  "T123",
		TransactionRef: "TR123",
	}

	// Create test server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify content type
		if ct := r.Header.Get("Content-Type"); ct != "text/plain" {
			t.Errorf("Expected Content-Type text/plain, got %s", ct)
		}

		// Write response
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		resp, err := xml.Marshal(exampleResponse)
		if err != nil {
			t.Fatal(err)
		}

		signedJWE, err := client.encryptJWEAndSignJWS(resp)
		if err != nil {
			t.Fatalf("Failed to encrypt response: %v", err)
		}
		log.Printf("Signed token: %s", signedJWE)
		w.Write([]byte(signedJWE))
	}))
	defer ts.Close()

	// Create client with test server URL
	client, err = NewClient(Config{
		SecretKey:                "your_secret_key",
		MerchantID:               "JT01",
		PaymentGatewayURL:        "https://pgw.example.com",
		FrontendURL:              ts.URL,
		CombinedPEM:              "testdata/combined_private_public.pem",
		ServerJWTPublicKeyFile:   "testdata/public_cert.pem", // we have to decrypt what we encrypted in this test
		ServerPKCS7PublicKeyFile: "testdata/server.pkcs7.public_cert.pem",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test successful void/cancel
	resp, err := client.VoidCancel(context.Background(), request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify response fields
	if resp.RespCode != "0000" {
		t.Errorf("Expected response code 0000, got %s", resp.RespCode)
	}
	if resp.ProcessType != "V" {
		t.Errorf("Expected process type V, got %s", resp.ProcessType)
	}
	if resp.InvoiceNo != request.InvoiceNo {
		t.Errorf("Expected invoice number %s, got %s", request.InvoiceNo, resp.InvoiceNo)
	}

	// Test validation errors
	testCases := []struct {
		name    string
		req     *VoidCancelRequest
		wantErr string
	}{
		{
			name: "missing invoice number",
			req: &VoidCancelRequest{
				ActionAmount: Cents(100.00).ToDollars(),
			},
			wantErr: "invoice number is required",
		},
		{
			name: "invalid action amount",
			req: &VoidCancelRequest{
				InvoiceNo:    "test_invoice",
				ActionAmount: Cents(0).ToDollars(),
			},
			wantErr: "action amount must be greater than 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.VoidCancel(context.Background(), tc.req)
			if err == nil || err.Error() != tc.wantErr {
				t.Errorf("Expected error %q, got %v", tc.wantErr, err)
			}
		})
	}
}
