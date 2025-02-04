package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/choonkeat/api2c2p"
)

func main() {
	var (
		merchantID    = flag.String("merchantID", "", "Merchant ID")
		secretKey     = flag.String("secretKey", "", "Secret Key")
		baseURL       = flag.String("baseURL", "https://sandbox-pgw.2c2p.com", "API Base URL")
		currencyCode  = flag.String("currencyCode", "", "Currency Code (e.g., THB)")
		amount        = flag.Float64("amount", 0, "Payment Amount")
		invoiceNo     = flag.String("invoiceNo", "", "Invoice Number")
		description   = flag.String("description", "", "Payment Description")
		locale        = flag.String("locale", "en", "Locale")
		frontendURL   = flag.String("frontendURL", "", "Frontend Return URL")
		backendURL    = flag.String("backendURL", "", "Backend Return URL")
		paymentExpiry = flag.String("paymentExpiry", "", "Payment Expiry (e.g., 2024-12-31 23:59:59)")
	)
	flag.Parse()

	if *merchantID == "" || *secretKey == "" || *currencyCode == "" || *amount == 0 || *invoiceNo == "" {
		fmt.Println("Required flags: -merchantID, -secretKey, -currencyCode, -amount, -invoiceNo")
		flag.Usage()
		os.Exit(1)
	}

	client := api2c2p.NewClient(*merchantID, *secretKey, *baseURL)

	req := &api2c2p.PaymentTokenRequest{
		MerchantID:    *merchantID,
		CurrencyCode:  *currencyCode,
		Amount:        *amount,
		InvoiceNo:     *invoiceNo,
		Description:   *description,
		Locale:        *locale,
		FrontendURL:   *frontendURL,
		BackendURL:    *backendURL,
		PaymentExpiry: *paymentExpiry,
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
