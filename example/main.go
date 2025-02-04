package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	api2c2p "github.com/choonkeat/2c2p"
)

func main() {
	// Define command line flags
	function := flag.String("function", "", "Function to call (e.g., PaymentInquiry)")
	merchantID := flag.String("merchantID", "", "Merchant ID")
	invoiceNo := flag.String("invoiceNo", "", "Invoice number")
	locale := flag.String("locale", "en", "Locale (optional)")
	baseURL := flag.String("baseURL", "https://sandbox-pgw.2c2p.com", "Base URL for API")
	secretKey := flag.String("secretKey", "", "Secret key for JWT signing")

	flag.Parse()

	// Validate required flags
	if *function == "" || *merchantID == "" || *secretKey == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Create client
	client := api2c2p.NewClient(*merchantID, *secretKey, *baseURL)

	// Execute requested function
	switch *function {
	case "PaymentInquiry":
		if *invoiceNo == "" {
			fmt.Println("Error: invoiceNo is required for PaymentInquiry")
			flag.Usage()
			os.Exit(1)
		}

		request := &api2c2p.PaymentInquiryRequest{
			MerchantID: *merchantID,
			InvoiceNo:  *invoiceNo,
			Locale:     *locale,
		}

		response, err := client.PaymentInquiry(request)
		if err != nil {
			log.Fatalf("Error making payment inquiry: %v", err)
		}

		// Print response in a readable format
		fmt.Printf("Payment Inquiry Response:\n"+
			"MerchantID: %s\n"+
			"InvoiceNo: %s\n"+
			"Amount: %.2f\n"+
			"CurrencyCode: %s\n"+
			"TransactionDateTime: %s\n"+
			"RespCode: %s\n"+
			"RespDesc: %s\n",
			response.MerchantID,
			response.InvoiceNo,
			response.Amount,
			response.CurrencyCode,
			response.TransactionDateTime,
			response.RespCode,
			response.RespDesc,
		)

	default:
		fmt.Printf("Unknown function: %s\n", *function)
		flag.Usage()
		os.Exit(1)
	}
}