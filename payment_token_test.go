package api2c2p

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/choonkeat/2c2p/testutil"
)

func TestPaymentTokenRequest_SignatureString(t *testing.T) {
	req := &PaymentTokenRequest{
		MerchantID:                    "JT01",
		CurrencyCodeISO4217:           "THB",
		AmountCents:                   10000,
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

	if unmarshaled.AmountCents != req.AmountCents {
		t.Errorf("Expected amount %#v, got %#v", req.AmountCents, unmarshaled.AmountCents)
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
		AmountCents:         9910,
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

	if unmarshaled.AmountCents != req.AmountCents {
		t.Errorf("Expected amount %#v, got %#v", req.AmountCents, unmarshaled.AmountCents)
	}

}

func TestNewPaymentTokenRequest(t *testing.T) {
	client := NewClient("your_secret_key", "JT01", "https://example.com")
	req := &PaymentTokenRequest{
		MerchantID:          "JT01",
		InvoiceNo:           "INV123",
		Description:         "Test payment",
		AmountCents:         10050,
		Request3DS:          "Y",
		CurrencyCodeISO4217: "SGD",
		PaymentChannel:      []PaymentTokenPaymentChannel{"CC"},
	}

	// Create request
	httpReq, err := client.newPaymentTokenRequest(ctx, req)
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
			"amount":         "000000000100.50000",
			"currencyCode":   "SGD",
			"request3DS":     "Y",
			"paymentChannel": []string{"CC"},
		},
	})
}

func TestCentsJSON(t *testing.T) {
	testCases := []struct {
		name     string
		cents    Cents
		wantJSON string
		wantErr  bool
	}{
		{
			name:     "zero value",
			cents:    Cents(0),
			wantJSON: `"000000000000.00000"`,
		},
		{
			name:     "positive value",
			cents:    Cents(1234),
			wantJSON: `"000000000012.34000"`,
		},
		{
			name:     "large value",
			cents:    Cents(123456789012),
			wantJSON: `"001234567890.12000"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test marshaling
			got, err := json.Marshal(tc.cents)
			if (err != nil) != tc.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && string(got) != tc.wantJSON {
				t.Errorf("MarshalJSON() = %v, want %v", string(got), tc.wantJSON)
			}

			// Test unmarshaling
			if !tc.wantErr {
				var c Cents
				err := json.Unmarshal([]byte(tc.wantJSON), &c)
				if err != nil {
					t.Errorf("UnmarshalJSON() error = %v", err)
					return
				}
				if c != tc.cents {
					t.Errorf("UnmarshalJSON() = %v, want %v", c, tc.cents)
				}
			}
		})
	}
}

func TestCentsUnmarshalJSONErrors(t *testing.T) {
	testCases := []struct {
		name    string
		json    string
		wantErr string
	}{
		{
			name:    "invalid format - missing decimal",
			json:    `"000000001234"`,
			wantErr: "invalid format",
		},
		{
			name:    "invalid format - not a number",
			json:    `"abcdef.12000"`,
			wantErr: "strconv.ParseInt",
		},
		{
			name:    "invalid format - decimal part not a number",
			json:    `"000000001234.abcde"`,
			wantErr: "strconv.ParseInt",
		},
		{
			name:    "invalid json",
			json:    `not_json`,
			wantErr: "invalid character",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var c Cents
			err := json.Unmarshal([]byte(tc.json), &c)
			if err == nil {
				t.Error("UnmarshalJSON() expected error, got nil")
				return
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("UnmarshalJSON() error = %v, want error containing %v", err, tc.wantErr)
			}
		})
	}
}
