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
type PaymentTokenRequest struct {
	AgentChannel                  []string                     `json:"agentChannel,omitempty"`
	AllowAccumulate               bool                         `json:"allowAccumulate,omitempty"`
	AmountCents                   Cents                        `json:"amount"`
	CAVV                          string                       `json:"cavv,omitempty"`
	CardTokens                    []string                     `json:"customerToken,omitempty"`
	ChargeNextDateYYYYMMDD        string                       `json:"chargeNextDate,omitempty"`
	ChargeOnDateYYYYMMDD          string                       `json:"chargeOnDate,omitempty"`
	CurrencyCodeISO4217           string                       `json:"currencyCode"`
	DSTransactionID               string                       `json:"dsTransactionId,omitempty"`
	Description                   string                       `json:"description"`
	ECI                           string                       `json:"eci,omitempty"`
	ExternalSubMerchantID         string                       `json:"externalSubMerchantID,omitempty"`
	FXRateID                      string                       `json:"fxRateID,omitempty"`
	FxProviderCode                string                       `json:"fxProviderCode,omitempty"`
	IdempotencyID                 string                       `json:"idempotencyID,omitempty"`
	IframeMode                    bool                         `json:"iframeMode,omitempty"`
	ImmediatePayment              bool                         `json:"immediatePayment,omitempty"`
	InstallmentBankFilter         []string                     `json:"installmentBankFilter,omitempty"`
	InstallmentPeriodFilterMonths []int                        `json:"installmentPeriodFilter,omitempty"`
	InterestType                  PaymentTokenInterestType     `json:"interestType,omitempty"`
	InvoiceNo                     string                       `json:"invoiceNo"`
	InvoicePrefix                 string                       `json:"invoicePrefix,omitempty"`
	LoyaltyPoints                 *LoyaltyPoints               `json:"loyaltyPoints,omitempty"`
	MaxAccumulateAmount           float64                      `json:"maxAccumulateAmount,omitempty"`
	MerchantID                    string                       `json:"merchantID"`
	OriginalAmount                float64                      `json:"originalAmount,omitempty"`
	PaymentChannel                []PaymentTokenPaymentChannel `json:"paymentChannel,omitempty"`
	PaymentExpiryYYYYMMDDHHMMSS   string                       `json:"paymentExpiry,omitempty"`
	PaymentRouteID                string                       `json:"paymentRouteID,omitempty"`
	ProductCode                   string                       `json:"productCode,omitempty"`
	PromotionCode                 string                       `json:"promotionCode,omitempty"`
	ProtocolVersion               string                       `json:"protocolVersion,omitempty"`
	Recurring                     bool                         `json:"recurring,omitempty"`
	RecurringAmount               float64                      `json:"recurringAmount,omitempty"`
	RecurringCount                int                          `json:"recurringCount,omitempty"`
	RecurringIntervalDays         int                          `json:"recurringInterval,omitempty"`
	Request3DS                    PaymentTokenRequest3DSType   `json:"request3DS,omitempty"`
	StatementDescriptor           string                       `json:"statementDescriptor,omitempty"`
	StoreCredentials              string                       `json:"storeCredentials,omitempty"`
	SubMerchants                  []PaymentTokenSubMerchant    `json:"subMerchants,omitempty"`
	Tokenize                      bool                         `json:"tokenize,omitempty"`
	TokenizeOnly                  bool                         `json:"tokenizeOnly,omitempty"`
	UIParams                      *paymentTokenUiParams        `json:"uiParams,omitempty"`
	UserDefined1                  string                       `json:"userDefined1,omitempty"`
	UserDefined2                  string                       `json:"userDefined2,omitempty"`
	UserDefined3                  string                       `json:"userDefined3,omitempty"`
	UserDefined4                  string                       `json:"userDefined4,omitempty"`
	UserDefined5                  string                       `json:"userDefined5,omitempty"`
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
			RespCode string `json:"respCode"`
			RespDesc string `json:"respDesc"`
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
	if tokenResp.RespCode != "0000" {
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
