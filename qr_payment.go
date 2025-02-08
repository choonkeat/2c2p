package api2c2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// QRPaymentRequest represents a request to create a QR code payment
type QRPaymentRequest struct {
	ClientID     string `json:"clientID"`
	ClientIP     string `json:"clientIP"`
	PaymentToken string `json:"paymentToken"`
	ReturnURL    string `json:"responseReturnUrl"`
}

// QRPaymentResponse represents a response from the QR code payment API
type QRPaymentResponse struct {
	ChannelCode       string `json:"channelCode"`
	Data              string `json:"data"`
	ExpiryDescription string `json:"expiryDescription"`
	ExpiryTimer       string `json:"expiryTimer"`
	RespCode          string `json:"respCode"`
	RespDesc          string `json:"respDesc"`
	Type              string `json:"type"`
}

// DoPaymentRequest represents a request to do a QR payment
type DoPaymentRequest struct {
	Locale  string `json:"locale"`
	Payment struct {
		Code struct {
			ChannelCode string `json:"channelCode"`
		} `json:"code"`
		Data map[string]interface{} `json:"data,omitempty"`
	} `json:"payment"`
	PaymentToken      string `json:"paymentToken"`
	ResponseReturnUrl string `json:"responseReturnUrl"`
}

// PaymentOptionRequest represents a request to get available payment options
// Documentation: https://developer.2c2p.com/v4.3.1/docs/api-payment-option-request-parameter
type PaymentOptionRequest struct {
	ClientID     string `json:"clientID,omitempty"`
	Locale       string `json:"locale,omitempty"`
	PaymentToken string `json:"paymentToken"`
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
	CategoryCode string `json:"categoryCode"`
	ClientID     string `json:"clientID,omitempty"`
	GroupCode    string `json:"groupCode"`
	Locale       string `json:"locale,omitempty"`
	PaymentToken string `json:"paymentToken"`
}

// APIResponse represents a response from the payment option details API
type APIResponse struct {
	Payload  string `json:"payload"`
	RespCode string `json:"respCode"`
	RespDesc string `json:"respDesc"`
}

// DoPaymentParams represents parameters for creating a new do payment request
type DoPaymentParams struct {
	ClientID           string
	ClientIP           string
	Locale             string
	PaymentChannelCode string
	PaymentData        map[string]any
	PaymentToken       string
	ResponseReturnUrl  string
}

// CreateQRPaymentParams represents parameters for creating a new QR payment
type CreateQRPaymentParams struct {
	ClientIP           string
	PaymentChannelCode string
	PaymentToken       string
	ResponseReturnUrl  string
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

// getPaymentOptions calls the payment option API to get available payment options
func (c *Client) getPaymentOptions(ctx context.Context, paymentToken string) (*PaymentOptionResponse, error) {
	req, err := c.newPaymentOptionsRequest(ctx, paymentToken)
	if err != nil {
		return nil, err
	}

	// Call payment option API
	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("payment option request: %w", err)
	}
	defer resp.Body.Close()

	// Read payment option response
	paymentOptionRespBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading payment option response: %v", err)
	}

	// Parse payment option response
	var paymentOptionRespData PaymentOptionResponse
	if err := json.Unmarshal(paymentOptionRespBody, &paymentOptionRespData); err != nil {
		return nil, fmt.Errorf("decode payment option response: %w", err)
	}

	return &paymentOptionRespData, nil
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

// getPaymentOptionDetails calls the payment option details API to get specific payment option details
func (c *Client) getPaymentOptionDetails(ctx context.Context, paymentToken string) (*APIResponse, error) {
	req, err := c.newPaymentOptionDetailsRequest(ctx, paymentToken)
	if err != nil {
		return nil, err
	}

	// Call payment option details API
	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("payment option details request: %w", err)
	}
	defer resp.Body.Close()

	// Read payment option details response
	paymentOptionDetailsRespBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading payment option details response: %v", err)
	}

	// Parse payment option details response
	var paymentOptionDetailsRespData APIResponse
	if err := json.Unmarshal(paymentOptionDetailsRespBody, &paymentOptionDetailsRespData); err != nil {
		return nil, fmt.Errorf("decode payment option details response: %w", err)
	}

	return &paymentOptionDetailsRespData, nil
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
func (c *Client) CreateQRPayment(ctx context.Context, params *CreateQRPaymentParams) (*QRPaymentResponse, error) {
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
	var doPaymentRespData QRPaymentResponse
	if err := json.Unmarshal(respBody, &doPaymentRespData); err != nil {
		return nil, fmt.Errorf("unmarshal do payment response: %w", err)
	}

	return &doPaymentRespData, nil
}

func (c *Client) GetQRPayment(req QRPaymentRequest) (*QRPaymentResponse, error) {
	// Get payment options
	paymentOptionData := map[string]string{
		"paymentToken": req.PaymentToken,
	}
	paymentOptionJSON, err := json.Marshal(paymentOptionData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payment option request: %w", err)
	}

	token, err := c.generateJWTToken(paymentOptionJSON)
	if err != nil {
		return nil, fmt.Errorf("error generating JWT token for payment option: %w", err)
	}

	paymentOptionReq, err := c.newRequest("POST", c.endpoint("paymentOption"), []byte(`{"payload":"`+token+`"}`))
	if err != nil {
		return nil, fmt.Errorf("error creating payment option request: %w", err)
	}

	resp, err := c.do(paymentOptionReq)
	if err != nil {
		return nil, fmt.Errorf("error getting payment options: %w", err)
	}
	defer resp.Body.Close()

	var paymentOptionResp struct {
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&paymentOptionResp); err != nil {
		return nil, fmt.Errorf("error decoding payment option response: %w", err)
	}

	// Get payment option details
	paymentOptionDetailsData := map[string]string{
		"paymentToken": req.PaymentToken,
		"categoryCode": "QR",
		"groupCode":    "SGQR",
	}
	paymentOptionDetailsJSON, err := json.Marshal(paymentOptionDetailsData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payment option details request: %w", err)
	}

	token, err = c.generateJWTToken(paymentOptionDetailsJSON)
	if err != nil {
		return nil, fmt.Errorf("error generating JWT token for payment option details: %w", err)
	}

	paymentOptionDetailsReq, err := c.newRequest("POST", c.endpoint("paymentOptionDetails"), []byte(`{"payload":"`+token+`"}`))
	if err != nil {
		return nil, fmt.Errorf("error creating payment option details request: %w", err)
	}

	resp, err = c.do(paymentOptionDetailsReq)
	if err != nil {
		return nil, fmt.Errorf("error getting payment option details: %w", err)
	}
	defer resp.Body.Close()

	var paymentOptionDetailsResp struct {
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&paymentOptionDetailsResp); err != nil {
		return nil, fmt.Errorf("error decoding payment option details response: %w", err)
	}

	// Create QR payment
	doPaymentData, err := json.Marshal(map[string]interface{}{
		"paymentToken":      req.PaymentToken,
		"responseReturnUrl": req.ReturnURL,
		"locale":            "en",
		"payment": map[string]interface{}{
			"code": map[string]string{
				"channelCode": "PNQR",
			},
			"data": map[string]string{
				"qrType": "URL",
			},
		},
		"clientID": req.ClientID,
		"clientIP": req.ClientIP,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshaling do payment request: %w", err)
	}

	token, err = c.generateJWTToken(doPaymentData)
	if err != nil {
		return nil, fmt.Errorf("error generating JWT token for do payment: %w", err)
	}

	doPaymentReq, err := c.newRequest("POST", c.endpoint("payment"), []byte(`{"payload":"`+token+`"}`))
	if err != nil {
		return nil, fmt.Errorf("error creating do payment request: %w", err)
	}

	resp, err = c.do(doPaymentReq)
	if err != nil {
		return nil, fmt.Errorf("error making payment: %w", err)
	}
	defer resp.Body.Close()

	var paymentResp struct {
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, fmt.Errorf("error decoding payment response: %w", err)
	}

	var result QRPaymentResponse
	if err := c.decodeJWTToken(paymentResp.Payload, &result); err != nil {
		return nil, fmt.Errorf("error decoding JWT token: %w", err)
	}

	return &result, nil
}
