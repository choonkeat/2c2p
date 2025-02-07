package api2c2p

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/choonkeat/2c2p/testutil"
)

func TestPaymentTokenRequest_SignatureString(t *testing.T) {
	req := &PaymentTokenRequest{
		MerchantID:                    "JT01",
		CurrencyCodeISO4217:           "THB",
		Amount:                        100.00,
		Description:                   "Test Payment",
		PaymentChannel:                []PaymentTokenPaymentChannel{PaymentChannelCC},
		AgentChannel:                  []string{},
		Request3DS:                    PaymentTokenRequest3DSType("Y"),
		CardTokens:                    []string{},
		TokenizeOnly:                  false,
		StoreCredentials:              "",
		InterestType:                  "",
		InstallmentPeriodFilterMonths: []int{},
		InstallmentBankFilter:         []string{},
		InvoicePrefix:                 "",
		InvoiceNo:                     "inv001",
		ProductCode:                   "prod001",
		Recurring:                     true,
		RecurringAmount:               100.00,
		AllowAccumulate:               true,
		MaxAccumulateAmount:           1000.00,
		RecurringIntervalDays:         30,
		RecurringCount:                12,
		ChargeNextDateYYYYMMDD:        "20250201",
		ChargeOnDateYYYYMMDD:          "20250215",
		PaymentExpiryYYYYMMDDHHMMSS:   "2025-02-04 23:59:59",
		PromotionCode:                 "PROMO1",
		PaymentRouteID:                "route1",
		FxProviderCode:                "fx1",
		FXRateID:                      "",
		OriginalAmount:                0,
		ImmediatePayment:              true,
		IframeMode:                    false,
		UserDefined1:                  "user1",
		UserDefined2:                  "user2",
		UserDefined3:                  "user3",
		UserDefined4:                  "user4",
		UserDefined5:                  "user5",
		StatementDescriptor:           "Test Payment",
	}
	req.UIParams = &paymentTokenUiParams{
		UserInfo: &paymentTokenUserInfo{
			Name:                "John Doe",
			Email:               "john@example.com",
			MobileNo:            "0123456789",
			CountryCodeISO3166:  "TH",
			MobileNoPrefix:      "",
			CurrencyCodeISO4217: "",
			Address:             "",
		},
	}

	// Test JSON marshaling
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled PaymentTokenRequest
	if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal request: %v", err)
	}

	// Test field values
	if unmarshaled.MerchantID != req.MerchantID {
		t.Errorf("Expected merchantID %s, got %s", req.MerchantID, unmarshaled.MerchantID)
	}

	if unmarshaled.CurrencyCodeISO4217 != req.CurrencyCodeISO4217 {
		t.Errorf("Expected currencyCodeISO4217 %s, got %s", req.CurrencyCodeISO4217, unmarshaled.CurrencyCodeISO4217)
	}

	if unmarshaled.Amount != req.Amount {
		t.Errorf("Expected amount %f, got %f", req.Amount, unmarshaled.Amount)
	}

	if unmarshaled.RecurringIntervalDays != req.RecurringIntervalDays {
		t.Errorf("Expected recurringIntervalDays %d, got %d", req.RecurringIntervalDays, unmarshaled.RecurringIntervalDays)
	}

	if unmarshaled.ChargeNextDateYYYYMMDD != req.ChargeNextDateYYYYMMDD {
		t.Errorf("Expected chargeNextDateYYYYMMDD %s, got %s", req.ChargeNextDateYYYYMMDD, unmarshaled.ChargeNextDateYYYYMMDD)
	}

	if unmarshaled.PaymentExpiryYYYYMMDDHHMMSS != req.PaymentExpiryYYYYMMDDHHMMSS {
		t.Errorf("Expected paymentExpiryYYYYMMDDHHMMSS %s, got %s", req.PaymentExpiryYYYYMMDDHHMMSS, unmarshaled.PaymentExpiryYYYYMMDDHHMMSS)
	}

	if unmarshaled.UIParams == nil {
		t.Error("Expected UIParams to not be nil")
	} else if unmarshaled.UIParams.UserInfo == nil {
		t.Error("Expected UIParams.UserInfo to not be nil")
	} else if !reflect.DeepEqual(unmarshaled.UIParams.UserInfo, req.UIParams.UserInfo) {
		t.Errorf("Expected UIParams.UserInfo to be %+v, got %+v", req.UIParams.UserInfo, unmarshaled.UIParams.UserInfo)
	}
}

func TestPaymentTokenRequest_SignatureString_Simple(t *testing.T) {
	req := &PaymentTokenRequest{
		MerchantID:          "JT01",
		InvoiceNo:           "1234567890",
		Description:         "Test payment",
		Amount:              99.10,
		CurrencyCodeISO4217: "702",
		UIParams: &paymentTokenUiParams{
			UserInfo: &paymentTokenUserInfo{
				Name:                "John Doe",
				Email:               "john@example.com",
				MobileNo:            "0123456789",
				CountryCodeISO3166:  "TH",
				MobileNoPrefix:      "",
				CurrencyCodeISO4217: "",
				Address:             "",
			},
		},
	}

	// Test JSON marshaling
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled PaymentTokenRequest
	if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal request: %v", err)
	}

	// Test field values
	if unmarshaled.MerchantID != req.MerchantID {
		t.Errorf("Expected merchantID %s, got %s", req.MerchantID, unmarshaled.MerchantID)
	}

	if unmarshaled.InvoiceNo != req.InvoiceNo {
		t.Errorf("Expected invoiceNo %s, got %s", req.InvoiceNo, unmarshaled.InvoiceNo)
	}

	if unmarshaled.Amount != req.Amount {
		t.Errorf("Expected amount %f, got %f", req.Amount, unmarshaled.Amount)
	}

	got := req.toMap()
	if got["merchantID"] != req.MerchantID {
		t.Errorf("Expected merchantID %s, got %s", req.MerchantID, got["merchantID"])
	}

	if got["invoiceNo"] != req.InvoiceNo {
		t.Errorf("Expected invoiceNo %s, got %s", req.InvoiceNo, got["invoiceNo"])
	}

	if got["currencyCode"] != req.CurrencyCodeISO4217 {
		t.Errorf("Expected currencyCode %s, got %s", req.CurrencyCodeISO4217, got["currencyCode"])
	}
}

func TestPaymentTokenRequest_ToMap(t *testing.T) {
	req := PaymentTokenRequest{
		MerchantID:                    "merchant123",
		InvoiceNo:                     "inv123",
		Description:                   "Test payment",
		Amount:                        100.50,
		CurrencyCodeISO4217:           "THB",
		PaymentChannel:                []PaymentTokenPaymentChannel{PaymentChannelCC},
		AgentChannel:                  []string{"agent1"},
		Request3DS:                    Request3DSYes,
		CardTokens:                    []string{"token1"},
		TokenizeOnly:                  true,
		StoreCredentials:              "F",
		InterestType:                  InterestTypeAll,
		InstallmentPeriodFilterMonths: []int{3, 6},
		InstallmentBankFilter:         []string{"bank1"},
		InvoicePrefix:                 "prefix",
		ProductCode:                   "prod123",
		Recurring:                     true,
		RecurringAmount:               90.00,
		AllowAccumulate:               true,
		MaxAccumulateAmount:           1000.00,
		RecurringIntervalDays:         30,
		RecurringCount:                12,
		ChargeNextDateYYYYMMDD:        "20250101",
		ChargeOnDateYYYYMMDD:          "20250101",
		PaymentExpiryYYYYMMDDHHMMSS:   "2025-01-01 23:59:59",
		PromotionCode:                 "PROMO123",
		PaymentRouteID:                "route123",
		FxProviderCode:                "FX123",
		FXRateID:                      "rate123",
		OriginalAmount:                95.00,
		ImmediatePayment:              true,
		IframeMode:                    true,
		UserDefined1:                  "user1",
		UserDefined2:                  "user2",
		UserDefined3:                  "user3",
		UserDefined4:                  "user4",
		UserDefined5:                  "user5",
		StatementDescriptor:           "STMT*DESC",
		SubMerchants: []PaymentTokenSubMerchant{
			{
				MerchantID:  "sub123",
				Amount:      50.25,
				InvoiceNo:   "subinv123",
				Description: "Sub merchant payment",
			},
		},
		UIParams: &paymentTokenUiParams{
			UserInfo: &paymentTokenUserInfo{
				Name:                "John Doe",
				Email:               "john@example.com",
				MobileNo:            "0123456789",
				CountryCodeISO3166:  "TH",
				MobileNoPrefix:      "",
				CurrencyCodeISO4217: "",
				Address:             "",
			},
		},
	}

	m := req.toMap()

	// Test a few key fields
	if m["merchantID"] != req.MerchantID {
		t.Errorf("merchantID = %v, want %v", m["merchantID"], req.MerchantID)
	}
	if m["invoiceNo"] != req.InvoiceNo {
		t.Errorf("invoiceNo = %v, want %v", m["invoiceNo"], req.InvoiceNo)
	}
	if m["currencyCode"] != req.CurrencyCodeISO4217 {
		t.Errorf("currencyCode = %v, want %v", m["currencyCode"], req.CurrencyCodeISO4217)
	}
}

func TestNewPaymentTokenRequest(t *testing.T) {
	client := NewClient("your_secret_key", "JT01", "https://example.com")
	req := &PaymentTokenRequest{
		MerchantID:          "JT01",
		InvoiceNo:           "INV123",
		Description:         "Test payment",
		Amount:              100.50,
		Request3DS:          "Y",
		CurrencyCodeISO4217: "SGD",
		PaymentChannel:      []PaymentTokenPaymentChannel{"CC"},
	}

	// Create request
	httpReq, err := client.newPaymentTokenRequest(context.Background(), req)
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
		URL:         "https://example.com/payment/4.3/paymentToken",
		ContentType: "application/json",
		Body: map[string]any{
			"merchantID":     "JT01",
			"invoiceNo":      "INV123",
			"description":    "Test payment",
			"amount":         100.5,
			"currencyCode":   "SGD",
			"request3DS":     "Y",
			"paymentChannel": []string{"CC"},
		},
	})
}
