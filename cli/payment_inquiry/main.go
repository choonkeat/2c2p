package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	api2c2p "github.com/choonkeat/2c2p"
)

func main() {
	var (
		merchantID = flag.String("merchantID", "", "2C2P merchant ID")
		secretKey  = flag.String("secretKey", "", "Secret Key")
		invoiceNo  = flag.String("invoiceNo", "", "Invoice number")
	)
	flag.Parse()

	if *merchantID == "" || *secretKey == "" || *invoiceNo == "" {
		fmt.Println("Required flags: -merchantID, -secretKey, -invoiceNo")
		flag.Usage()
		os.Exit(1)
	}

	client := api2c2p.NewClient(api2c2p.Config{
		SecretKey:  *secretKey,
		MerchantID: *merchantID,
	})
	resp, err := client.PaymentInquiry(context.Background(), &api2c2p.PaymentInquiryRequest{
		InvoiceNo: *invoiceNo,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if resp != nil {
			fmt.Fprintf(os.Stderr, "Response Code: %s (%s)\n", resp.RespCode, api2c2p.PaymentResponseCode(resp.RespCode).Description())
		}
		os.Exit(1)
	}

	fmt.Printf("Response Code: %s\n", resp.RespCode)
	fmt.Printf("Response Description: %s\n", api2c2p.PaymentResponseCode(resp.RespCode).Description())
	fmt.Printf("Transaction Status: %s\n", resp.TransactionStatus)
	fmt.Printf("Amount: %.2f\n", resp.Amount)
	fmt.Printf("Currency Code: %s\n", resp.CurrencyCode)
	fmt.Printf("Masked Pan: %s\n", resp.MaskedPan)
	fmt.Printf("Payment Channel: %s\n", resp.PaymentChannel)
	fmt.Printf("Payment Status: %s\n", resp.PaymentStatus)
	fmt.Printf("Channel Response Code: %s\n", resp.ChannelResponseCode)
	fmt.Printf("Channel Response Description: %s\n", resp.ChannelResponseDescription)
	fmt.Printf("Approval Code: %s\n", resp.ApprovalCode)
	fmt.Printf("ECI: %s\n", resp.ECI)
	fmt.Printf("Transaction DateTime: %s\n", resp.TransactionDateTime)
	fmt.Printf("Paid Agent: %s\n", resp.PaidAgent)
	fmt.Printf("Paid Channel: %s\n", resp.PaidChannel)
	fmt.Printf("Paid DateTime: %s\n", resp.PaidDateTime)
	fmt.Printf("User Defined 1: %s\n", resp.UserDefined1)
	fmt.Printf("User Defined 2: %s\n", resp.UserDefined2)
	fmt.Printf("User Defined 3: %s\n", resp.UserDefined3)
	fmt.Printf("User Defined 4: %s\n", resp.UserDefined4)
	fmt.Printf("User Defined 5: %s\n", resp.UserDefined5)
}
