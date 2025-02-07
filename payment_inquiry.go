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
	    "your_secret_key",
	    "your_merchant_id",
	    "https://sandbox-pgw.2c2p.com", // or https://pgw.2c2p.com for production
	)

	request := &api2c2p.PaymentInquiryRequest{
	    MerchantID:   "your_merchant_id",
	    InvoiceNo:    "your_invoice_number",  // Either InvoiceNo or PaymentToken is required
	    PaymentToken: "payment_token",        // Optional, alternative to InvoiceNo
	    Locale:       "en",                   // Optional
	}

	response, err := client.PaymentInquiry(ctx, request)
	if err != nil {
	    log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Payment status: %s - %s\n", response.RespCode, response.RespDesc)
*/
package api2c2p

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// PaymentInquiryRequest represents the request payload for payment inquiry
type PaymentInquiryRequest struct {
	MerchantID   string `json:"merchantID"`
	PaymentToken string `json:"paymentToken,omitempty"` // Either paymentToken or invoiceNo must be present
	InvoiceNo    string `json:"invoiceNo,omitempty"`    // Either paymentToken or invoiceNo must be present
	Locale       string `json:"locale,omitempty"`       // Based on ISO 639 codes
}

// PaymentInquiryResponse represents the decoded response from payment inquiry
type PaymentInquiryResponse struct {
	MerchantID                    string  `json:"merchantID"`                    // C 25, M
	InvoiceNo                     string  `json:"invoiceNo"`                     // AN 50, M
	Amount                        float64 `json:"amount"`                        // D (12,5), M
	CurrencyCode                  string  `json:"currencyCode"`                  // A 3, M
	TransactionDateTime           string  `json:"transactionDateTime"`           // N 14, M
	AgentCode                     string  `json:"agentCode"`                     // AN 30, M
	ChannelCode                   string  `json:"channelCode"`                   // AN 30, M
	ApprovalCode                  string  `json:"approvalCode"`                  // C 6, C
	ReferenceNo                   string  `json:"referenceNo"`                   // AN 50, M
	TranRef                       string  `json:"tranRef"`                       // AN 28, O
	AccountNo                     string  `json:"accountNo"`                     // N 19, M
	CustomerToken                 string  `json:"customerToken"`                 // AN 20, O
	CustomerTokenExpiry           string  `json:"customerTokenExpiry"`           // AN 8, O
	CardType                      string  `json:"cardType"`                      // C 20, C
	IssuerCountry                 string  `json:"issuerCountry"`                 // A 2, C
	IssuerBank                    string  `json:"issuerBank"`                    // C 200, C
	ECI                           string  `json:"eci"`                           // C 2, C
	InstallmentPeriod             int     `json:"installmentPeriod"`             // N 2, C
	InterestType                  string  `json:"interestType"`                  // A 1, C
	InterestRate                  float64 `json:"interestRate"`                  // D (3,5), C
	InstallmentMerchantAbsorbRate float64 `json:"installmentMerchantAbsorbRate"` // D (3,5), C
	RecurringUniqueID             string  `json:"recurringUniqueID"`             // N 20, C
	RecurringSequenceNo           int     `json:"recurringSequenceNo"`           // N 10, C
	FxAmount                      float64 `json:"fxAmount"`                      // D (12,5), C
	FxRate                        float64 `json:"fxRate"`                        // D (12,7), C
	FxCurrencyCode                string  `json:"fxCurrencyCode"`                // A 3, C
	UserDefined1                  string  `json:"userDefined1"`                  // C 150, O
	UserDefined2                  string  `json:"userDefined2"`                  // C 150, O
	UserDefined3                  string  `json:"userDefined3"`                  // C 150, O
	UserDefined4                  string  `json:"userDefined4"`                  // C 150, O
	UserDefined5                  string  `json:"userDefined5"`                  // C 150, O
	AcquirerReferenceNo           string  `json:"acquirerReferenceNo"`           // C 50, O
	AcquirerMerchantID            string  `json:"acquirerMerchantId"`            // C 50, O
	TransactionStatus             string  `json:"transactionStatus"`             // C 20, M
	MaskedPan                     string  `json:"maskedPan"`                     // C 19, C
	PaymentChannel                string  `json:"paymentChannel"`                // C 20, M
	PaymentStatus                 string  `json:"paymentStatus"`                 // C 20, M
	ChannelResponseCode           string  `json:"channelResponseCode"`           // C 20, C
	ChannelResponseDescription    string  `json:"channelResponseDescription"`    // C 255, C
	PaidAgent                     string  `json:"paidAgent"`                     // C 30, C
	PaidChannel                   string  `json:"paidChannel"`                   // C 30, C
	PaidDateTime                  string  `json:"paidDateTime"`                  // C 14, C
	RespCode                      string  `json:"respCode"`                      // C 4, M
	RespDesc                      string  `json:"respDesc"`                      // C 255, M
	LoyaltyPoints                 float64 `json:"loyaltyPoints,omitempty"`       // Type not specified in docs
	PaymentScheme                 string  `json:"paymentScheme"`                 // C 30, C
	IdempotencyID                 string  `json:"idempotencyID"`                 // C 100, O
}

func (c *Client) newPaymentInquiryRequest(ctx context.Context, req *PaymentInquiryRequest) (*http.Request, error) {
	url := c.endpoint("paymentInquiry")
	if req.MerchantID == "" {
		req.MerchantID = c.MerchantID
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Generate JWT token
	token, err := c.generateJWTToken(jsonData)
	if err != nil {
		return nil, fmt.Errorf("generate jwt token: %w", err)
	}

	// Prepare the request body
	requestBody := map[string]string{
		"payload": token,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w\nURL: %s", err, url)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	return httpReq, nil
}

// PaymentInquiry checks the status of a payment by invoice number
func (c *Client) PaymentInquiry(ctx context.Context, req *PaymentInquiryRequest) (*PaymentInquiryResponse, error) {
	// Create and make request
	httpReq, err := c.newPaymentInquiryRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// Make request with debug info
	resp, debug, err := c.doRequestWithDebug(httpReq)
	if err != nil {
		return nil, c.formatErrorWithDebug(fmt.Errorf("do request: %w", err), debug)
	}
	defer resp.Body.Close()
	log.Printf("Payment inquiry response body: %s", debug.Response.Body)

	// Try to decode response
	var jwtResponse struct {
		Payload string `json:"payload"`
	}
	if err := json.Unmarshal([]byte(debug.Response.Body), &jwtResponse); err != nil || jwtResponse.Payload == "" {
		// Try decoding as direct response
		var response struct {
			RespCode string `json:"respCode"`
			RespDesc string `json:"respDesc"`
		}
		if err := json.Unmarshal([]byte(debug.Response.Body), &response); err != nil {
			return nil, c.formatErrorWithDebug(fmt.Errorf("decode response: %w", err), debug)
		}
		return &PaymentInquiryResponse{
			RespCode: response.RespCode,
			RespDesc: response.RespDesc,
		}, nil
	}

	// If we got a JWT response, decode it
	var inquiryResp PaymentInquiryResponse
	if err := c.decodeJWTToken(jwtResponse.Payload, &inquiryResp); err != nil {
		return nil, c.formatErrorWithDebug(fmt.Errorf("decode jwt token: %w", err), debug)
	}

	// Check response code
	switch inquiryResp.RespCode {
	case "0000", "0001", "1005", "2001":
		return &inquiryResp, nil
	}
	return &inquiryResp, c.formatErrorWithDebug(fmt.Errorf("payment inquiry failed: %s (%s)", inquiryResp.RespCode, inquiryResp.RespDesc), debug)
}
