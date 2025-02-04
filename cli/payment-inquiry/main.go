package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	api2c2p "github.com/choonkeat/2c2p"
)

func main() {
	merchantID := flag.String("merchantID", "", "Merchant ID (required)")
	secretKey := flag.String("secretKey", "", "Secret Key (required)")
	invoiceNo := flag.String("invoiceNo", "", "Invoice Number (required if paymentToken not provided)")
	paymentToken := flag.String("paymentToken", "", "Payment Token (required if invoiceNo not provided)")
	locale := flag.String("locale", "en", "Locale (optional)")
	baseURL := flag.String("baseURL", "https://sandbox-pgw.2c2p.com", "API Base URL")
	flag.Parse()

	// Validate required flags
	if *merchantID == "" || *secretKey == "" {
		fmt.Println("Error: merchantID and secretKey are required")
		flag.Usage()
		os.Exit(1)
	}
	if *invoiceNo == "" && *paymentToken == "" {
		fmt.Println("Error: either invoiceNo or paymentToken must be provided")
		flag.Usage()
		os.Exit(1)
	}

	// Create client
	client := api2c2p.NewClient(*merchantID, *secretKey, *baseURL)

	// Create request
	request := &api2c2p.PaymentInquiryRequest{
		MerchantID:   *merchantID,
		InvoiceNo:    *invoiceNo,
		PaymentToken: *paymentToken,
		Locale:       *locale,
	}

	// Make request
	response, err := client.PaymentInquiry(request)
	if err != nil {
		log.Fatalf("Error making payment inquiry: %v", err)
	}

	// Print response
	fmt.Printf("Response Code: %s\n", response.RespCode)
	fmt.Printf("Description: %s\n", response.RespDesc)
	if response.RespCode == "0000" {
		fmt.Printf("Invoice No: %s\n", response.InvoiceNo)
		fmt.Printf("Amount: %.2f %s\n", response.Amount, response.CurrencyCode)
		fmt.Printf("Transaction Date: %s\n", response.TransactionDateTime)
		if response.PaymentScheme != "" {
			fmt.Printf("Payment Scheme: %s\n", response.PaymentScheme)
		}
		if response.IssuerBank != "" {
			fmt.Printf("Issuer Bank: %s\n", response.IssuerBank)
		}
	}
}
