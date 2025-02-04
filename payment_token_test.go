package api2c2p

import (
	"encoding/json"
	"testing"
)

func TestPaymentTokenRequest_SignatureString(t *testing.T) {
	req := &PaymentTokenRequest{
		MerchantID:    "JT01",
		CurrencyCodeISO4217:  "THB",
		Amount:        100.00,
		Description:   "Test Payment",
		PaymentChannel: []PaymentChannel{PaymentChannelCC},
		Request3DS:    Request3DSType("Y"),
		CardTokens:    []string{"token1", "token2"},
		InvoiceNo:     "inv001",
		ProductCode:   "prod001",
		Recurring:     true,
		InvoicePrefix: "INV",
		RecurringAmount: 100.00,
		AllowAccumulate: true,
		MaxAccumulateAmount: 1000.00,
		RecurringIntervalDays: 30,
		RecurringCount: 12,
		ChargeNextDateYYYYMMDD: "2025-02-01",
		ChargeOnDateYYYYMMDD:  "2025-02-15",
		PaymentExpiryYYYYMMDDHHMMSS: "2025-02-04 23:59:59",
		PromotionCode: "PROMO1",
		PaymentRouteID: "route1",
		FxProviderCode: "fx1",
		ImmediatePayment: true,
		UserDefined1:  "user1",
		UserDefined2:  "user2",
		UserDefined3:  "user3",
		UserDefined4:  "user4",
		UserDefined5:  "user5",
		StatementDescriptor: "Test Payment",
		Locale:        "en",
		FrontendReturnURL: "https://example.com/frontend",
		BackendReturnURL:  "https://example.com/backend",
		NonceStr:      "abc123",
	}
	req.UIParams = &UIParams{
		UserInfo: &UserInfo{
			Name:           "John Doe",
			Email:          "john@example.com",
			MobileNo:       "1234567890",
			CountryCodeISO3166:    "SG",
			MobileNoPrefix: "65",
			CurrencyCodeISO4217:   "SGD",
		},
	}

	// Test request marshaling
	payload, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var unmarshaled PaymentTokenRequest
	if err := json.Unmarshal(payload, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	// Test field values
	if unmarshaled.MerchantID != req.MerchantID {
		t.Errorf("Expected MerchantID %s, got %s", req.MerchantID, unmarshaled.MerchantID)
	}

	if unmarshaled.CurrencyCodeISO4217 != req.CurrencyCodeISO4217 {
		t.Errorf("Expected CurrencyCodeISO4217 %s, got %s", req.CurrencyCodeISO4217, unmarshaled.CurrencyCodeISO4217)
	}

	if unmarshaled.Amount != req.Amount {
		t.Errorf("Expected Amount %.2f, got %.2f", req.Amount, unmarshaled.Amount)
	}

	if unmarshaled.RecurringIntervalDays != req.RecurringIntervalDays {
		t.Errorf("Expected RecurringIntervalDays %d, got %d", req.RecurringIntervalDays, unmarshaled.RecurringIntervalDays)
	}

	if unmarshaled.ChargeNextDateYYYYMMDD != req.ChargeNextDateYYYYMMDD {
		t.Errorf("Expected ChargeNextDateYYYYMMDD %s, got %s", req.ChargeNextDateYYYYMMDD, unmarshaled.ChargeNextDateYYYYMMDD)
	}

	if unmarshaled.PaymentExpiryYYYYMMDDHHMMSS != req.PaymentExpiryYYYYMMDDHHMMSS {
		t.Errorf("Expected PaymentExpiryYYYYMMDDHHMMSS %s, got %s", req.PaymentExpiryYYYYMMDDHHMMSS, unmarshaled.PaymentExpiryYYYYMMDDHHMMSS)
	}

	if unmarshaled.UIParams.UserInfo.CountryCodeISO3166 != req.UIParams.UserInfo.CountryCodeISO3166 {
		t.Errorf("Expected CountryCodeISO3166 %s, got %s", req.UIParams.UserInfo.CountryCodeISO3166, unmarshaled.UIParams.UserInfo.CountryCodeISO3166)
	}
}

func TestPaymentTokenRequest_SignatureString_Simple(t *testing.T) {
	req := &PaymentTokenRequest{
		MerchantID:    "JT01",
		CurrencyCodeISO4217:  "THB",
		Amount:        100.00,
		Description:   "Test Payment",
		InvoiceNo:     "inv001",
	}

	// Test JSON marshaling
	payload, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var unmarshaled PaymentTokenRequest
	if err := json.Unmarshal(payload, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	// Test field values
	if unmarshaled.MerchantID != req.MerchantID {
		t.Errorf("Expected MerchantID %s, got %s", req.MerchantID, unmarshaled.MerchantID)
	}

	if unmarshaled.CurrencyCodeISO4217 != req.CurrencyCodeISO4217 {
		t.Errorf("Expected CurrencyCodeISO4217 %s, got %s", req.CurrencyCodeISO4217, unmarshaled.CurrencyCodeISO4217)
	}

	if unmarshaled.Amount != req.Amount {
		t.Errorf("Expected Amount %.2f, got %.2f", req.Amount, unmarshaled.Amount)
	}

	if unmarshaled.Description != req.Description {
		t.Errorf("Expected Description %s, got %s", req.Description, unmarshaled.Description)
	}

	if unmarshaled.InvoiceNo != req.InvoiceNo {
		t.Errorf("Expected InvoiceNo %s, got %s", req.InvoiceNo, unmarshaled.InvoiceNo)
	}
}
