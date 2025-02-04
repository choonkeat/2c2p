package api2c2p

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

// PaymentChannel represents a payment channel
type PaymentChannel string

const (
	// PaymentChannelCC represents credit card payment
	PaymentChannelCC PaymentChannel = "CC"
	// PaymentChannelIPP represents installment payment plan
	PaymentChannelIPP PaymentChannel = "IPP"
	// PaymentChannelAPM represents alternative payment methods
	PaymentChannelAPM PaymentChannel = "APM"
)

// InterestType represents the installment interest type
type InterestType string

const (
	// InterestTypeAll shows all available interest options
	InterestTypeAll InterestType = "A"
	// InterestTypeCustomer shows only customer pay interest options
	InterestTypeCustomer InterestType = "C"
	// InterestTypeMerchant shows only merchant pay interest options
	InterestTypeMerchant InterestType = "M"
)

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
	Amount float64 `json:"amount"`

	// CurrencyCodeISO4217 is the payment currency code (ISO 4217) (required)
	// Length: 3 characters
	CurrencyCodeISO4217 string `json:"currencyCode"`

	// PaymentChannel is the list of enabled payment channels (optional)
	// If empty, all payment channels will be used
	PaymentChannel []PaymentChannel `json:"paymentChannel,omitempty"`

	// AgentChannel is the list of enabled agent channels (optional)
	// Required if paymentChannel includes "123"
	AgentChannel []string `json:"agentChannel,omitempty"`

	// Request3DS specifies whether to enable 3DS authentication (optional)
	// Y - Enable 3DS (default)
	// F - Force 3DS
	// N - Disable 3DS
	Request3DS Request3DSType `json:"request3DS,omitempty"`

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
	InterestType InterestType `json:"interestType,omitempty"`

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
	FXRateID string `json:"fxRateId,omitempty"`

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
	SubMerchants []SubMerchant `json:"subMerchants,omitempty"`

	// UIParams contains UI customization parameters (optional)
	UIParams *UIParams `json:"uiParams,omitempty"`
}

// ToMap converts PaymentTokenRequest to a map[string]string for JWT payload
func (r *PaymentTokenRequest) ToMap() map[string]string {
	m := make(map[string]string)

	// Required fields
	m["merchantID"] = r.MerchantID
	m["invoiceNo"] = r.InvoiceNo
	m["description"] = r.Description
	m["amount"] = fmt.Sprintf("%015.5f", r.Amount)
	m["currencyCode"] = r.CurrencyCodeISO4217

	// Optional fields
	if r.IdempotencyID != "" {
		m["idempotencyID"] = r.IdempotencyID
	}

	if len(r.PaymentChannel) > 0 {
		channels := make([]string, len(r.PaymentChannel))
		for i, ch := range r.PaymentChannel {
			channels[i] = string(ch)
		}
		m["paymentChannel"] = strings.Join(channels, ",")
	}

	if len(r.AgentChannel) > 0 {
		m["agentChannel"] = strings.Join(r.AgentChannel, ",")
	}

	if r.Request3DS != "" {
		m["request3DS"] = string(r.Request3DS)
	}

	if r.ProtocolVersion != "" {
		m["protocolVersion"] = r.ProtocolVersion
	}

	if r.ECI != "" {
		m["eci"] = r.ECI
	}

	if r.CAVV != "" {
		m["cavv"] = r.CAVV
	}

	if r.DSTransactionID != "" {
		m["dsTransactionId"] = r.DSTransactionID
	}

	if r.Tokenize {
		m["tokenize"] = "Y"
	}

	if len(r.CardTokens) > 0 {
		m["customerToken"] = strings.Join(r.CardTokens, ",")
	}

	if r.TokenizeOnly {
		m["tokenizeOnly"] = "Y"
	}

	if r.StoreCredentials != "" {
		m["storeCredentials"] = r.StoreCredentials
	}

	if r.InterestType != "" {
		m["interestType"] = string(r.InterestType)
	}

	if len(r.InstallmentPeriodFilterMonths) > 0 {
		periods := make([]string, len(r.InstallmentPeriodFilterMonths))
		for i, p := range r.InstallmentPeriodFilterMonths {
			periods[i] = strconv.Itoa(p)
		}
		m["installmentPeriodFilter"] = strings.Join(periods, ",")
	}

	if len(r.InstallmentBankFilter) > 0 {
		m["installmentBankFilter"] = strings.Join(r.InstallmentBankFilter, ",")
	}

	if r.ProductCode != "" {
		m["productCode"] = r.ProductCode
	}

	if r.Recurring {
		m["recurring"] = "Y"
	}

	if r.InvoicePrefix != "" {
		m["invoicePrefix"] = r.InvoicePrefix
	}

	if r.RecurringAmount > 0 {
		m["recurringAmount"] = fmt.Sprintf("%015.5f", r.RecurringAmount)
	}

	if r.AllowAccumulate {
		m["allowAccumulate"] = "Y"
	}

	if r.MaxAccumulateAmount > 0 {
		m["maxAccumulateAmount"] = fmt.Sprintf("%015.5f", r.MaxAccumulateAmount)
	}

	if r.RecurringIntervalDays > 0 {
		m["recurringInterval"] = strconv.Itoa(r.RecurringIntervalDays)
	}

	if r.RecurringCount > 0 {
		m["recurringCount"] = strconv.Itoa(r.RecurringCount)
	}

	if r.ChargeNextDateYYYYMMDD != "" {
		m["chargeNextDate"] = r.ChargeNextDateYYYYMMDD
	}

	if r.ChargeOnDateYYYYMMDD != "" {
		m["chargeOnDate"] = r.ChargeOnDateYYYYMMDD
	}

	if r.PaymentExpiryYYYYMMDDHHMMSS != "" {
		m["paymentExpiry"] = r.PaymentExpiryYYYYMMDDHHMMSS
	}

	if r.PromotionCode != "" {
		m["promotionCode"] = r.PromotionCode
	}

	if r.PaymentRouteID != "" {
		m["paymentRouteID"] = r.PaymentRouteID
	}

	if r.FxProviderCode != "" {
		m["fxProviderCode"] = r.FxProviderCode
	}

	if r.FXRateID != "" {
		m["fxRateId"] = r.FXRateID
	}

	if r.OriginalAmount > 0 {
		m["originalAmount"] = fmt.Sprintf("%015.5f", r.OriginalAmount)
	}

	if r.ImmediatePayment {
		m["immediatePayment"] = "Y"
	}

	if r.IframeMode {
		m["iframeMode"] = "Y"
	}

	if r.UserDefined1 != "" {
		m["userDefined1"] = r.UserDefined1
	}

	if r.UserDefined2 != "" {
		m["userDefined2"] = r.UserDefined2
	}

	if r.UserDefined3 != "" {
		m["userDefined3"] = r.UserDefined3
	}

	if r.UserDefined4 != "" {
		m["userDefined4"] = r.UserDefined4
	}

	if r.UserDefined5 != "" {
		m["userDefined5"] = r.UserDefined5
	}

	if r.StatementDescriptor != "" {
		m["statementDescriptor"] = r.StatementDescriptor
	}

	if r.ExternalSubMerchantID != "" {
		m["externalSubMerchantID"] = r.ExternalSubMerchantID
	}

	if len(r.SubMerchants) > 0 {
		for i, sm := range r.SubMerchants {
			prefix := fmt.Sprintf("subMerchant[%d].", i)
			m[prefix+"merchantID"] = sm.MerchantID
			m[prefix+"amount"] = fmt.Sprintf("%015.5f", sm.Amount)
			m[prefix+"invoiceNo"] = sm.InvoiceNo
			m[prefix+"description"] = sm.Description
		}
	}

	if r.UIParams != nil {
		if r.UIParams.UserInfo != nil {
			ui := r.UIParams.UserInfo
			if ui.Name != "" {
				m["userInfo.name"] = ui.Name
			}
			if ui.Email != "" {
				m["userInfo.email"] = ui.Email
			}
			if ui.Address != "" {
				m["userInfo.address"] = ui.Address
			}
			if ui.CountryCodeISO3166 != "" {
				m["userInfo.countryCode"] = ui.CountryCodeISO3166
			}
		}
	}

	return m
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

	// Address is the customer's address
	// Max length: 255 characters
	Address string `json:"address,omitempty"`

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

	// Make HTTP request
	url := c.endpoint("paymentToken")

	// Format amount as D(12,5) - 12 digits before decimal, 5 after
	amountStr := fmt.Sprintf("%012.5f", float64(req.Amount)/100.0)

	// Prepare payload
	payload := map[string]interface{}{
		"merchantID":     req.MerchantID,
		"invoiceNo":      req.InvoiceNo,
		"description":    req.Description,
		"amount":         amountStr,
		"currencyCode":   req.CurrencyCodeISO4217,
		"locale":         "en",
		"request3DS":     "Y",
		"paymentChannel": req.PaymentChannel,
		"subMerchants":   req.SubMerchants,
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Generate JWT token
	token, err := c.GenerateJWTToken(jsonData)
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

	// Make request with debug info
	resp, debug, err := c.doRequestWithDebug(httpReq)
	if err != nil {
		return nil, c.formatErrorWithDebug(fmt.Errorf("do request: %w", err), debug)
	}
	defer resp.Body.Close()

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
	if err := c.DecodeJWTToken(jwtResponse.Payload, &tokenResp); err != nil {
		return nil, c.formatErrorWithDebug(fmt.Errorf("decode jwt token: %w", err), debug)
	}

	// Check response code
	if tokenResp.RespCode != "0000" {
		return &tokenResp, c.formatErrorWithDebug(fmt.Errorf("payment token failed: %s (%s)", tokenResp.RespCode, tokenResp.RespDesc), debug)
	}

	return &tokenResp, nil
}
