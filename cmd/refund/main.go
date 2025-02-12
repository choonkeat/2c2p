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
		merchantID             = flag.String("merchantID", "", "Merchant ID")
		secretKey              = flag.String("secretKey", "", "Secret Key")
		invoiceNo              = flag.String("invoiceNo", "", "Invoice number of the transaction to refund")
		amountCents            = flag.Int64("amountCents", 0, "Amount to refund in cents")
		processType            = flag.String("processType", "R", "Process type (R for refund, V for void/cancel)")
		combinedPem            = flag.String("combinedPem", "dist/combined_private_public.pem", "Path to combined private key and certificate PEM file generated by cmd/server-to-server-key/main.go")
		serverJWTPublicKeyFile = flag.String("serverJWTPublicKey", "dist/sandbox-jwt-2c2p.demo.2.1(public).cer", "Path to 2C2P's public key certificate (.cer file)")
		serverPKCS7PublicKey   = flag.String("serverPKCS7PublicKey", "dist/sandbox-pkcs7-demo2.2c2p.com(public).cer", "Path to 2C2P's public key certificate (.cer file)")
		paymentGatewayURL      = flag.String("paymentGatewayURL", "https://sandbox-pgw.2c2p.com", "2C2P Payment Gateway URL")
		frontendURL            = flag.String("frontendURL", "https://demo2.2c2p.com", "2C2P Frontend URL")
	)
	flag.Parse()

	// Validate required flags
	if *merchantID == "" || *secretKey == "" || *invoiceNo == "" || *amountCents <= 0 || (*processType != "R" && *processType != "V") {
		flag.Usage()
		os.Exit(1)
	}

	// Create client
	client, err := api2c2p.NewClient(api2c2p.Config{
		SecretKey:                *secretKey,
		MerchantID:               *merchantID,
		PaymentGatewayURL:        *paymentGatewayURL,
		FrontendURL:              *frontendURL,
		CombinedPEM:              *combinedPem,
		ServerJWTPublicKeyFile:   *serverJWTPublicKeyFile,
		ServerPKCS7PublicKeyFile: *serverPKCS7PublicKey,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	switch *processType {
	case "R":
		// Process refund
		resp, err := client.Refund(context.Background(), *invoiceNo, api2c2p.Cents(*amountCents))
		if err != nil {
			log.Fatalf("Failed to process refund: %v", err)
		}
		// Print response
		fmt.Printf("Response Code: %s\n", resp.RespCode)
		fmt.Printf("Response Description: %s\n", resp.RespDesc)

	case "V":
		// Process void/cancel
		resp, err := client.VoidCancel(context.Background(), *invoiceNo, api2c2p.Cents(*amountCents))
		if err != nil {
			log.Fatalf("Failed to process void: %v", err)
		}
		// Print response
		fmt.Printf("Response Code: %s\n", resp.RespCode)
		fmt.Printf("Response Description: %s\n", resp.RespDesc)
	}
}
