package api2c2p

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	// Amount is the payment amount (required)
	// Format: 12 digits with 5 decimal places (e.g., 000000002500.90000)
	AmountCents Cents `json:"amount"`

	// LoyaltyPoints is the loyalty points (optional)
	LoyaltyPoints *LoyaltyPoints `json:"loyaltyPoints,omitempty"`

	// CurrencyCodeISO4217 is the payment currency code (ISO 4217) (required)
	// Length: 3 characters
	CurrencyCodeISO4217 string `json:"currencyCode"`

	// PaymentChannel is the list of enabled payment channels (optional)
	// If empty, all payment channels will be used
	PaymentChannel []PaymentTokenPaymentChannel `json:"paymentChannel,omitempty"`

	// AgentChannel is the list of enabled agent channels (optional)
	// Required if paymentChannel includes "123"
	AgentChannel []string `json:"agentChannel,omitempty"`

	// Request3DS specifies whether to enable 3DS authentication (optional)
	// Y - Enable 3DS (default)
	// F - Force 3DS
	// N - Disable 3DS
	Request3DS PaymentTokenRequest3DSType `json:"request3DS,omitempty"`

	// ProtocolVersion is the 3DS protocol version (optional)
	// Default: "2.1.0"
	ProtocolVersion string `json:"protocolVersion,omitempty"`

	// ECI is the Electronic Commerce Indicator (conditional)
	// Required if protocolVersion, cavv, or dsTransactionId is provided
	ECI string `json:"eci,omitempty"`

	// CAVV is the Cardholder Authentication Verification Value (conditional)
	// Required if protocolVersion, eci, or dsTransactionId is provided
	CAVV string `json:"cavv,omitempty"`

	// DSTransactionID is the Directory Server Transaction ID (conditional)
	// Required if protocolVersion, eci, or cavv is provided
	DSTransactionID string `json:"dsTransactionId,omitempty"`

	// Tokenize enables card tokenization (optional)
	Tokenize bool `json:"tokenize,omitempty"`

	// CardTokens is a list of registered wallet tokens for payment (optional)
	CardTokens []string `json:"customerToken,omitempty"`

	// TokenizeOnly specifies whether to require tokenization with authorization (optional)
	// true - Tokenization without authorization
	// false - Tokenization with authorization (default)
	TokenizeOnly bool `json:"tokenizeOnly,omitempty"`

	// StoreCredentials specifies whether to store credentials (optional)
	// F - First time payment
	// S - Subsequent payment
	// N - Not using
	StoreCredentials string `json:"storeCredentials,omitempty"`

	// InterestType specifies the installment interest type (optional)
	// A - All available options (default)
	// C - Customer Pay Interest Option ONLY
	// M - Merchant Pay Interest Option ONLY
	InterestType PaymentTokenInterestType `json:"interestType,omitempty"`

	// InstallmentPeriodFilterMonths specifies which installment periods to offer (optional)
	InstallmentPeriodFilterMonths []int `json:"installmentPeriodFilter,omitempty"`

	// InstallmentBankFilter specifies which banks to offer installments for (optional)
	InstallmentBankFilter []string `json:"installmentBankFilter,omitempty"`

	// ProductCode is the installment product code (optional)
	ProductCode string `json:"productCode,omitempty"`

	// Recurring enables recurring payment (optional)
	Recurring bool `json:"recurring,omitempty"`

	// InvoicePrefix is used for recurring transactions (conditional)
	// Required if recurring is true
	InvoicePrefix string `json:"invoicePrefix,omitempty"`

	// RecurringAmount is the recurring charge amount (optional)
	// If not set, system will use transaction amount
	RecurringAmount float64 `json:"recurringAmount,omitempty"`

	// AllowAccumulate allows accumulation of failed recurring amounts (conditional)
	// Required if recurring is true
	AllowAccumulate bool `json:"allowAccumulate,omitempty"`

	// MaxAccumulateAmount is the maximum recurring accumulated amount (conditional)
	// Required if recurring is true
	MaxAccumulateAmount float64 `json:"maxAccumulateAmount,omitempty"`

	// RecurringIntervalDays is the interval in days between charges (conditional)
	// Required if recurring is true
	RecurringIntervalDays int `json:"recurringInterval,omitempty"`

	// RecurringCount is the number of recurring payment cycles (conditional)
	// Required if recurring is true
	// Set to 0 for indefinite recurring until terminated
	RecurringCount int `json:"recurringCount,omitempty"`

	// ChargeNextDateYYYYMMDD is the next recurring payment date (optional)
	// Format: ddMMyyyy
	ChargeNextDateYYYYMMDD string `json:"chargeNextDate,omitempty"`

	// ChargeOnDateYYYYMMDD is the specific day for recurring payments (conditional)
	// Format: ddMM
	// Required if recurring is true
	ChargeOnDateYYYYMMDD string `json:"chargeOnDate,omitempty"`

	// PaymentExpiryYYYYMMDDHHMMSS is the payment completion deadline (optional)
	// Format: yyyy-MM-dd HH:mm:ss
	// Default: 30 minutes
	PaymentExpiryYYYYMMDDHHMMSS string `json:"paymentExpiry,omitempty"`

	// PromotionCode is the promotion code for the payment (optional)
	PromotionCode string `json:"promotionCode,omitempty"`

	// PaymentRouteID specifies custom payment routing rules (optional)
	PaymentRouteID string `json:"paymentRouteID,omitempty"`

	// FxProviderCode is the forex provider code for multi-currency payments (optional)
	FxProviderCode string `json:"fxProviderCode,omitempty"`

	// FXRateID is the forex rate ID (optional)
	FXRateID string `json:"fxRateID,omitempty"`

	// OriginalAmount is the original currency amount (optional)
	OriginalAmount float64 `json:"originalAmount,omitempty"`

	// ImmediatePayment triggers payment immediately (optional)
	ImmediatePayment bool `json:"immediatePayment,omitempty"`

	// IframeMode enables iframe mode (optional)
	IframeMode bool `json:"iframeMode,omitempty"`

	// UserDefined1-5 are merchant-specific data fields (optional)
	UserDefined1 string `json:"userDefined1,omitempty"`
	UserDefined2 string `json:"userDefined2,omitempty"`
	UserDefined3 string `json:"userDefined3,omitempty"`
	UserDefined4 string `json:"userDefined4,omitempty"`
	UserDefined5 string `json:"userDefined5,omitempty"`

	// StatementDescriptor is a dynamic statement description (optional)
	// Length: 5-22 characters
	// Cannot contain special characters: < > \ ' " *
	StatementDescriptor string `json:"statementDescriptor,omitempty"`

	// ExternalSubMerchantID is an external sub-merchant ID (optional)
	ExternalSubMerchantID string `json:"externalSubMerchantID,omitempty"`

	// SubMerchants is a list of sub-merchants for split payments (optional)
	SubMerchants []PaymentTokenSubMerchant `json:"subMerchants,omitempty"`

	// UIParams contains user interface parameters for pre-filling payment forms
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
	log.Printf("Payment token request body: %s", string(jsonBody))

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
	resp, debug, err := c.doRequestWithDebug(httpReq)
	if err != nil {
		return nil, c.formatErrorWithDebug(fmt.Errorf("do request: %w", err), debug)
	}
	defer resp.Body.Close()
	log.Printf("Payment token response body: %s", debug.Response.Body)

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
		return &PaymentTokenResponse{
			RespCode: response.RespCode,
			RespDesc: response.RespDesc,
		}, nil
	}

	// If we got a JWT response, decode it
	var tokenResp PaymentTokenResponse
	if err := c.decodeJWTToken(jwtResponse.Payload, &tokenResp); err != nil {
		return nil, c.formatErrorWithDebug(fmt.Errorf("decode jwt token: %w", err), debug)
	}

	// Check response code
	if tokenResp.RespCode != "0000" {
		return &tokenResp, c.formatErrorWithDebug(fmt.Errorf("payment token failed: %s (%s)", tokenResp.RespCode, tokenResp.RespDesc), debug)
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

// PaymentTokenResponse represents the response from the Payment Token API
type PaymentTokenResponse struct {
	// RespCode is the response code
	// "0000" indicates success
	RespCode string `json:"respCode"`

	// RespDesc is the response description
	RespDesc string `json:"respDesc"`

	// PaymentToken is the generated payment token
	// Used for initiating the payment
	PaymentToken string `json:"paymentToken"`

	// WebPaymentURL is the URL to redirect customers for payment
	WebPaymentURL string `json:"webPaymentUrl"`
}
