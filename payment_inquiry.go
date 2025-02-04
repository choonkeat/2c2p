/*
Package api2c2p implements a Go client for the 2C2P Payment Gateway API v4.3.1.

2C2P API Documentation:
  - API Overview: https://developer.2c2p.com/v4.3.1/docs
  - Payment Inquiry API: https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry
  - Request Parameters: https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry-request-parameter
  - Response Parameters: https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry-response-parameter
  - Response Codes: https://developer.2c2p.com/v4.3.1/docs/response-code-payment
  - JWT Token Guide: https://developer.2c2p.com/v4.3.1/docs/json-web-tokens-jwt

Example Usage:
    client := api2c2p.NewClient(
        "your_merchant_id",
        "your_secret_key",
        "https://sandbox-pgw.2c2p.com", // or https://pgw.2c2p.com for production
    )

    request := &api2c2p.PaymentInquiryRequest{
        MerchantID:   "your_merchant_id",
        InvoiceNo:    "your_invoice_number",  // Either InvoiceNo or PaymentToken is required
        PaymentToken: "payment_token",        // Optional, alternative to InvoiceNo
        Locale:       "en",                   // Optional
    }

    response, err := client.PaymentInquiry(request)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }

    fmt.Printf("Payment status: %s - %s\n", response.RespCode, response.RespDesc)
*/
package api2c2p

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// PaymentInquiryRequest represents the request payload for payment inquiry
type PaymentInquiryRequest struct {
	MerchantID   string `json:"merchantID"`
	PaymentToken string `json:"paymentToken,omitempty"` // Either paymentToken or invoiceNo must be present
	InvoiceNo    string `json:"invoiceNo,omitempty"`   // Either paymentToken or invoiceNo must be present
	Locale       string `json:"locale,omitempty"`       // Based on ISO 639 codes
}

// PaymentInquiryResponse represents the decoded response from payment inquiry
type PaymentInquiryResponse struct {
	MerchantID                     string  `json:"merchantID"`                      // C 25, M
	InvoiceNo                      string  `json:"invoiceNo"`                       // AN 50, M
	Amount                         float64 `json:"amount"`                          // D (12,5), M
	CurrencyCode                   string  `json:"currencyCode"`                    // A 3, M
	TransactionDateTime            string  `json:"transactionDateTime"`             // N 14, M
	AgentCode                      string  `json:"agentCode"`                       // AN 30, M
	ChannelCode                    string  `json:"channelCode"`                     // AN 30, M
	ApprovalCode                   string  `json:"approvalCode"`                    // C 6, C
	ReferenceNo                    string  `json:"referenceNo"`                     // AN 50, M
	TranRef                       string  `json:"tranRef"`                         // AN 28, O
	AccountNo                      string  `json:"accountNo"`                       // N 19, M
	CustomerToken                  string  `json:"customerToken"`                   // AN 20, O
	CustomerTokenExpiry           string  `json:"customerTokenExpiry"`             // AN 8, O
	CardType                      string  `json:"cardType"`                        // C 20, C
	IssuerCountry                  string  `json:"issuerCountry"`                   // A 2, C
	IssuerBank                    string  `json:"issuerBank"`                      // C 200, C
	ECI                           string  `json:"eci"`                             // C 2, C
	InstallmentPeriod             int     `json:"installmentPeriod"`               // N 2, C
	InterestType                  string  `json:"interestType"`                    // A 1, C
	InterestRate                  float64 `json:"interestRate"`                    // D (3,5), C
	InstallmentMerchantAbsorbRate float64 `json:"installmentMerchantAbsorbRate"`   // D (3,5), C
	RecurringUniqueID            string  `json:"recurringUniqueID"`               // N 20, C
	RecurringSequenceNo          int     `json:"recurringSequenceNo"`             // N 10, C
	FxAmount                      float64 `json:"fxAmount"`                         // D (12,5), C
	FxRate                        float64 `json:"fxRate"`                           // D (12,7), C
	FxCurrencyCode               string  `json:"fxCurrencyCode"`                  // A 3, C
	UserDefined1                  string  `json:"userDefined1"`                     // C 150, O
	UserDefined2                  string  `json:"userDefined2"`                     // C 150, O
	UserDefined3                  string  `json:"userDefined3"`                     // C 150, O
	UserDefined4                  string  `json:"userDefined4"`                     // C 150, O
	UserDefined5                  string  `json:"userDefined5"`                     // C 150, O
	AcquirerReferenceNo          string  `json:"acquirerReferenceNo"`             // C 50, O
	AcquirerMerchantID          string  `json:"acquirerMerchantId"`              // C 50, O
	PaymentScheme                string  `json:"paymentScheme"`                    // C 30, C
	IdempotencyID               string  `json:"idempotencyID"`                    // C 100, O
	LoyaltyPoints               float64 `json:"loyaltyPoints,omitempty"`          // Type not specified in docs
	RespCode                     string  `json:"respCode"`                         // C 4, M
	RespDesc                     string  `json:"respDesc"`                         // C 255, M
}

// PaymentInquiry makes a payment inquiry request to 2C2P
func (c *Client) PaymentInquiry(request *PaymentInquiryRequest) (*PaymentInquiryResponse, error) {
	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Generate JWT token
	payload, err := c.GenerateJWTToken(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error generating JWT token: %v", err)
	}

	// Prepare the request body
	requestBody := map[string]string{
		"payload": payload,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %v", err)
	}

	// Create HTTP request
	url := c.BaseURL + "/payment/4.3/paymentInquiry"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	var jwtResponse struct {
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwtResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Decode JWT token
	var response PaymentInquiryResponse
	if err := c.DecodeJWTToken(jwtResponse.Payload, &response); err != nil {
		return nil, fmt.Errorf("error decoding JWT token: %v", err)
	}

	return &response, nil
}
