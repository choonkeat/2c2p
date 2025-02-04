package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/choonkeat/2c2p"
)

func main() {
	var (
		merchantID    = flag.String("merchantID", "", "2C2P merchant ID")
		secretKey     = flag.String("secretKey", "", "Secret Key")
		baseURL       = flag.String("baseURL", "https://sandbox-pgw.2c2p.com", "API Base URL")
		currencyCodeISO4217  = flag.String("currencyCodeISO4217", "", "Currency code (ISO 4217)")
		amount        = flag.Float64("amount", 0, "Payment amount")
		invoiceNo     = flag.String("invoiceNo", "", "Invoice number")
		description   = flag.String("description", "", "Payment description")
		paymentChannelStr = flag.String("paymentChannel", "", "Comma-separated list of payment channels")
		request3DS    = flag.String("request3DS", string(api2c2p.Request3DSYes), "Request 3DS (Y/N)")
		tokenize      = flag.Bool("tokenize", false, "Enable tokenization")
		cardTokenStr  = flag.String("cardTokens", "", "Comma-separated list of card tokens")
		cardTokenOnly = flag.Bool("cardTokenOnly", false, "Use only card tokens")
		tokenizeOnly  = flag.Bool("tokenizeOnly", false, "Only tokenize, don't process payment")
		interestType  = flag.String("interestType", "", "Interest type for installment payments")
		installmentPeriodFilterMonthsStr = flag.String("installmentPeriodFilterMonths", "", "Comma-separated list of installment periods in months")
		productCode   = flag.String("productCode", "", "Product code")
		recurring     = flag.Bool("recurring", false, "Enable recurring payment")
		invoicePrefix = flag.String("invoicePrefix", "", "Invoice prefix for recurring payments")
		recurringAmount = flag.Float64("recurringAmount", 0, "Amount for recurring payments")
		allowAccumulate = flag.Bool("allowAccumulate", false, "Allow accumulation of recurring payments")
		maxAccumulateAmount = flag.Float64("maxAccumulateAmount", 0, "Maximum amount for accumulated payments")
		recurringIntervalDays = flag.Int("recurringIntervalDays", 0, "Interval in days between recurring payments")
		recurringCount = flag.Int("recurringCount", 0, "Total number of recurring payments")
		chargeNextDateYYYYMMDD = flag.String("chargeNextDateYYYYMMDD", "", "Next charge date (YYYY-MM-DD)")
		chargeOnDateYYYYMMDD = flag.String("chargeOnDateYYYYMMDD", "", "Specific charge date (YYYY-MM-DD)")
		paymentExpiryYYYYMMDDHHMMSS = flag.String("paymentExpiryYYYYMMDDHHMMSS", "", "Payment expiry (YYYY-MM-DD HH:mm:ss)")
		promotionCode = flag.String("promotionCode", "", "Promotion code")
		paymentRouteID = flag.String("paymentRouteID", "", "Payment route ID")
		fxProviderCode = flag.String("fxProviderCode", "", "FX provider code")
		immediatePayment = flag.Bool("immediatePayment", false, "Require immediate payment")
		userDefined1  = flag.String("userDefined1", "", "User defined field 1")
		userDefined2  = flag.String("userDefined2", "", "User defined field 2")
		userDefined3  = flag.String("userDefined3", "", "User defined field 3")
		userDefined4  = flag.String("userDefined4", "", "User defined field 4")
		userDefined5  = flag.String("userDefined5", "", "User defined field 5")
		statementDescriptor = flag.String("statementDescriptor", "", "Statement descriptor")
		locale        = flag.String("locale", "en", "Payment page language")
		frontendReturnURL = flag.String("frontendReturnURL", "", "Frontend return URL")
		backendReturnURL = flag.String("backendReturnURL", "", "Backend return URL")
		nonceStr      = flag.String("nonceStr", "", "Random string for request uniqueness")
		userName      = flag.String("userName", "", "Customer name")
		userEmail     = flag.String("userEmail", "", "Customer email")
		mobileNo      = flag.String("mobileNo", "", "Customer mobile number")
		countryCodeISO3166 = flag.String("countryCodeISO3166", "", "Customer country code (ISO 3166)")
		mobileNoPrefix = flag.String("mobileNoPrefix", "", "Customer mobile number prefix")
		currencyCodeISO4217UI = flag.String("currencyCodeISO4217UI", "", "Customer preferred currency (ISO 4217)")
	)
	flag.Parse()

	if *merchantID == "" || *secretKey == "" || *currencyCodeISO4217 == "" || *amount == 0 || *invoiceNo == "" {
		fmt.Println("Required flags: -merchantID, -secretKey, -currencyCodeISO4217, -amount, -invoiceNo")
		flag.Usage()
		os.Exit(1)
	}

	client := api2c2p.NewClient(*merchantID, *secretKey, *baseURL)

	// Convert payment channels string to slice
	var paymentChannels []api2c2p.PaymentChannel
	if *paymentChannelStr != "" {
		for _, ch := range strings.Split(*paymentChannelStr, ",") {
			paymentChannels = append(paymentChannels, api2c2p.PaymentChannel(strings.TrimSpace(ch)))
		}
	}

	// Convert card tokens string to slice
	var cardTokens []string
	if *cardTokenStr != "" {
		cardTokens = strings.Split(*cardTokenStr, ",")
		for i := range cardTokens {
			cardTokens[i] = strings.TrimSpace(cardTokens[i])
		}
	}

	// Convert installment periods string to slice
	var installmentPeriodFilterMonths []int
	if *installmentPeriodFilterMonthsStr != "" {
		for _, p := range strings.Split(*installmentPeriodFilterMonthsStr, ",") {
			period, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				log.Fatalf("Invalid installment period: %v", err)
			}
			installmentPeriodFilterMonths = append(installmentPeriodFilterMonths, period)
		}
	}

	req := &api2c2p.PaymentTokenRequest{
		MerchantID:    *merchantID,
		CurrencyCodeISO4217:  *currencyCodeISO4217,
		Amount:        *amount,
		InvoiceNo:     *invoiceNo,
		Description:   *description,
		PaymentChannel: paymentChannels,
		Request3DS:    api2c2p.Request3DSType(*request3DS),
		Tokenize:      *tokenize,
		CardTokens:    cardTokens,
		CardTokenOnly: *cardTokenOnly,
		TokenizeOnly:  *tokenizeOnly,
		InterestType:  api2c2p.InterestType(*interestType),
		InstallmentPeriodFilterMonths: installmentPeriodFilterMonths,
		ProductCode:   *productCode,
		Recurring:     *recurring,
		InvoicePrefix: *invoicePrefix,
		RecurringAmount: *recurringAmount,
		AllowAccumulate: *allowAccumulate,
		MaxAccumulateAmount: *maxAccumulateAmount,
		RecurringIntervalDays: *recurringIntervalDays,
		RecurringCount: *recurringCount,
		ChargeNextDateYYYYMMDD: *chargeNextDateYYYYMMDD,
		ChargeOnDateYYYYMMDD:  *chargeOnDateYYYYMMDD,
		PaymentExpiryYYYYMMDDHHMMSS: *paymentExpiryYYYYMMDDHHMMSS,
		PromotionCode: *promotionCode,
		PaymentRouteID: *paymentRouteID,
		FxProviderCode: *fxProviderCode,
		ImmediatePayment: *immediatePayment,
		UserDefined1:  *userDefined1,
		UserDefined2:  *userDefined2,
		UserDefined3:  *userDefined3,
		UserDefined4:  *userDefined4,
		UserDefined5:  *userDefined5,
		StatementDescriptor: *statementDescriptor,
		Locale:        *locale,
		FrontendReturnURL: *frontendReturnURL,
		BackendReturnURL:  *backendReturnURL,
		NonceStr:      *nonceStr,
	}

	// Add UI params if user info is provided
	if *userName != "" || *userEmail != "" || *mobileNo != "" || *countryCodeISO3166 != "" || *mobileNoPrefix != "" || *currencyCodeISO4217UI != "" {
		req.UIParams = &api2c2p.UIParams{
			UserInfo: &api2c2p.UserInfo{
				Name:           *userName,
				Email:          *userEmail,
				MobileNo:       *mobileNo,
				CountryCodeISO3166:    *countryCodeISO3166,
				MobileNoPrefix: *mobileNoPrefix,
				CurrencyCodeISO4217:   *currencyCodeISO4217UI,
			},
		}
	}

	resp, err := client.PaymentToken(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Response Code: %s\n", resp.RespCode)
	fmt.Printf("Response Description: %s\n", resp.RespDesc)
	fmt.Printf("Payment Token: %s\n", resp.PaymentToken)
	fmt.Printf("Web Payment URL: %s\n", resp.WebPaymentURL)
}
