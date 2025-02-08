package api2c2p

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// PaymentTokenRequest3DSType represents the 3DS request type
type PaymentTokenRequest3DSType string

const (
	// Request3DSYes - Perform 3DS check
	Request3DSYes PaymentTokenRequest3DSType = "Y"
	// Request3DSNo - Do not perform 3DS check
	Request3DSNo PaymentTokenRequest3DSType = "N"
	// Request3DSFrictionless - Perform frictionless flow
	Request3DSFrictionless PaymentTokenRequest3DSType = "F"
)

// PaymentTokenPaymentChannel represents a payment channel
type PaymentTokenPaymentChannel string

const (
	// PaymentChannelCC represents credit card payment
	PaymentChannelCC PaymentTokenPaymentChannel = "CC"
	// PaymentChannelIPP represents installment payment plan
	PaymentChannelIPP PaymentTokenPaymentChannel = "IPP"
	// PaymentChannelAPM represents alternative payment methods
	PaymentChannelAPM PaymentTokenPaymentChannel = "APM"
)

// PaymentTokenInterestType represents the installment interest type
type PaymentTokenInterestType string

const (
	// InterestTypeAll shows all available interest options
	InterestTypeAll PaymentTokenInterestType = "A"
	// InterestTypeCustomer shows only customer pay interest options
	InterestTypeCustomer PaymentTokenInterestType = "C"
	// InterestTypeMerchant shows only merchant pay interest options
	InterestTypeMerchant PaymentTokenInterestType = "M"
)

type LoyaltyPoints struct {
	RedeemAmount float64 `json:"redeemAmount"`
}

type Cents int64

func (c Cents) XMLString() string {
	return fmt.Sprintf("%012d", c)
}

// Format: 12 digits with 5 decimal places (e.g., 000000002500.90000)
func (c Cents) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%012d.%02d000\"", c/100, c%100)), nil
}

// Decodes "000000000012.34000" into 1234
func (c *Cents) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Split by decimal point
	split := bytes.SplitN([]byte(s), []byte("."), 2)
	if len(split) != 2 {
		return fmt.Errorf("invalid format")
	}
	// Parse first part
	val, err := strconv.ParseInt(string(split[0]), 10, 64)
	if err != nil {
		return err
	}
	// Parse second part
	val2, err := strconv.ParseInt(string(split[1]), 10, 64)
	if err != nil {
		return err
	}
	val = val*100 + (val2 / 1000)
	*c = Cents(val)
	return nil
}

// PaymentTokenRequest represents a request to the Payment Token API
// Documentation: https://developer.2c2p.com/v4.3.1/docs/api-payment-token-request-parameter
type PaymentTokenRequest struct {
	// MerchantID is the 2C2P merchant ID (required)
	// Max length: 8 characters
	MerchantID string `json:"merchantID"`

	// IdempotencyID is a unique value for retrying same requests (optional)
	// Max length: 100 characters
	IdempotencyID string `json:"idempotencyID,omitempty"`

	// InvoiceNo is the merchant's invoice number (required)
	// Max length: 50 characters
	InvoiceNo string `json:"invoiceNo"`

	// Description is the payment description (required)
	// Max length: 250 characters
	Description string `json:"description"`

	// AmountCents is the payment amount (required)
	// Format: 12 digits with 5 decimal places (e.g., 000000002500.90000)
	AmountCents Cents `json:"amount"`

	// LoyaltyPoints is the loyalty points (optional)
	LoyaltyPoints *LoyaltyPoints `json:"loyaltyPoints,omitempty"`

	// CurrencyCodeISO4217 is the payment currency code (ISO 4217) (required)
	// Length: 3 characters
	CurrencyCodeISO4217 string `json:"currencyCode"`

	// PaymentChannel is a comma-separated list of payment channels (optional)
	// Default: "CC"
	PaymentChannel []PaymentTokenPaymentChannel `json:"paymentChannel,omitempty"`

	// PaymentExpiryYYYYMMDDHHMMSS is the payment expiry date/time (optional)
	// Format: YYYY-MM-DD HH:mm:ss
	PaymentExpiryYYYYMMDDHHMMSS string `json:"paymentExpiry,omitempty"`

	// UserDefined1 is a custom field (optional)
	// Max length: 255 characters
	UserDefined1 string `json:"userDefined1,omitempty"`

	// UserDefined2 is a custom field (optional)
	// Max length: 255 characters
	UserDefined2 string `json:"userDefined2,omitempty"`

	// UserDefined3 is a custom field (optional)
	// Max length: 255 characters
	UserDefined3 string `json:"userDefined3,omitempty"`

	// UserDefined4 is a custom field (optional)
	// Max length: 255 characters
	UserDefined4 string `json:"userDefined4,omitempty"`

	// UserDefined5 is a custom field (optional)
	// Max length: 255 characters
	UserDefined5 string `json:"userDefined5,omitempty"`

	// StatementDescriptor is the dynamic statement description (optional)
	// Max length: 25 characters
	StatementDescriptor string `json:"statementDescriptor,omitempty"`

	// CardTokens is a comma-separated list of card tokens (optional)
	CardTokens []string `json:"cardTokens,omitempty"`

	// Request3DS specifies the 3DS request type (optional)
	// Values: "Y" (enforce 3DS), "N" (bypass 3DS), "F" (follow rules)
	// Default: "Y"
	Request3DS PaymentTokenRequest3DSType `json:"request3DS,omitempty"`

	// ProtocolVersion is the 3DS protocol version (optional)
	ProtocolVersion string `json:"protocolVersion,omitempty"`

	// ECI is the Electronic Commerce Indicator (optional)
	ECI string `json:"eci,omitempty"`

	// CAVV is the Cardholder Authentication Verification Value (optional)
	CAVV string `json:"cavv,omitempty"`

	// DSTransactionID is the Directory Server Transaction ID (optional)
	DSTransactionID string `json:"dsTransactionID,omitempty"`

	// StoreCredentials specifies whether to store credentials (optional)
	// Values: "F" (First time), "S" (Subsequent), "N" (No)
	StoreCredentials string `json:"storeCredentials,omitempty"`

	// Tokenize enables tokenization (optional)
	Tokenize bool `json:"tokenize,omitempty"`

	// TokenizeOnly only tokenizes without processing payment (optional)
	TokenizeOnly bool `json:"tokenizeOnly,omitempty"`

	// IframeMode enables iframe mode (optional)
	IframeMode bool `json:"iframeMode,omitempty"`

	// PaymentRouteID specifies the payment route ID (optional)
	PaymentRouteID string `json:"paymentRouteID,omitempty"`

	// ProductCode is the product code (optional)
	ProductCode string `json:"productCode,omitempty"`

	// PromotionCode is the promotion code (optional)
	PromotionCode string `json:"promotionCode,omitempty"`

	// InstallmentBankFilter is a comma-separated list of installment banks (optional)
	InstallmentBankFilter []string `json:"installmentBankFilter,omitempty"`

	// InstallmentPeriodFilterMonths is a comma-separated list of installment periods in months (optional)
	InstallmentPeriodFilterMonths []int `json:"installmentPeriodFilter,omitempty"`

	// InterestType specifies the interest type (optional)
	// Values: "A" (Advance), "C" (Customer), "M" (Merchant)
	InterestType PaymentTokenInterestType `json:"interestType,omitempty"`

	// AgentChannel is a comma-separated list of agent channels (optional)
	AgentChannel []string `json:"agentChannel,omitempty"`

	// FXRateID is the forex rate ID (optional)
	FXRateID string `json:"fxRateID,omitempty"`

	// FxProviderCode is the forex provider code (optional)
	FxProviderCode string `json:"fxProviderCode,omitempty"`

	// OriginalAmount is the original currency amount (optional)
	OriginalAmount float64 `json:"originalAmount,omitempty"`

	// SubMerchantID is the sub-merchant ID (optional)
	SubMerchantID string `json:"subMerchantID,omitempty"`

	// ExternalSubMerchantID is the external sub-merchant ID (optional)
	ExternalSubMerchantID string `json:"externalSubMerchantID,omitempty"`

	// SubMerchantInvoiceNo is the sub-merchant invoice number (optional)
	SubMerchantInvoiceNo string `json:"subMerchantInvoiceNo,omitempty"`

	// SubMerchantDescription is the sub-merchant description (optional)
	SubMerchantDescription string `json:"subMerchantDescription,omitempty"`

	// SubMerchantAmount is the sub-merchant amount (optional)
	SubMerchantAmount float64 `json:"subMerchantAmount,omitempty"`

	// Recurring enables recurring payment (optional)
	Recurring bool `json:"recurring,omitempty"`

	// RecurringAmount is the amount for recurring payments (optional)
	RecurringAmount float64 `json:"recurringAmount,omitempty"`

	// RecurringCount is the total number of recurring payments (optional)
	RecurringCount int `json:"recurringCount,omitempty"`

	// RecurringIntervalDays is the interval in days between recurring payments (optional)
	RecurringIntervalDays int `json:"recurringInterval,omitempty"`

	// ChargeNextDateYYYYMMDD is the next charge date (optional)
	// Format: YYYYMMDD
	ChargeNextDateYYYYMMDD string `json:"chargeNextDate,omitempty"`

	// ChargeOnDateYYYYMMDD is the specific charge date (optional)
	// Format: YYYYMMDD
	ChargeOnDateYYYYMMDD string `json:"chargeOnDate,omitempty"`

	// AllowAccumulate allows accumulation of recurring payments (optional)
	AllowAccumulate bool `json:"allowAccumulate,omitempty"`

	// MaxAccumulateAmount is the maximum amount for accumulated payments (optional)
	MaxAccumulateAmount float64 `json:"maxAccumulateAmount,omitempty"`

	// InvoicePrefix is the invoice prefix for recurring payments (optional)
	InvoicePrefix string `json:"invoicePrefix,omitempty"`

	// ImmediatePayment triggers payment immediately (optional)
	ImmediatePayment bool `json:"immediatePayment,omitempty"`

	// SubMerchants is a list of sub-merchants for split payments (optional)
	SubMerchants []PaymentTokenSubMerchant `json:"subMerchants,omitempty"`

	// UIParams is the UI parameters for payment token requests (optional)
	UIParams *paymentTokenUiParams `json:"uiParams,omitempty"`
}

func (c *Client) newPaymentTokenRequest(ctx context.Context, req *PaymentTokenRequest) (*http.Request, error) {
	url := c.endpoint("paymentToken")
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

// PaymentToken creates a payment token for processing a payment
func (c *Client) PaymentToken(ctx context.Context, req *PaymentTokenRequest) (*PaymentTokenResponse, error) {
	if req.MerchantID == "" {
		req.MerchantID = c.MerchantID
	}

	// Create and make request
	httpReq, err := c.newPaymentTokenRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// Make request with debug info
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
			RespCode PaymentResponseCode `json:"respCode"`
			RespDesc string              `json:"respDesc"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &PaymentTokenResponse{
			RespCode: response.RespCode,
			RespDesc: response.RespDesc,
		}, nil
	}

	// If we got a JWT response, decode it
	var tokenResp PaymentTokenResponse
	if err := c.decodeJWTToken(jwtResponse.Payload, &tokenResp); err != nil {
		return nil, fmt.Errorf("decode jwt token: %w", err)
	}

	// Check response code
	if tokenResp.RespCode != Code0000Successful {
		return &tokenResp, fmt.Errorf("payment token failed: %s (%s)", tokenResp.RespCode, tokenResp.RespDesc)
	}

	return &tokenResp, nil
}

// PaymentTokenSubMerchant represents a sub-merchant for split payments
type PaymentTokenSubMerchant struct {
	// MerchantID is the sub-merchant's 2C2P merchant ID (required)
	// Max length: 8 characters
	MerchantID string `json:"merchantID"`

	// InvoiceNo is the sub-merchant's unique transaction ID (required)
	// Max length: 30 characters
	InvoiceNo string `json:"invoiceNo"`

	// Amount is the payment amount for this sub-merchant (required)
	Amount float64 `json:"amount"`

	// Description is the payment description for this sub-merchant (required)
	// Max length: 250 characters
	Description string `json:"description"`
}

// paymentTokenUiParams represents UI parameters for payment token requests
type paymentTokenUiParams struct {
	// UserInfo contains customer information for pre-filling payment forms
	UserInfo *paymentTokenUserInfo `json:"userInfo,omitempty"`
}

// paymentTokenUserInfo represents user information for payment token requests
type paymentTokenUserInfo struct {
	// Name is the customer's full name
	Name string `json:"name"`

	// Email is the customer's email address
	Email string `json:"email"`

	// Address is the customer's address
	Address string `json:"address"`

	// MobileNo is the customer's mobile number
	MobileNo string `json:"mobileNo"`

	// CountryCodeISO3166 is the customer's country code (ISO 3166)
	CountryCodeISO3166 string `json:"countryCode"`

	// MobileNoPrefix is the customer's mobile number prefix
	MobileNoPrefix string `json:"mobileNoPrefix"`

	// CurrencyCodeISO4217 is the customer's preferred currency code (ISO 4217)
	CurrencyCodeISO4217 string `json:"currencyCode"`
}

// PaymentTokenResponse represents a response from the Payment Token API
// Documentation: https://developer.2c2p.com/v4.3.1/docs/api-payment-token-response-parameter
type PaymentTokenResponse struct {
	// RespCode is the response code
	// "0000" indicates success
	RespCode PaymentResponseCode `json:"respCode"`

	// RespDesc is the response description
	RespDesc string `json:"respDesc"`

	// PaymentToken is the token to be used for payment
	PaymentToken string `json:"paymentToken"`

	// WebPaymentURL is the URL to redirect customers for payment
	WebPaymentURL string `json:"webPaymentUrl"`
}

// IsSuccess returns true if the response code indicates success
func (r *PaymentTokenResponse) IsSuccess() bool {
	return r.RespCode == Code0000Successful
}
