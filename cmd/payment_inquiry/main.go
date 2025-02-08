package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	api2c2p "github.com/choonkeat/2c2p"
)

func main() {
	var (
		merchantID   = flag.String("merchantID", "", "2C2P merchant ID")
		secretKey    = flag.String("secretKey", "", "Secret Key")
		invoiceNo    = flag.String("invoiceNo", "", "Invoice number to query")
		paymentToken = flag.String("paymentToken", "", "Payment token to query")
	)
	flag.Parse()

	if *merchantID == "" || *secretKey == "" {
		fmt.Println("Required flags: -merchantID, -secretKey")
		flag.Usage()
		os.Exit(1)
	}

	if *invoiceNo != "" && *paymentToken != "" {
		log.Fatal("Cannot specify both -invoiceNo and -paymentToken")
	}

	if *invoiceNo == "" && *paymentToken == "" {
		fmt.Println("Required flags: -invoiceNo or -paymentToken")
		flag.Usage()
		os.Exit(1)
	}

	client := api2c2p.NewClient(api2c2p.Config{
		SecretKey:  *secretKey,
		MerchantID: *merchantID,
	})
	var resp interface{}
	var err error

	ctx := context.Background()
	if *paymentToken != "" {
		resp, err = client.PaymentInquiryByToken(ctx, &api2c2p.PaymentInquiryByTokenRequest{
			PaymentToken: *paymentToken,
		})
	} else {
		resp, err = client.PaymentInquiryByInvoice(ctx, &api2c2p.PaymentInquiryByInvoiceRequest{
			InvoiceNo: *invoiceNo,
		})
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintf(os.Stderr, "resp: %#v\n", resp)
		os.Exit(1)
	}

	paymentResponse, ok := resp.(*api2c2p.PaymentInquiryResponse)
	if !ok {
		log.Fatal("Unexpected response type")
	}

	fmt.Printf("Response Code: %s\n", paymentResponse.RespCode)
	fmt.Printf("Response Description: %s\n", api2c2p.PaymentResponseCode(paymentResponse.RespCode).Description())
	fmt.Printf("Transaction Status: %s\n", paymentResponse.TransactionStatus)
	fmt.Printf("Amount: %.2f\n", paymentResponse.Amount)
	fmt.Printf("Currency Code: %s\n", paymentResponse.CurrencyCode)
	fmt.Printf("Masked Pan: %s\n", paymentResponse.MaskedPan)
	fmt.Printf("Payment Channel: %s\n", paymentResponse.PaymentChannel)
	fmt.Printf("Payment Status: %s\n", paymentResponse.PaymentStatus)
	fmt.Printf("Channel Response Code: %s\n", paymentResponse.ChannelResponseCode)
	fmt.Printf("Channel Response Description: %s\n", paymentResponse.ChannelResponseDescription)
	fmt.Printf("Approval Code: %s\n", paymentResponse.ApprovalCode)
	fmt.Printf("ECI: %s\n", paymentResponse.ECI)
	fmt.Printf("Transaction DateTime: %s\n", paymentResponse.TransactionDateTime)
	fmt.Printf("Paid Agent: %s\n", paymentResponse.PaidAgent)
	fmt.Printf("Paid Channel: %s\n", paymentResponse.PaidChannel)
	fmt.Printf("Paid DateTime: %s\n", paymentResponse.PaidDateTime)
	fmt.Printf("User Defined 1: %s\n", paymentResponse.UserDefined1)
	fmt.Printf("User Defined 2: %s\n", paymentResponse.UserDefined2)
	fmt.Printf("User Defined 3: %s\n", paymentResponse.UserDefined3)
	fmt.Printf("User Defined 4: %s\n", paymentResponse.UserDefined4)
	fmt.Printf("User Defined 5: %s\n", paymentResponse.UserDefined5)
}
