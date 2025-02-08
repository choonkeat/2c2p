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

	tokenRequest := &api2c2p.PaymentInquiryByTokenRequest{
	    MerchantID:   "your_merchant_id",
	    PaymentToken: "payment_token",
	    Locale:       "en", // Optional
	}

	invoiceRequest := &api2c2p.PaymentInquiryByInvoiceRequest{
	    MerchantID: "your_merchant_id",
	    InvoiceNo:  "your_invoice_number",
	    Locale:     "en", // Optional
	}

	tokenResponse, err := client.PaymentInquiryByToken(ctx, tokenRequest)
	if err != nil {
	    fmt.Printf("Error: %v\n", err)
	}

	invoiceResponse, err := client.PaymentInquiryByInvoice(ctx, invoiceRequest)
	if err != nil {
	    fmt.Printf("Error: %v\n", err)
	}

	fmt.Printf("Payment status by token: %s - %s\n", tokenResponse.RespCode, tokenResponse.RespDesc)
	fmt.Printf("Payment status by invoice: %s - %s\n", invoiceResponse.RespCode, invoiceResponse.RespDesc)
*/
package api2c2p

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// PaymentInquiryByTokenRequest represents the request payload for payment inquiry by payment token
// Documentation: https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry-request-parameter
type PaymentInquiryByTokenRequest struct {
	// Locale is the language code for the response (Optional)
	// Based on ISO 639. If not set, API response will be based on payment token locale
	Locale string `json:"locale,omitempty"`

	// MerchantID is the 2C2P merchant ID (Required)
	// Max length: 8 characters
	MerchantID string `json:"merchantID"`

	// PaymentToken is the payment token to query (Required)
	PaymentToken string `json:"paymentToken"` // Required
}

// PaymentInquiryByInvoiceRequest represents the request payload for payment inquiry by invoice number
// Documentation: https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry-request-parameter
type PaymentInquiryByInvoiceRequest struct {
	// InvoiceNo is the invoice number to query (Required)
	// Max length: 50 characters
	InvoiceNo string `json:"invoiceNo"` // Required

	// Locale is the language code for the response (Optional)
	// Based on ISO 639. If not set, API response will be based on payment token locale
	Locale string `json:"locale,omitempty"`

	// MerchantID is the 2C2P merchant ID (Required)
	// Max length: 8 characters
	MerchantID string `json:"merchantID"`
}

// PaymentInquiryResponse represents the decoded response from payment inquiry
// Documentation: https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry-response-parameter
type PaymentInquiryResponse struct {
	// MerchantID is the 2C2P merchant ID (C 25, M)
	MerchantID string `json:"merchantID"`

	// InvoiceNo is the invoice number (AN 50, M)
	InvoiceNo string `json:"invoiceNo"`

	// Amount is the transaction amount (D 12.5, M)
	Amount float64 `json:"amount"`

	// CurrencyCode is the transaction currency code (A 3, M)
	// Based on ISO 4217
	CurrencyCode string `json:"currencyCode"`

	// TransactionDateTime is the transaction date time (N 14, M)
	// Format: yyyyMMddHHmmss
	TransactionDateTime string `json:"transactionDateTime"`

	// AgentCode is the agent code (AN 6, C)
	AgentCode string `json:"agentCode,omitempty"`

	// ChannelCode is the payment channel code (AN 6, M)
	ChannelCode string `json:"channelCode"`

	// ReferenceNo is the transaction reference number (AN 12, M)
	ReferenceNo string `json:"referenceNo"`

	// TranRef is the transaction reference number (AN 20, M)
	TranRef string `json:"tranRef"`

	// RespCode is the response code (C 4, M)
	RespCode PaymentFlowResponseCode `json:"respCode"`

	// RespDesc is the response description (C 255, M)
	RespDesc string `json:"respDesc"`

	// ApprovalCode is the approval code (C 6, C)
	ApprovalCode string `json:"approvalCode"`

	// AccountNo is the account number (N 19, M)
	AccountNo string `json:"accountNo"`

	// CustomerToken is the customer token (AN 20, O)
	CustomerToken string `json:"customerToken"`

	// CustomerTokenExpiry is the customer token expiry (AN 8, O)
	CustomerTokenExpiry string `json:"customerTokenExpiry"`

	// CardType is the card type (C 20, C)
	CardType string `json:"cardType"`

	// IssuerCountry is the issuer country (A 2, C)
	IssuerCountry string `json:"issuerCountry"`

	// IssuerBank is the issuer bank (C 200, C)
	IssuerBank string `json:"issuerBank"`

	// ECI is the electronic commerce indicator (C 2, C)
	ECI string `json:"eci"`

	// InstallmentPeriod is the installment period (N 2, C)
	InstallmentPeriod int `json:"installmentPeriod"`

	// InterestType is the interest type (A 1, C)
	InterestType string `json:"interestType"`

	// InterestRate is the interest rate (D 3.5, C)
	InterestRate float64 `json:"interestRate"`

	// InstallmentMerchantAbsorbRate is the installment merchant absorb rate (D 3.5, C)
	InstallmentMerchantAbsorbRate float64 `json:"installmentMerchantAbsorbRate"`

	// RecurringUniqueID is the recurring unique ID (N 20, C)
	RecurringUniqueID string `json:"recurringUniqueID"`

	// RecurringSequenceNo is the recurring sequence number (N 10, C)
	RecurringSequenceNo int `json:"recurringSequenceNo"`

	// FxAmount is the foreign exchange amount (D 12.5, C)
	FxAmount float64 `json:"fxAmount"`

	// FxRate is the foreign exchange rate (D 12.7, C)
	FxRate float64 `json:"fxRate"`

	// FxCurrencyCode is the foreign exchange currency code (A 3, C)
	FxCurrencyCode string `json:"fxCurrencyCode"`

	// UserDefined1 is the user defined 1 (C 150, O)
	UserDefined1 string `json:"userDefined1"`

	// UserDefined2 is the user defined 2 (C 150, O)
	UserDefined2 string `json:"userDefined2"`

	// UserDefined3 is the user defined 3 (C 150, O)
	UserDefined3 string `json:"userDefined3"`

	// UserDefined4 is the user defined 4 (C 150, O)
	UserDefined4 string `json:"userDefined4"`

	// UserDefined5 is the user defined 5 (C 150, O)
	UserDefined5 string `json:"userDefined5"`

	// AcquirerReferenceNo is the acquirer reference number (C 50, O)
	AcquirerReferenceNo string `json:"acquirerReferenceNo"`

	// AcquirerMerchantID is the acquirer merchant ID (C 50, O)
	AcquirerMerchantID string `json:"acquirerMerchantId"`

	// TransactionStatus is the transaction status (C 20, M)
	TransactionStatus string `json:"transactionStatus"`

	// MaskedPan is the masked PAN (C 19, C)
	MaskedPan string `json:"maskedPan"`

	// PaymentChannel is the payment channel (C 20, M)
	PaymentChannel string `json:"paymentChannel"`

	// PaymentStatus is the payment status (C 20, M)
	PaymentStatus string `json:"paymentStatus"`

	// ChannelResponseCode is the channel response code (C 20, C)
	ChannelResponseCode string `json:"channelResponseCode"`

	// ChannelResponseDescription is the channel response description (C 255, C)
	ChannelResponseDescription string `json:"channelResponseDescription"`

	// PaidAgent is the paid agent (C 30, C)
	PaidAgent string `json:"paidAgent"`

	// PaidChannel is the paid channel (C 30, C)
	PaidChannel string `json:"paidChannel"`

	// PaidDateTime is the paid date time (C 14, C)
	PaidDateTime string `json:"paidDateTime"`

	// LoyaltyPoints is the loyalty points (D 12.5, O)
	LoyaltyPoints float64 `json:"loyaltyPoints,omitempty"`

	// PaymentScheme is the payment scheme (C 30, C)
	PaymentScheme string `json:"paymentScheme"`

	// IdempotencyID is the idempotency ID (C 100, O)
	IdempotencyID string `json:"idempotencyID"`
}

// IsSuccess returns true if the response code indicates success
func (r *PaymentInquiryResponse) IsSuccess() bool {
	switch r.RespCode {
	case FlowOtherTransactionFailedOrRejectedPerformPaymentInquiryToGetPayment:
		return false
	default:
		return true
	}
}

func (c *Client) newPaymentInquiryRequest(ctx context.Context, merchantID string, payload interface{}) (*http.Request, error) {
	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payload: %v", err)
	}

	// Create JWT token
	token, err := c.generateJWTToken(payloadBytes)
	if err != nil {
		return nil, fmt.Errorf("error generating JWT token: %v", err)
	}

	// Create request body
	requestBody := map[string]string{
		"payload": token,
	}
	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %v", err)
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint("paymentInquiry"), bytes.NewReader(requestBytes))
	if err != nil {
		return nil, fmt.Errorf("create payment inquiry request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// PaymentInquiryByToken checks the status of a payment using a payment token
func (c *Client) PaymentInquiryByToken(ctx context.Context, req *PaymentInquiryByTokenRequest) (*PaymentInquiryResponse, error) {
	if req.PaymentToken == "" {
		return nil, fmt.Errorf("payment token is required")
	}
	if req.MerchantID == "" {
		req.MerchantID = c.MerchantID
	}

	httpReq, err := c.newPaymentInquiryRequest(ctx, req.MerchantID, req)
	if err != nil {
		return nil, err
	}

	// Make request
	resp, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Try to decode response
	var jwtResponse struct {
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwtResponse); err != nil || jwtResponse.Payload == "" {
		// Try decoding as direct response
		var response struct {
			RespCode PaymentFlowResponseCode `json:"respCode"`
			RespDesc string                  `json:"respDesc"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &PaymentInquiryResponse{
			RespCode: response.RespCode,
			RespDesc: response.RespDesc,
		}, nil
	}

	// If we got a JWT response, decode it
	var inquiryResp PaymentInquiryResponse
	if err := c.decodeJWTToken(jwtResponse.Payload, &inquiryResp); err != nil {
		return nil, fmt.Errorf("decode jwt token: %w", err)
	}

	// Check response code
	if inquiryResp.IsSuccess() {
		return &inquiryResp, nil
	}
	return &inquiryResp, fmt.Errorf("payment inquiry failed: %s (%s)", inquiryResp.RespCode, inquiryResp.RespDesc)
}

// PaymentInquiryByInvoice checks the status of a payment using an invoice number
func (c *Client) PaymentInquiryByInvoice(ctx context.Context, req *PaymentInquiryByInvoiceRequest) (*PaymentInquiryResponse, error) {
	if req.InvoiceNo == "" {
		return nil, fmt.Errorf("invoice number is required")
	}
	if req.MerchantID == "" {
		req.MerchantID = c.MerchantID
	}

	httpReq, err := c.newPaymentInquiryRequest(ctx, req.MerchantID, req)
	if err != nil {
		return nil, err
	}

	// Make request
	resp, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Try to decode response
	var jwtResponse struct {
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwtResponse); err != nil || jwtResponse.Payload == "" {
		// Try decoding as direct response
		var response struct {
			RespCode PaymentFlowResponseCode `json:"respCode"`
			RespDesc string                  `json:"respDesc"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &PaymentInquiryResponse{
			RespCode: response.RespCode,
			RespDesc: response.RespDesc,
		}, nil
	}

	// If we got a JWT response, decode it
	var inquiryResp PaymentInquiryResponse
	if err := c.decodeJWTToken(jwtResponse.Payload, &inquiryResp); err != nil {
		return nil, fmt.Errorf("decode jwt token: %w", err)
	}

	// Check response code
	if inquiryResp.IsSuccess() {
		return &inquiryResp, nil
	}
	return &inquiryResp, fmt.Errorf("payment inquiry failed: %s (%s)", inquiryResp.RespCode, inquiryResp.RespDesc)
}
