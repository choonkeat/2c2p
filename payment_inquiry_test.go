package api2c2p

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/choonkeat/2c2p/testutil"
)

func TestPaymentInquiry(t *testing.T) {
	var ts *httptest.Server

	// Example request data from documentation
	request := &PaymentInquiryByInvoiceRequest{
		InvoiceNo: "254b77aabc",
		Locale:    "en",
	}

	// Example response data from documentation
	exampleResponse := PaymentInquiryResponse{
		MerchantID:                    "JT01",
		InvoiceNo:                     "1523953661",
		Amount:                        1000.00,
		CurrencyCode:                  "SGD",
		TransactionDateTime:           "311220235959",
		AgentCode:                     "OCBC",
		ChannelCode:                   "VI",
		ApprovalCode:                  "717282",
		ReferenceNo:                   "00010001",
		TranRef:                       "",
		AccountNo:                     "411111XXXXXX1111",
		CustomerToken:                 "",
		CustomerTokenExpiry:           "",
		CardType:                      "",
		IssuerCountry:                 "SG",
		IssuerBank:                    "",
		ECI:                           "05",
		InstallmentPeriod:             6,
		InterestType:                  "M",
		InterestRate:                  0.3,
		InstallmentMerchantAbsorbRate: 0.0,
		RecurringUniqueID:             "",
		RecurringSequenceNo:           0,
		FxAmount:                      25000.00,
		FxRate:                        25.0000001,
		FxCurrencyCode:                "THB",
		UserDefined1:                  "",
		UserDefined2:                  "",
		UserDefined3:                  "",
		UserDefined4:                  "",
		UserDefined5:                  "",
		AcquirerReferenceNo:           "",
		AcquirerMerchantID:            "",
		PaymentScheme:                 "",
		IdempotencyID:                 "",
		LoyaltyPoints:                 0,
		TransactionStatus:             "Success",
		MaskedPan:                     "411111XXXXXX1111",
		PaymentChannel:                "VI",
		PaymentStatus:                 "Success",
		ChannelResponseCode:           "00",
		ChannelResponseDescription:    "Success",
		PaidAgent:                     "OCBC",
		PaidChannel:                   "VI",
		PaidDateTime:                  "311220235959",
		RespCode:                      Flow1000LoadRedirectUrlWithIframeWebview,
		RespDesc:                      "Transaction is successful.",
	}

	// Create test server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Decode request body
		var reqBody struct {
			Payload string `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Error decoding request body: %v", err)
			return
		}

		// Create mock response
		responseData, err := json.Marshal(exampleResponse)
		if err != nil {
			t.Errorf("Error marshaling response: %v", err)
			return
		}

		// Create JWT token from response data
		mockClient := NewClient(Config{
			SecretKey:         "your_secret_key",
			MerchantID:        "JT01",
			PaymentGatewayURL: ts.URL,
			FrontendURL:       "https://frontend.example.com",
		})
		token, err := mockClient.generateJWTTokenForJSON(responseData)
		if err != nil {
			t.Errorf("Error generating JWT token: %v", err)
			return
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"payload": token,
		})
	}))
	defer ts.Close()

	// Create client with test server URL
	client := NewClient(Config{
		SecretKey:         "your_secret_key",
		MerchantID:        "JT01",
		PaymentGatewayURL: ts.URL,
		FrontendURL:       "https://frontend.example.com",
	})

	// Make request
	response, err := client.PaymentInquiryByInvoice(ctx, request)
	if err != nil {
		t.Fatalf("Error making payment inquiry: %v", err)
	}

	// Verify response fields
	if response.MerchantID != exampleResponse.MerchantID {
		t.Errorf("Expected MerchantID %s, got %s", exampleResponse.MerchantID, response.MerchantID)
	}
	if response.InvoiceNo != exampleResponse.InvoiceNo {
		t.Errorf("Expected InvoiceNo %s, got %s", exampleResponse.InvoiceNo, response.InvoiceNo)
	}
	if response.Amount != exampleResponse.Amount {
		t.Errorf("Expected Amount %.2f, got %.2f", exampleResponse.Amount, response.Amount)
	}
	if response.RespCode != exampleResponse.RespCode {
		t.Errorf("Expected RespCode %s, got %s", exampleResponse.RespCode, response.RespCode)
	}
	if response.RespDesc != exampleResponse.RespDesc {
		t.Errorf("Expected RespDesc %s, got %s", exampleResponse.RespDesc, response.RespDesc)
	}
}

func TestNewPaymentInquiryRequest(t *testing.T) {
	client := NewClient(Config{
		SecretKey:         "your_secret_key",
		MerchantID:        "JT01",
		PaymentGatewayURL: "https://pgw.example.com",
		FrontendURL:       "https://frontend.example.com",
	})
	req := &PaymentInquiryByInvoiceRequest{
		MerchantID: "JT01",
		InvoiceNo:  "254b77aabc",
		Locale:     "en",
	}

	// Create request
	httpReq, err := client.newPaymentInquiryRequest(ctx, req.MerchantID, req)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Verify request using testutil
	testutil.AssertRequest(t, httpReq, struct {
		Method      string
		URL         string
		ContentType string
		Headers     map[string]string
		Body        any
	}{
		Method:      "POST",
		URL:         "https://pgw.example.com/payment/4.3/paymentInquiry",
		ContentType: "application/json",
		Body: map[string]any{
			"merchantID": "JT01",
			"invoiceNo":  "254b77aabc",
			"locale":     "en",
		},
	})
}

func TestPaymentInquiryByToken(t *testing.T) {
	var ts *httptest.Server

	// Example request data
	request := &PaymentInquiryByTokenRequest{
		PaymentToken: "payment_token_123",
		Locale:       "en",
	}

	// Example response data
	exampleResponse := PaymentInquiryResponse{
		MerchantID:                    "JT01",
		InvoiceNo:                     "1523953661",
		Amount:                        1000.00,
		CurrencyCode:                  "SGD",
		TransactionDateTime:           "311220235959",
		AgentCode:                     "OCBC",
		ChannelCode:                   "VI",
		ApprovalCode:                  "717282",
		ReferenceNo:                   "00010001",
		TranRef:                       "",
		AccountNo:                     "411111XXXXXX1111",
		CustomerToken:                 "",
		CustomerTokenExpiry:           "",
		CardType:                      "",
		IssuerCountry:                 "SG",
		IssuerBank:                    "",
		ECI:                           "05",
		InstallmentPeriod:             6,
		InterestType:                  "M",
		InterestRate:                  0.3,
		InstallmentMerchantAbsorbRate: 0.0,
		RecurringUniqueID:             "",
		RecurringSequenceNo:           0,
		FxAmount:                      25000.00,
		FxRate:                        25.0000001,
		FxCurrencyCode:                "THB",
		UserDefined1:                  "",
		UserDefined2:                  "",
		UserDefined3:                  "",
		UserDefined4:                  "",
		UserDefined5:                  "",
		AcquirerReferenceNo:           "",
		AcquirerMerchantID:            "",
		PaymentScheme:                 "",
		IdempotencyID:                 "",
		LoyaltyPoints:                 0,
		TransactionStatus:             "Success",
		MaskedPan:                     "411111XXXXXX1111",
		PaymentChannel:                "VI",
		PaymentStatus:                 "Success",
		ChannelResponseCode:           "00",
		ChannelResponseDescription:    "Success",
		PaidAgent:                     "OCBC",
		PaidChannel:                   "VI",
		PaidDateTime:                  "311220235959",
		RespCode:                      Flow10051DisplayGeneratedQrAndWaitForCustomerToScan,
		RespDesc:                      "Transaction is successful.",
	}

	// Create test server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Decode request body
		var reqBody struct {
			Payload string `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Error decoding request body: %v", err)
			return
		}

		// Create mock response
		responseData, err := json.Marshal(exampleResponse)
		if err != nil {
			t.Errorf("Error marshaling response: %v", err)
			return
		}

		// Create JWT token from response data
		mockClient := NewClient(Config{
			SecretKey:         "your_secret_key",
			MerchantID:        "JT01",
			PaymentGatewayURL: ts.URL,
			FrontendURL:       "https://frontend.example.com",
		})
		token, err := mockClient.generateJWTTokenForJSON(responseData)
		if err != nil {
			t.Errorf("Error generating JWT token: %v", err)
			return
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"payload": token,
		})
	}))
	defer ts.Close()

	// Create client with test server URL
	client := NewClient(Config{
		SecretKey:         "your_secret_key",
		MerchantID:        "JT01",
		PaymentGatewayURL: ts.URL,
		FrontendURL:       "https://frontend.example.com",
	})

	// Make request
	response, err := client.PaymentInquiryByToken(ctx, request)
	if err != nil {
		t.Fatalf("Error making payment inquiry by token: %v", err)
	}

	// Verify response fields
	if response.MerchantID != exampleResponse.MerchantID {
		t.Errorf("Expected MerchantID %s, got %s", exampleResponse.MerchantID, response.MerchantID)
	}
	if response.InvoiceNo != exampleResponse.InvoiceNo {
		t.Errorf("Expected InvoiceNo %s, got %s", exampleResponse.InvoiceNo, response.InvoiceNo)
	}
	if response.Amount != exampleResponse.Amount {
		t.Errorf("Expected Amount %.2f, got %.2f", exampleResponse.Amount, response.Amount)
	}
	if response.RespCode != exampleResponse.RespCode {
		t.Errorf("Expected RespCode %s, got %s", exampleResponse.RespCode, response.RespCode)
	}
	if response.RespDesc != exampleResponse.RespDesc {
		t.Errorf("Expected RespDesc %s, got %s", exampleResponse.RespDesc, response.RespDesc)
	}
}

func TestNewPaymentInquiryByTokenRequest(t *testing.T) {
	client := NewClient(Config{
		SecretKey:         "your_secret_key",
		MerchantID:        "JT01",
		PaymentGatewayURL: "https://pgw.example.com",
		FrontendURL:       "https://frontend.example.com",
	})
	req := &PaymentInquiryByTokenRequest{
		MerchantID:   "JT01",
		PaymentToken: "payment_token_123",
		Locale:       "en",
	}

	// Create request
	httpReq, err := client.newPaymentInquiryRequest(ctx, req.MerchantID, req)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Verify request using testutil
	testutil.AssertRequest(t, httpReq, struct {
		Method      string
		URL         string
		ContentType string
		Headers     map[string]string
		Body        any
	}{
		Method:      "POST",
		URL:         "https://pgw.example.com/payment/4.3/paymentInquiry",
		ContentType: "application/json",
		Body: map[string]any{
			"merchantID":   "JT01",
			"paymentToken": "payment_token_123",
			"locale":       "en",
		},
	})
}
