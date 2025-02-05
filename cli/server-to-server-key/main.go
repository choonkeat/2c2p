package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

func main() {
	var (
		outDir = flag.String("out", ".", "output directory")
		cn     = flag.String("cn", "2C2P Test CA", "Common Name for the certificate")
	)
	flag.Parse()

	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: *cn,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0), // Valid for 10 years
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	// Write public certificate
	certOut, err := os.Create(*outDir + "/public_cert.pem")
	if err != nil {
		log.Fatalf("Failed to open public_cert.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("Failed to write data to public_cert.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing public_cert.pem: %v", err)
	}
	fmt.Printf("wrote %s/public_cert.pem\n", *outDir)

	// Write combined private key and certificate
	combinedOut, err := os.Create(*outDir + "/combined_private_public.pem")
	if err != nil {
		log.Fatalf("Failed to open combined_private_public.pem for writing: %v", err)
	}

	// Write private key
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatalf("Failed to marshal private key: %v", err)
	}
	if err := pem.Encode(combinedOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Failed to write data to combined_private_public.pem: %v", err)
	}

	// Write certificate
	if err := pem.Encode(combinedOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("Failed to write data to combined_private_public.pem: %v", err)
	}

	if err := combinedOut.Close(); err != nil {
		log.Fatalf("Error closing combined_private_public.pem: %v", err)
	}
	fmt.Printf("wrote %s/combined_private_public.pem\n", *outDir)
}
