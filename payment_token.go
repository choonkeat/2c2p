package api2c2p

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Request3DSType represents the 3DS request type
type Request3DSType string

const (
	// Request3DSYes - Perform 3DS check
	Request3DSYes Request3DSType = "Y"
	// Request3DSNo - Do not perform 3DS check
	Request3DSNo Request3DSType = "N"
	// Request3DSFrictionless - Perform frictionless flow
	Request3DSFrictionless Request3DSType = "F"
)

// PaymentChannel represents available payment channels
type PaymentChannel string

const (
	// PaymentChannelCC - Credit Card
	PaymentChannelCC PaymentChannel = "CC"
	// PaymentChannelIPP - Installment Payment Plan
	PaymentChannelIPP PaymentChannel = "IPP"
	// PaymentChannelAPM - Alternative Payment Method
	PaymentChannelAPM PaymentChannel = "APM"
)

// InterestType represents the interest type for installment payments
type InterestType string

const (
	// InterestMerchant - Merchant pays the interest
	InterestMerchant InterestType = "M"
	// InterestCustomer - Customer pays the interest
	InterestCustomer InterestType = "C"
)

// PaymentTokenRequest represents a request to the Payment Token API
type PaymentTokenRequest struct {
	// MerchantID is the 2C2P merchant ID (required)
	// Max length: 8 characters
	MerchantID string `json:"merchantID"`

	// CurrencyCodeISO4217 is the payment currency (ISO 4217) (required)
	// Example: "THB", "USD"
	CurrencyCodeISO4217 string `json:"currencyCode"`

	// Amount is the payment amount (required)
	// Up to 2 decimal places
	Amount float64 `json:"amount"`

	// InvoiceNo is a unique transaction ID (required)
	// Max length: 30 characters
	InvoiceNo string `json:"invoiceNo"`

	// Description is the payment description (required)
	// Max length: 250 characters
	Description string `json:"description"`

	// PaymentChannel specifies allowed payment methods (optional)
	// Example: ["CC", "IPP"]
	PaymentChannel []PaymentChannel `json:"paymentChannel,omitempty"`

	// Request3DS enables 3D Secure for credit card payments (optional)
	// Default: "Y" (Yes)
	Request3DS Request3DSType `json:"request3DS,omitempty"`

	// Tokenize enables card tokenization (optional)
	Tokenize bool `json:"tokenize,omitempty"`

	// CardTokens specifies card tokens to use for payment (optional)
	// Max length per token: 128 characters
	CardTokens []string `json:"cardTokens,omitempty"`

	// CardTokenOnly restricts payment to tokenized cards only (optional)
	CardTokenOnly bool `json:"cardTokenOnly,omitempty"`

	// TokenizeOnly enables tokenization without charging (optional)
	TokenizeOnly bool `json:"tokenizeOnly,omitempty"`

	// InterestType specifies who pays the installment interest (optional)
	// Only applicable when PaymentChannel includes IPP
	InterestType InterestType `json:"interestType,omitempty"`

	// InstallmentPeriodFilterMonths specifies available installment periods (optional)
	// Only applicable when PaymentChannel includes IPP
	// Example: [3, 6, 9] for 3, 6, and 9 months
	InstallmentPeriodFilterMonths []int `json:"installmentPeriodFilter,omitempty"`

	// ProductCode is the product code for specific payment providers (optional)
	// Max length: 50 characters
	ProductCode string `json:"productCode,omitempty"`

	// Recurring enables recurring payment (optional)
	Recurring bool `json:"recurring,omitempty"`

	// InvoicePrefix is used for recurring payment invoice generation (optional)
	// Max length: 20 characters
	InvoicePrefix string `json:"invoicePrefix,omitempty"`

	// RecurringAmount is the amount for recurring payments (optional)
	// Required if Recurring is true
	RecurringAmount float64 `json:"recurringAmount,omitempty"`

	// AllowAccumulate allows accumulation of recurring payments (optional)
	AllowAccumulate bool `json:"allowAccumulate,omitempty"`

	// MaxAccumulateAmount is the maximum amount for accumulated payments (optional)
	// Required if AllowAccumulate is true
	MaxAccumulateAmount float64 `json:"maxAccumulateAmount,omitempty"`

	// RecurringIntervalDays is the interval in days between recurring payments (optional)
	// Required if Recurring is true
	RecurringIntervalDays int `json:"recurringInterval,omitempty"`

	// RecurringCount is the total number of recurring payments (optional)
	// Required if Recurring is true
	RecurringCount int `json:"recurringCount,omitempty"`

	// ChargeNextDateYYYYMMDD is the next charge date (optional)
	// Format: YYYY-MM-DD
	ChargeNextDateYYYYMMDD string `json:"chargeNextDate,omitempty"`

	// ChargeOnDateYYYYMMDD is the specific charge date (optional)
	// Format: YYYY-MM-DD
	ChargeOnDateYYYYMMDD string `json:"chargeOnDate,omitempty"`

	// PaymentExpiryYYYYMMDDHHMMSS is the payment token expiry date and time (optional)
	// Format: YYYY-MM-DD HH:mm:ss
	PaymentExpiryYYYYMMDDHHMMSS string `json:"paymentExpiry,omitempty"`

	// PromotionCode is the promotion code to apply (optional)
	// Max length: 50 characters
	PromotionCode string `json:"promotionCode,omitempty"`

	// PaymentRouteID specifies a specific payment route (optional)
	PaymentRouteID string `json:"paymentRouteID,omitempty"`

	// FxProviderCode specifies the FX provider (optional)
	FxProviderCode string `json:"fxProviderCode,omitempty"`

	// ImmediatePayment requires immediate payment processing (optional)
	ImmediatePayment bool `json:"immediatePayment,omitempty"`

	// UserDefined1-5 are custom fields for your use (optional)
	// Max length: 255 characters each
	UserDefined1 string `json:"userDefined1,omitempty"`
	UserDefined2 string `json:"userDefined2,omitempty"`
	UserDefined3 string `json:"userDefined3,omitempty"`
	UserDefined4 string `json:"userDefined4,omitempty"`
	UserDefined5 string `json:"userDefined5,omitempty"`

	// StatementDescriptor is the descriptor shown on customer's card statement (optional)
	// Max length: 50 characters
	StatementDescriptor string `json:"statementDescriptor,omitempty"`

	// SubMerchants contains sub-merchant information for split payments (optional)
	SubMerchants []SubMerchant `json:"subMerchants,omitempty"`

	// Locale specifies the payment page language (optional)
	// Default: "en"
	Locale string `json:"locale,omitempty"`

	// FrontendReturnURL is the URL to return to after frontend payment completion (optional)
	// Must be HTTPS URL
	FrontendReturnURL string `json:"frontendReturnUrl,omitempty"`

	// BackendReturnURL is the URL for payment notification (optional)
	// Must be HTTPS URL
	BackendReturnURL string `json:"backendReturnUrl,omitempty"`

	// NonceStr is a random string for request uniqueness (optional)
	// Max length: 32 characters
	NonceStr string `json:"nonceStr,omitempty"`

	// UIParams contains user interface customization parameters (optional)
	UIParams *UIParams `json:"uiParams,omitempty"`
}

// SubMerchant represents a sub-merchant for split payments
type SubMerchant struct {
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

// UIParams represents the UI parameters for payment customization
type UIParams struct {
	// UserInfo contains customer information for pre-filling payment forms
	UserInfo *UserInfo `json:"userInfo,omitempty"`
}

// UserInfo represents customer information for payment forms
type UserInfo struct {
	// Name is the customer's full name
	// Max length: 255 characters
	Name string `json:"name,omitempty"`

	// Email is the customer's email address
	// Max length: 255 characters
	Email string `json:"email,omitempty"`

	// MobileNo is the customer's mobile number without country code
	// Max length: 50 characters
	MobileNo string `json:"mobileNo,omitempty"`

	// CountryCodeISO3166 is the customer's country code (ISO 3166-1 alpha-2)
	// Example: "SG", "TH"
	CountryCodeISO3166 string `json:"countryCode,omitempty"`

	// MobileNoPrefix is the customer's mobile number country code
	// Example: "65" for Singapore
	MobileNoPrefix string `json:"mobileNoPrefix,omitempty"`

	// CurrencyCodeISO4217 is the customer's preferred currency (ISO 4217)
	CurrencyCodeISO4217 string `json:"currencyCode,omitempty"`
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

// PaymentToken creates a payment token for processing a payment
func (c *Client) PaymentToken(ctx context.Context, req *PaymentTokenRequest) (*PaymentTokenResponse, error) {
	if req.MerchantID == "" {
		req.MerchantID = c.MerchantID
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	token, err := c.GenerateJWTToken(payload)
	if err != nil {
		return nil, fmt.Errorf("generate jwt token: %w", err)
	}

	// Make HTTP request
	url := fmt.Sprintf("%s/paymentToken", c.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	var tokenResp PaymentTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &tokenResp, nil
}
