package api2c2p

import (
	"encoding/json"
	"testing"
)

func TestPaymentTokenRequest_SignatureString(t *testing.T) {
	req := &PaymentTokenRequest{
		MerchantID:                    "JT01",
		CurrencyCodeISO4217:           "THB",
		Amount:                        100.00,
		Description:                   "Test Payment",
		PaymentChannel:                []PaymentChannel{PaymentChannelCC},
		AgentChannel:                  []string{},
		Request3DS:                    Request3DSType("Y"),
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
	req.UIParams = &UIParams{
		UserInfo: &UserInfo{
			Name:               "John Doe",
			Email:              "john@example.com",
			MobileNo:           "0123456789",
			CountryCodeISO3166: "TH",
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
		t.Errorf("Expected MerchantID %s, got %s", req.MerchantID, unmarshaled.MerchantID)
	}

	if unmarshaled.CurrencyCodeISO4217 != req.CurrencyCodeISO4217 {
		t.Errorf("Expected CurrencyCodeISO4217 %s, got %s", req.CurrencyCodeISO4217, unmarshaled.CurrencyCodeISO4217)
	}

	if unmarshaled.Amount != req.Amount {
		t.Errorf("Expected Amount %f, got %f", req.Amount, unmarshaled.Amount)
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
		t.Errorf("Expected UserInfo.CountryCodeISO3166 %s, got %s", req.UIParams.UserInfo.CountryCodeISO3166, unmarshaled.UIParams.UserInfo.CountryCodeISO3166)
	}
}

func TestPaymentTokenRequest_SignatureString_Simple(t *testing.T) {
	req := &PaymentTokenRequest{
		MerchantID:          "JT01",
		CurrencyCodeISO4217: "THB",
		Amount:              100.00,
		Description:         "Test Payment",
		InvoiceNo:           "inv001",
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
		t.Errorf("Expected MerchantID %s, got %s", req.MerchantID, unmarshaled.MerchantID)
	}

	if unmarshaled.CurrencyCodeISO4217 != req.CurrencyCodeISO4217 {
		t.Errorf("Expected CurrencyCodeISO4217 %s, got %s", req.CurrencyCodeISO4217, unmarshaled.CurrencyCodeISO4217)
	}

	if unmarshaled.Amount != req.Amount {
		t.Errorf("Expected Amount %f, got %f", req.Amount, unmarshaled.Amount)
	}

	if unmarshaled.InvoiceNo != req.InvoiceNo {
		t.Errorf("Expected InvoiceNo %s, got %s", req.InvoiceNo, unmarshaled.InvoiceNo)
	}
}

func TestPaymentTokenRequest_ToMap(t *testing.T) {
	req := PaymentTokenRequest{
		MerchantID:                    "merchant123",
		InvoiceNo:                     "inv123",
		Description:                   "Test payment",
		Amount:                        100.50,
		CurrencyCodeISO4217:           "THB",
		PaymentChannel:                []PaymentChannel{PaymentChannelCC},
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
		SubMerchants: []SubMerchant{
			{
				MerchantID:  "sub123",
				Amount:      50.25,
				InvoiceNo:   "subinv123",
				Description: "Sub merchant payment",
			},
		},
	}

	m := req.ToMap()

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
