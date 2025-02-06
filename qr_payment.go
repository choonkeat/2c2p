package api2c2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// QRPaymentRequest represents a request to create a QR payment
type QRPaymentRequest struct {
	// MerchantID is the merchant's unique identifier
	MerchantID string `json:"merchantID"`

	// InvoiceNo is the merchant's order number
	InvoiceNo string `json:"invoiceNo"`

	// Description is the payment description
	Description string `json:"description"`

	// Amount is the payment amount in smallest currency unit (e.g., cents)
	Amount float64 `json:"amount"`

	// CurrencyCodeISO4217 is the payment currency code (e.g., SGD, USD)
	CurrencyCodeISO4217 string `json:"currencyCode"`

	// PaymentChannel specifies the payment channel(s) to use
	PaymentChannel []string `json:"paymentChannel"`

	// ServerURL is your server's URL for handling callbacks
	ServerURL string `json:"-"`
}

// QRPaymentResponse represents a response from creating a QR payment
type QRPaymentResponse struct {
	// PaymentToken is the token used for this payment
	PaymentToken string `json:"paymentToken"`
	// Type is the QR code type (URL)
	Type string `json:"type"`

	// ExpiryTimer is the expiry time in milliseconds
	ExpiryTimer string `json:"expiryTimer"`

	// ExpiryDescription is the expiry description template
	ExpiryDescription string `json:"expiryDescription"`

	// Data is the QR code data (URL to QR code image)
	Data string `json:"data"`

	// ChannelCode is the payment channel code (PNQR)
	ChannelCode string `json:"channelCode"`

	// RespCode is the response code
	// 1005: Pending for user scan QR (success)
	RespCode string `json:"respCode"`

	// RespDesc is the response description
	RespDesc string `json:"respDesc"`
}

// DoPaymentRequest represents a request to do a QR payment
type DoPaymentRequest struct {
	// PaymentToken is the token from the payment token request
	PaymentToken string `json:"paymentToken"`

	// Locale is the language code for the response
	Locale string `json:"locale"`

	// ResponseReturnUrl is the URL to return to after payment
	ResponseReturnUrl string `json:"responseReturnUrl"`

	// Payment contains the payment details
	Payment struct {
		// Code contains the payment channel code
		Code struct {
			// ChannelCode is the payment channel code (e.g., SGQR)
			ChannelCode string `json:"channelCode"`
		} `json:"code"`

		// Data contains additional payment data
		Data struct {
			// Name is the customer's name
			Name string `json:"name"`

			// Email is the customer's email
			Email string `json:"email"`

			// // QRType is the QR data type (RAW, BASE64, or URL)
			// QRType string `json:"qrType"`
		} `json:"data"`
	} `json:"payment"`

	// ClientIP is the IP address of the client
	ClientIP string `json:"clientIP,omitempty"`

	// PaymentExpiry is the payment expiry date/time in yyyy-MM-dd HH:mm:ss format
	PaymentExpiry string `json:"paymentExpiry,omitempty"`

	// Request3DS is the request 3DS flag
	Request3DS string `json:"request3DS,omitempty"`

	// FrontendReturnUrl is the frontend return URL
	FrontendReturnUrl string `json:"frontendReturnUrl,omitempty"`
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

// PaymentOptionResponse represents a response from the payment option API
type PaymentOptionResponse struct {
	Payload  string `json:"payload"`
	RespCode string `json:"respCode"`
	RespDesc string `json:"respDesc"`
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
	PaymentToken       string
	PaymentChannelCode string
	PaymentData        map[string]any
	Locale             string
	ResponseReturnUrl  string
	ClientID           string
	ClientIP           string
}

type CreateQRPaymentParams struct {
	PaymentToken       string
	PaymentChannelCode string
	ResponseReturnUrl  string
	ClientIP           string
}

func (c *Client) newPaymentOptionsRequest(ctx context.Context, paymentToken string) (*http.Request, error) {
	paymentOptionURL := c.endpoint("paymentOption")
	log.Printf("Getting payment options at %s", paymentOptionURL)

	// Prepare payment option payload
	paymentOptionPayload := &PaymentOptionRequest{
		PaymentToken: paymentToken,
		// Locale is optional, omitting it to use payment token locale
	}
	paymentOptionData, err := json.Marshal(paymentOptionPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payment option request: %v", err)
	}
	log.Printf("Payment option request payload (before JWT): %s", string(paymentOptionData))

	paymentOptionReqData, err := json.Marshal(paymentOptionPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payment option request with payload: %v", err)
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", paymentOptionURL, strings.NewReader(string(paymentOptionReqData)))
	if err != nil {
		return nil, fmt.Errorf("create payment option request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// getPaymentOptions calls the payment option API to get available payment options
func (c *Client) getPaymentOptions(ctx context.Context, paymentToken string) (*PaymentOptionResponse, error) {
	req, err := c.newPaymentOptionsRequest(ctx, paymentToken)
	if err != nil {
		return nil, err
	}

	// Call payment option API with debug info
	resp, debug, err := c.doRequestWithDebug(req)
	if err != nil {
		return nil, c.formatErrorWithDebug(fmt.Errorf("payment option request: %w", err), debug)
	}
	defer resp.Body.Close()

	// Read payment option response
	paymentOptionRespBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading payment option response: %v", err)
	}
	log.Printf("Payment option response body: %s", string(paymentOptionRespBody))

	// Parse payment option response
	var paymentOptionRespData PaymentOptionResponse
	if err := json.Unmarshal(paymentOptionRespBody, &paymentOptionRespData); err != nil {
		return nil, c.formatErrorWithDebug(fmt.Errorf("decode payment option response: %w", err), debug)
	}

	return &paymentOptionRespData, nil
}

func (c *Client) newPaymentOptionDetailsRequest(ctx context.Context, paymentToken string) (*http.Request, error) {
	paymentOptionDetailsURL := c.endpoint("paymentOptionDetails")
	log.Printf("Getting payment option details at %s", paymentOptionDetailsURL)

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
	log.Printf("Payment option details request payload (before JWT): %s", string(paymentOptionDetailsData))

	paymentOptionDetailsReqData, err := json.Marshal(paymentOptionDetailsPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payment option details request with payload: %v", err)
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", paymentOptionDetailsURL, strings.NewReader(string(paymentOptionDetailsReqData)))
	if err != nil {
		return nil, fmt.Errorf("create payment option details request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// getPaymentOptionDetails calls the payment option details API to get specific payment option details
func (c *Client) getPaymentOptionDetails(ctx context.Context, paymentToken string) (*APIResponse, error) {
	req, err := c.newPaymentOptionDetailsRequest(ctx, paymentToken)
	if err != nil {
		return nil, err
	}

	// Call payment option details API with debug info
	resp, debug, err := c.doRequestWithDebug(req)
	if err != nil {
		return nil, c.formatErrorWithDebug(fmt.Errorf("payment option details request: %w", err), debug)
	}
	defer resp.Body.Close()

	// Read payment option details response
	paymentOptionDetailsRespBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading payment option details response: %v", err)
	}
	log.Printf("Payment option details response body: %s", string(paymentOptionDetailsRespBody))

	// Parse payment option details response
	var paymentOptionDetailsRespData APIResponse
	if err := json.Unmarshal(paymentOptionDetailsRespBody, &paymentOptionDetailsRespData); err != nil {
		return nil, c.formatErrorWithDebug(fmt.Errorf("decode payment option details response: %w", err), debug)
	}

	return &paymentOptionDetailsRespData, nil
}

func (c *Client) newDoPaymentRequest(ctx context.Context, params *DoPaymentParams) (*http.Request, error) {
	doPaymentURL := c.endpoint("payment")
	log.Printf("Calling do payment API at %s", doPaymentURL)

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
	log.Printf("Do payment request payload (before JWT): %s", string(doPaymentData))

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", doPaymentURL, strings.NewReader(string(doPaymentData)))
	if err != nil {
		return nil, fmt.Errorf("create do payment request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// CreateQRPayment creates a new QR payment
func (c *Client) CreateQRPayment(ctx context.Context, params *CreateQRPaymentParams) (*QRPaymentResponse, error) {
	doPaymentParams := &DoPaymentParams{
		PaymentToken:       params.PaymentToken,
		PaymentChannelCode: params.PaymentChannelCode,
		PaymentData:        map[string]any{"qrType": "URL"},
		Locale:             "en",
		ResponseReturnUrl:  params.ResponseReturnUrl,
		ClientIP:           params.ClientIP,
	}

	// Create request
	req, err := c.newDoPaymentRequest(ctx, doPaymentParams)
	if err != nil {
		return nil, err
	}

	// Call do payment API with debug info
	resp, debug, err := c.doRequestWithDebug(req)
	if err != nil {
		return nil, c.formatErrorWithDebug(fmt.Errorf("do payment request: %w", err), debug)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read do payment response body: %w", err)
	}
	log.Printf("Do payment response body: %s", string(respBody))

	// Parse response
	var doPaymentRespData QRPaymentResponse
	if err := json.Unmarshal(respBody, &doPaymentRespData); err != nil {
		return nil, c.formatErrorWithDebug(fmt.Errorf("unmarshal do payment response: %w", err), debug)
	}

	return &doPaymentRespData, nil
}

// GetQRPaymentStatus gets the current status of a QR payment
func (c *Client) GetQRPaymentStatus(ctx context.Context, paymentToken string) (*PaymentInquiryResponse, error) {
	// Create payment inquiry request
	inquiryReq := &PaymentInquiryRequest{
		MerchantID:   c.MerchantID,
		PaymentToken: paymentToken,
		Locale:       "en",
	}

	// Get payment status
	return c.PaymentInquiry(ctx, inquiryReq)
}
