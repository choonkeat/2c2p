package api2c2p

import (
	"os"
	"testing"
)

func TestDecryptPKCS7WithED25519(t *testing.T) {
	// Sample ED25519 private key in PEM format
	// read from file /Users/choonkeatchew/Downloads/demo2/demo2.pem
	privateKey, err := os.ReadFile("combined.pem")
	if err != nil {
		t.Errorf("os.ReadFile: %v", err)
	}

	// Your test PKCS7 data
	encryptedData := "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"
	decrypted, err := Decrypt([]byte(encryptedData), privateKey)
	if err != nil {
		t.Errorf("Failed to decrypt PKCS7 data: %v", err)
	}

	// Add assertions based on expected decrypted data
	if len(decrypted) == 0 {
		t.Error("Expected non-empty decrypted data")
	}
	if want, got := "<PaymentResponse><version>9.4</version><timeStamp>060225004633</timeStamp><merchantID>702702000003987</merchantID><respCode>99</respCode><pan>411111XXXXXX1111</pan><amt>000000010010</amt><uniqueTransactionCode>INV1738777591</uniqueTransactionCode><tranRef></tranRef><approvalCode></approvalCode><refNumber></refNumber><eci></eci><dateTime>060225014633</dateTime><status>F</status><failReason>Invalid Payment Currency.</failReason><userDefined1></userDefined1><userDefined2></userDefined2><userDefined3></userDefined3><userDefined4></userDefined4><userDefined5></userDefined5><ippPeriod></ippPeriod><ippInterestType></ippInterestType><ippInterestRate></ippInterestRate><ippMerchantAbsorbRate></ippMerchantAbsorbRate><paidChannel></paidChannel><paidAgent></paidAgent><paymentChannel></paymentChannel><backendInvoice></backendInvoice><issuerCountry>US</issuerCountry><issuerCountryA3>USA</issuerCountryA3><bankName>FIRST DATA CORPORATIONS</bankName><cardType>CREDIT</cardType><processBy>VI</processBy><paymentScheme>VI</paymentScheme><paymentID></paymentID><acquirerResponseCode></acquirerResponseCode><schemePaymentID></schemePaymentID><hashValue>B2F826D0615D3F551A4B1708F516EF274D816F65</hashValue></PaymentResponse>", string(decrypted); want != got {
		t.Errorf("Expected %q, got %q", want, got)
	}
}
