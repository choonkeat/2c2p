package api2c2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DoPaymentResponse represents a response from the QR payment API
// Documentation: https://developer.2c2p.com/v4.3.1/docs/api-do-payment-response-parameter
type DoPaymentResponse struct {
	// PaymentToken is the token from the payment token request
	PaymentToken string `json:"paymentToken"`

	// ChannelCode is the payment channel code (AN 6, M)
	ChannelCode string `json:"channelCode"`

	// InvoiceNo is the unique merchant order number (AN 50, C)
	// Only returned when respCode is 2000 (transaction completed)
	InvoiceNo string `json:"invoiceNo"`

	// Type is the data type (A 6, C)
	// For QR = QR data type
	Type string `json:"type"`

	// ExpiryTimer is the payment expiry timer in milliseconds (N 10, C)
	// For payment flow 1005 only
	ExpiryTimer string `json:"expiryTimer"`

	// ExpiryDescription is the payment expiry description (C 255, C)
	// For payment flow 1005 only
	ExpiryDescription string `json:"expiryDescription"`

	// Data is the data of the payment flow (C 5000, C)
	// 1. URL Endpoint / Deeplink - Required redirect to the endpoint
	// 2. QR Code - Required display the qr code
	Data string `json:"data"`

	// FallbackData is for payment flow response code 1004 only (C 255, C)
	// If user device doesn't have specific native app installed, this fallback allows payment via web
	FallbackData string `json:"fallbackData"`

	// RespCode is the response code (C 4, M)
	RespCode string `json:"respCode"`

	// RespDesc is the response description (C 255, M)
	RespDesc string `json:"respDesc"`
}

// DoPaymentRequest represents a request to do a QR payment
// Documentation: https://developer.2c2p.com/v4.3.1/docs/api-do-payment
type DoPaymentRequest struct {
	// PaymentToken is the token from the payment token request (Required)
	PaymentToken string `json:"paymentToken"`

	// ClientID is a unique identifier for this request (Optional)
	ClientID string `json:"clientID,omitempty"`

	// Locale is the language code for the response (Optional)
	Locale string `json:"locale,omitempty"`

	// ResponseReturnUrl is the URL to return to after payment (Required)
	ResponseReturnUrl string `json:"responseReturnUrl"`

	// Payment contains the payment details
	Payment struct {
		Code struct {
			ChannelCode string `json:"channelCode"`
			AgentCode   string `json:"agentCode,omitempty"`
		} `json:"code"`
		Data map[string]string `json:"data,omitempty"`
	} `json:"payment"`
}

// PaymentOptionResponse represents a response from the payment option API
// Documentation: https://developer.2c2p.com/v4.3.1/docs/api-payment-option-response-parameter
type PaymentOptionResponse struct {
	// PaymentToken is the payment token ID (C 255, M)
	PaymentToken string `json:"paymentToken"`

	// RespCode is the response code (N 4, M)
	RespCode string `json:"respCode"`

	// RespDesc is the response description (C 255, M)
	RespDesc string `json:"respDesc"`
}

// PaymentOptionRequest represents a request to get available payment options
// Documentation: https://developer.2c2p.com/v4.3.1/docs/api-payment-option-request-parameter
type PaymentOptionRequest struct {
	// PaymentToken is the token from the payment token request (Mandatory)
	PaymentToken string `json:"paymentToken"`

	// ClientID is a unique identifier for this request (Optional)
	// This ID will be created when UI SDK init and store in app preference
	ClientID string `json:"clientID,omitempty"`

	// Locale is the language code for the response (Optional)
	// Based on ISO 639. If not set, API response will be based on payment token locale
	// If clientID is set, API response will be based on user preference locale
	Locale string `json:"locale,omitempty"`
}

// PaymentOptionDetailsRequest represents a request to get payment option details
// Documentation: https://developer.2c2p.com/v4.3.1/docs/api-payment-option-details-request-parameter
type PaymentOptionDetailsRequest struct {
	// PaymentToken is the token from the payment token request (Mandatory)
	PaymentToken string `json:"paymentToken"`

	// ClientID is a unique identifier for this request (Optional)
	// This ID will be created when UI SDK init and store in app preference
	ClientID string `json:"clientID,omitempty"`

	// Locale is the language code for the response (Optional)
	// Based on ISO 639. If not set, API response will be based on payment token locale
	// If clientID is set, API response will be based on user preference locale
	Locale string `json:"locale,omitempty"`

	// CategoryCode is the payment category code (Mandatory, AN 6)
	// Get from Payment Options API
	// Only support payment channel which required payment details
	CategoryCode string `json:"categoryCode"`

	// GroupCode is the payment group code (Mandatory, AN 10)
	// Get from Payment Options API
	GroupCode string `json:"groupCode"`
}

// APIResponse represents a response from the payment option details API
type APIResponse struct {
	Payload  string `json:"payload"`
	RespCode string `json:"respCode"`
	RespDesc string `json:"respDesc"`
}

// DoPaymentParams represents parameters for creating a new do payment request
type DoPaymentParams struct {
	// PaymentToken is the token from the payment token request
	PaymentToken string

	// PaymentChannelCode is the payment channel code (e.g., SGQR)
	PaymentChannelCode string

	// PaymentData contains additional payment data
	PaymentData map[string]any

	// Locale is the language code for the response
	Locale string

	// ResponseReturnUrl is the URL to return to after payment
	ResponseReturnUrl string

	// ClientID is a unique identifier for this request
	ClientID string

	// ClientIP is the IP address of the client
	ClientIP string
}

// CreateQRPaymentParams represents parameters for creating a new QR payment
type CreateQRPaymentParams struct {
	// PaymentToken is the token from the payment token request
	PaymentToken string

	// PaymentChannelCode is the payment channel code (e.g., SGQR)
	PaymentChannelCode string

	// ResponseReturnUrl is the URL to return to after payment
	ResponseReturnUrl string

	// ClientIP is the IP address of the client
	ClientIP string
}

func (c *Client) newPaymentOptionsRequest(ctx context.Context, paymentToken string) (*http.Request, error) {
	paymentOptionURL := c.endpoint("paymentOption")

	// Prepare payment option payload
	paymentOptionPayload := &PaymentOptionRequest{
		PaymentToken: paymentToken,
		// Locale is optional, omitting it to use payment token locale
	}
	paymentOptionData, err := json.Marshal(paymentOptionPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payment option request: %v", err)
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", paymentOptionURL, strings.NewReader(string(paymentOptionData)))
	if err != nil {
		return nil, fmt.Errorf("create payment option request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *Client) newPaymentOptionDetailsRequest(ctx context.Context, paymentToken string) (*http.Request, error) {
	paymentOptionDetailsURL := c.endpoint("paymentOptionDetails")

	// Prepare payment option details payload
	paymentOptionDetailsPayload := &PaymentOptionDetailsRequest{
		PaymentToken: paymentToken,
		CategoryCode: "QR",   // For QR payments
		GroupCode:    "SGQR", // For QR payments
	}
	paymentOptionDetailsData, err := json.Marshal(paymentOptionDetailsPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payment option details request: %v", err)
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", paymentOptionDetailsURL, strings.NewReader(string(paymentOptionDetailsData)))
	if err != nil {
		return nil, fmt.Errorf("create payment option details request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *Client) newDoPaymentRequest(ctx context.Context, params *DoPaymentParams) (*http.Request, error) {
	doPaymentURL := c.endpoint("payment")

	// Prepare do payment payload using map for easier iteration
	doPaymentPayload := map[string]any{
		"locale":            params.Locale,
		"paymentToken":      params.PaymentToken,
		"responseReturnUrl": params.ResponseReturnUrl,
	}

	// Add optional fields if provided
	if params.ClientID != "" {
		doPaymentPayload["clientID"] = params.ClientID
	}
	if params.ClientIP != "" {
		doPaymentPayload["clientIP"] = params.ClientIP
	}

	// Add payment details
	doPaymentPayload["payment"] = map[string]any{
		"code": map[string]string{
			"channelCode": params.PaymentChannelCode,
		},
		"data": params.PaymentData,
	}

	doPaymentData, err := json.Marshal(doPaymentPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal do payment request: %w", err)
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", doPaymentURL, strings.NewReader(string(doPaymentData)))
	if err != nil {
		return nil, fmt.Errorf("create do payment request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// CreateQRPayment creates a new QR payment
func (c *Client) CreateQRPayment(ctx context.Context, params *CreateQRPaymentParams) (*DoPaymentResponse, error) {
	doPaymentParams := &DoPaymentParams{
		PaymentToken:       params.PaymentToken,
		PaymentChannelCode: params.PaymentChannelCode,
		PaymentData: map[string]any{
			"qrType": "URL",
		},
		Locale:            "en",
		ResponseReturnUrl: params.ResponseReturnUrl,
		ClientIP:          params.ClientIP,
	}

	// Create request
	req, err := c.newDoPaymentRequest(ctx, doPaymentParams)
	if err != nil {
		return nil, err
	}

	// Call do payment API
	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("do payment request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read do payment response body: %w", err)
	}

	// Parse response
	var doPaymentRespData DoPaymentResponse
	if err := json.Unmarshal(respBody, &doPaymentRespData); err != nil {
		return nil, fmt.Errorf("unmarshal do payment response: %w", err)
	}

	return &doPaymentRespData, nil
}
