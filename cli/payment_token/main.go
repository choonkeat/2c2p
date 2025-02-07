package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	api2c2p "github.com/choonkeat/2c2p"
)

func main() {
	var (
		secretKey           = flag.String("secretKey", "", "Merchant's secret key")
		merchantID          = flag.String("merchantID", "", "Merchant ID")
		amountCents         = flag.Int64("amountCents", 0, "Payment amount in cents")
		invoiceNo           = flag.String("invoiceNo", "", "Invoice number")
		description         = flag.String("description", "", "Payment description")
		currencyCodeISO4217 = flag.String("currencyCode", "", "Currency code (ISO 4217)")

		idempotencyID                    = flag.String("idempotencyID", "", "Unique value for retrying same requests")
		paymentChannelStr                = flag.String("paymentChannel", string(api2c2p.PaymentChannelCC), "Payment channel (comma-separated list)")
		agentChannelStr                  = flag.String("agentChannel", "", "Agent channel (comma-separated list)")
		request3DS                       = flag.String("request3DS", string(api2c2p.Request3DSYes), "Request 3DS (Y/N/F)")
		protocolVersion                  = flag.String("protocolVersion", "", "3DS protocol version")
		eci                              = flag.String("eci", "", "Electronic Commerce Indicator")
		cavv                             = flag.String("cavv", "", "Cardholder Authentication Verification Value")
		dsTransactionID                  = flag.String("dsTransactionID", "", "Directory Server Transaction ID")
		tokenize                         = flag.Bool("tokenize", false, "Enable tokenization")
		cardTokenStr                     = flag.String("cardTokens", "", "Comma-separated list of card tokens")
		tokenizeOnly                     = flag.Bool("tokenizeOnly", false, "Only tokenize, don't process payment")
		storeCredentials                 = flag.String("storeCredentials", "", "Store credentials (F/S/N)")
		interestType                     = flag.String("interestType", "", "Interest type (A/C/M)")
		installmentPeriodFilterMonthsStr = flag.String("installmentPeriod", "", "Comma-separated list of installment periods in months")
		installmentBankFilterStr         = flag.String("installmentBank", "", "Comma-separated list of installment banks")
		productCode                      = flag.String("productCode", "", "Product code")
		recurring                        = flag.Bool("recurring", false, "Enable recurring payment")
		invoicePrefix                    = flag.String("invoicePrefix", "", "Invoice prefix for recurring payments")
		recurringAmount                  = flag.Float64("recurringAmount", 0, "Amount for recurring payments")
		allowAccumulate                  = flag.Bool("allowAccumulate", false, "Allow accumulation of recurring payments")
		maxAccumulateAmount              = flag.Float64("maxAccumulateAmount", 0, "Maximum amount for accumulated payments")
		recurringIntervalDays            = flag.Int("recurringInterval", 0, "Interval in days between recurring payments")
		recurringCount                   = flag.Int("recurringCount", 0, "Total number of recurring payments")
		chargeNextDateYYYYMMDD           = flag.String("chargeNextDate", "", "Next charge date (YYYYMMDD)")
		chargeOnDateYYYYMMDD             = flag.String("chargeOnDate", "", "Specific charge date (YYYYMMDD)")
		paymentExpiryYYYYMMDDHHMMSS      = flag.String("paymentExpiry", "", "Payment expiry (YYYY-MM-DD HH:mm:ss)")
		promotionCode                    = flag.String("promotionCode", "", "Promotion code")
		paymentRouteID                   = flag.String("paymentRouteID", "", "Payment route ID")
		fxProviderCode                   = flag.String("fxProviderCode", "", "Forex provider code")
		fxRateID                         = flag.String("fxRateID", "", "Forex rate ID")
		originalAmount                   = flag.Float64("originalAmount", 0, "Original currency amount")
		immediatePayment                 = flag.Bool("immediatePayment", false, "Trigger payment immediately")
		iframeMode                       = flag.Bool("iframeMode", false, "Enable iframe mode")
		userDefined1                     = flag.String("userDefined1", "", "Custom field 1")
		userDefined2                     = flag.String("userDefined2", "", "Custom field 2")
		userDefined3                     = flag.String("userDefined3", "", "Custom field 3")
		userDefined4                     = flag.String("userDefined4", "", "Custom field 4")
		userDefined5                     = flag.String("userDefined5", "", "Custom field 5")
		statementDescriptor              = flag.String("statementDescriptor", "", "Dynamic statement description")
		externalSubMerchantID            = flag.String("externalSubMerchantID", "", "External sub-merchant ID")

		subMerchantID          = flag.String("subMerchantID", "", "Sub-merchant ID")
		subMerchantAmount      = flag.Float64("subMerchantAmount", 0, "Sub-merchant amount")
		subMerchantInvoiceNo   = flag.String("subMerchantInvoiceNo", "", "Sub-merchant invoice number")
		subMerchantDescription = flag.String("subMerchantDescription", "", "Sub-merchant description")
	)
	flag.Parse()

	if *secretKey == "" || *merchantID == "" || *invoiceNo == "" || *description == "" || *amountCents == 0 || *currencyCodeISO4217 == "" {
		flag.Usage()
		os.Exit(1)
	}

	client := api2c2p.NewClient(*secretKey, *merchantID)

	// Convert payment channels
	channels := strings.Split(*paymentChannelStr, ",")
	paymentChannels := make([]api2c2p.PaymentTokenPaymentChannel, len(channels))
	for i, ch := range channels {
		paymentChannels[i] = api2c2p.PaymentTokenPaymentChannel(ch)
	}

	// Convert agent channels
	var agentChannels []string
	if *agentChannelStr != "" {
		agentChannels = strings.Split(*agentChannelStr, ",")
	}

	// Convert card tokens
	var cardTokens []string
	if *cardTokenStr != "" {
		cardTokens = strings.Split(*cardTokenStr, ",")
	}

	// Convert installment periods
	var installmentPeriodFilterMonths []int
	if *installmentPeriodFilterMonthsStr != "" {
		for _, p := range strings.Split(*installmentPeriodFilterMonthsStr, ",") {
			var period int
			if _, err := fmt.Sscanf(p, "%d", &period); err == nil {
				installmentPeriodFilterMonths = append(installmentPeriodFilterMonths, period)
			}
		}
	}

	// Convert installment banks
	var installmentBanks []string
	if *installmentBankFilterStr != "" {
		installmentBanks = strings.Split(*installmentBankFilterStr, ",")
	}

	req := &api2c2p.PaymentTokenRequest{
		MerchantID:                    *merchantID,
		IdempotencyID:                 *idempotencyID,
		InvoiceNo:                     *invoiceNo,
		Description:                   *description,
		AmountCents:                   api2c2p.Cents(*amountCents),
		CurrencyCodeISO4217:           *currencyCodeISO4217,
		PaymentChannel:                paymentChannels,
		AgentChannel:                  agentChannels,
		Request3DS:                    api2c2p.PaymentTokenRequest3DSType(*request3DS),
		ProtocolVersion:               *protocolVersion,
		ECI:                           *eci,
		CAVV:                          *cavv,
		DSTransactionID:               *dsTransactionID,
		Tokenize:                      *tokenize,
		CardTokens:                    cardTokens,
		TokenizeOnly:                  *tokenizeOnly,
		StoreCredentials:              *storeCredentials,
		InterestType:                  api2c2p.PaymentTokenInterestType(*interestType),
		InstallmentPeriodFilterMonths: installmentPeriodFilterMonths,
		InstallmentBankFilter:         installmentBanks,
		ProductCode:                   *productCode,
		Recurring:                     *recurring,
		InvoicePrefix:                 *invoicePrefix,
		RecurringAmount:               *recurringAmount,
		AllowAccumulate:               *allowAccumulate,
		MaxAccumulateAmount:           *maxAccumulateAmount,
		RecurringIntervalDays:         *recurringIntervalDays,
		RecurringCount:                *recurringCount,
		ChargeNextDateYYYYMMDD:        *chargeNextDateYYYYMMDD,
		ChargeOnDateYYYYMMDD:          *chargeOnDateYYYYMMDD,
		PaymentExpiryYYYYMMDDHHMMSS:   *paymentExpiryYYYYMMDDHHMMSS,
		PromotionCode:                 *promotionCode,
		PaymentRouteID:                *paymentRouteID,
		FxProviderCode:                *fxProviderCode,
		FXRateID:                      *fxRateID,
		OriginalAmount:                *originalAmount,
		ImmediatePayment:              *immediatePayment,
		IframeMode:                    *iframeMode,
		UserDefined1:                  *userDefined1,
		UserDefined2:                  *userDefined2,
		UserDefined3:                  *userDefined3,
		UserDefined4:                  *userDefined4,
		UserDefined5:                  *userDefined5,
		StatementDescriptor:           *statementDescriptor,
		ExternalSubMerchantID:         *externalSubMerchantID,
	}

	// Add sub-merchant if provided
	if *subMerchantID != "" {
		if *subMerchantInvoiceNo == "" || *subMerchantAmount == 0 || *subMerchantDescription == "" {
			fmt.Fprintln(os.Stderr, "Sub-merchant requires all fields: ID, invoice number, amount, and description")
			os.Exit(1)
		}
		req.SubMerchants = []api2c2p.PaymentTokenSubMerchant{
			{
				MerchantID:  *subMerchantID,
				Amount:      *subMerchantAmount,
				InvoiceNo:   *subMerchantInvoiceNo,
				Description: *subMerchantDescription,
			},
		}
	}

	resp, err := client.PaymentToken(context.Background(), req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if resp != nil {
			fmt.Fprintf(os.Stderr, "Response Code: %s (%s)\n", resp.RespCode, api2c2p.PaymentResponseCode(resp.RespCode).Description())
		}
		os.Exit(1)
	}

	fmt.Printf("Response Code: %s\n", resp.RespCode)
	fmt.Printf("Response Description: %s\n", api2c2p.PaymentResponseCode(resp.RespCode).Description())
	fmt.Printf("Payment Token: %s\n", resp.PaymentToken)
	fmt.Printf("Web Payment URL: %s\n", resp.WebPaymentURL)
}
