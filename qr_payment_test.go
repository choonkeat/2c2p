package api2c2p

import (
	"testing"

	"github.com/choonkeat/2c2p/testutil"
)

func TestNewPaymentOptionsRequest(t *testing.T) {
	client := NewClient("your_secret_key", "JT01", "https://example.com")
	paymentToken := "test_payment_token"

	// Create request
	httpReq, err := client.newPaymentOptionsRequest(ctx, paymentToken)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Verify request using testutil
	testutil.AssertRequest(t, httpReq, struct {
		Method      string
		URL         string
		ContentType string
		Headers     map[string]string
		Body        any
	}{
		Method:      "POST",
		URL:         "https://example.com/payment/4.3/paymentOption",
		ContentType: "application/json",
		Body: map[string]any{
			"paymentToken": "test_payment_token",
		},
	})
}

func TestNewPaymentOptionDetailsRequest(t *testing.T) {
	client := NewClient("your_secret_key", "JT01", "https://example.com")
	paymentToken := "test_payment_token"

	// Create request
	httpReq, err := client.newPaymentOptionDetailsRequest(ctx, paymentToken)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Verify request using testutil
	testutil.AssertRequest(t, httpReq, struct {
		Method      string
		URL         string
		ContentType string
		Headers     map[string]string
		Body        any
	}{
		Method:      "POST",
		URL:         "https://example.com/payment/4.3/paymentOptionDetails",
		ContentType: "application/json",
		Body: map[string]any{
			"paymentToken": "test_payment_token",
			"categoryCode": "QR",
			"groupCode":    "SGQR",
		},
	})
}

func TestNewDoPaymentRequest(t *testing.T) {
	client := NewClient("your_secret_key", "JT01", "https://example.com")

	params := &DoPaymentParams{
		PaymentToken:       "test_payment_token",
		PaymentChannelCode: "PNQR",
		PaymentData:        map[string]any{"qrType": "URL"},
		Locale:             "en",
		ResponseReturnUrl:  "https://merchant.com/callback",
		ClientID:           "test-client-id",
		ClientIP:           "192.168.1.1",
	}

	// Create request
	httpReq, err := client.newDoPaymentRequest(ctx, params)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Verify request using testutil
	testutil.AssertRequest(t, httpReq, struct {
		Method      string
		URL         string
		ContentType string
		Headers     map[string]string
		Body        any
	}{
		Method:      "POST",
		URL:         "https://example.com/payment/4.3/payment",
		ContentType: "application/json",
		Body: map[string]any{
			"paymentToken":      "test_payment_token",
			"locale":            "en",
			"responseReturnUrl": "https://merchant.com/callback",
			"clientID":          "test-client-id",
			"clientIP":          "192.168.1.1",
			"payment": map[string]any{
				"code": map[string]string{
					"channelCode": "PNQR",
				},
				"data": map[string]any{
					"qrType": "URL",
				},
			},
		},
	})

	// Test without optional fields
	minimalParams := &DoPaymentParams{
		PaymentToken:       "test_payment_token",
		PaymentChannelCode: "whatever",
		PaymentData:        map[string]any{"key": "value"},
		Locale:             "en",
		ResponseReturnUrl:  "https://merchant.com/callback",
	}

	// Create request with minimal params
	minimalReq, err := client.newDoPaymentRequest(ctx, minimalParams)
	if err != nil {
		t.Fatalf("Failed to create request with minimal params: %v", err)
	}

	// Verify minimal request
	testutil.AssertRequest(t, minimalReq, struct {
		Method      string
		URL         string
		ContentType string
		Headers     map[string]string
		Body        any
	}{
		Method:      "POST",
		URL:         "https://example.com/payment/4.3/payment",
		ContentType: "application/json",
		Body: map[string]any{
			"paymentToken":      "test_payment_token",
			"locale":            "en",
			"responseReturnUrl": "https://merchant.com/callback",
			"payment": map[string]any{
				"code": map[string]string{
					"channelCode": "whatever",
				},
				"data": map[string]any{"key": "value"},
			},
		},
	})
}
