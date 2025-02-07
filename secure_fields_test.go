package api2c2p

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/fullsailor/pkcs7"
)

// TestDecryptPaymentResponse tests the decryption of 2C2P payment response data
// using the test private key and certificate. It reads encrypted test data from
// testdata/payment-response-*.txt files and compares the decrypted result with
// the expected XML in testdata/payment-response-*.txt.xml files.
//
// If a test fails because the expected XML needs to be updated, the test will
// write the actual decrypted result to the .xml file, making it pass on the next run.
func TestDecryptPaymentResponse(t *testing.T) {
	// Read private key from testdata
	privateKey, err := os.ReadFile("testdata/combined_private_public.pem")
	if err != nil {
		t.Fatalf("Failed to read private key: %v", err)
	}

	// Find all encrypted test data files
	matches, err := filepath.Glob("testdata/payment-response-*.txt")
	if err != nil {
		t.Fatalf("Failed to find test files: %v", err)
	}

	for _, encryptedFile := range matches {
		// Skip .xml files
		if strings.HasSuffix(encryptedFile, ".xml") {
			continue
		}

		testName := filepath.Base(encryptedFile)
		t.Run(testName, func(t *testing.T) {
			// Read encrypted data
			encryptedData, err := os.ReadFile(encryptedFile)
			if err != nil {
				t.Fatalf("Failed to read encrypted data: %v", err)
			}

			// Decrypt the data
			got, err := DecryptPKCS7(encryptedData, privateKey)
			if err != nil {
				t.Fatalf("Failed to decrypt data: %v", err)
			}

			// Read expected XML result
			expectedFile := encryptedFile + ".xml"
			want, err := os.ReadFile(expectedFile)
			if err != nil {
				t.Fatalf("Failed to read expected XML: %v", err)
			}

			if string(want) != string(got) {
				t.Errorf("Decrypted result does not match %s.\nGot:\n%s\nWant:\n%s", expectedFile, got, string(want))
			}
		})
	}
}

func TestDecryptPaymentResponseWithXML(t *testing.T) {
	// Create a test payment response
	testResp := PaymentResponseBackEnd{
		RespCode: "0000",
	}
	xmlData, err := xml.Marshal(testResp)
	if err != nil {
		t.Fatalf("Failed to marshal XML: %v", err)
	}

	// Read test certificate
	certPEM, err := os.ReadFile("testdata/public_cert.pem")
	if err != nil {
		t.Fatalf("Failed to read public cert: %v", err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("Failed to decode PEM block")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Encrypt the XML data
	encrypted, err := pkcs7.Encrypt(xmlData, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	// Create form with encrypted data
	form := mockFormValuer{
		values: map[string]string{
			"paymentResponse": base64.StdEncoding.EncodeToString(encrypted),
		},
	}

	// Read private key for decryption
	privateKey, err := os.ReadFile("testdata/combined_private_public.pem")
	if err != nil {
		t.Fatalf("Failed to read private key: %v", err)
	}

	// Test successful decryption
	response, decrypted, err := DecryptPaymentResponseBackend(form, privateKey)
	if err != nil {
		t.Fatalf("DecryptPaymentResponse failed: %v", err)
	}

	// Verify decrypted response
	if len(decrypted) == 0 {
		t.Error("Expected non-empty decrypted response")
	}

	// Verify response fields
	if response.RespCode != testResp.RespCode {
		t.Errorf("Expected response code %q, got %q", testResp.RespCode, response.RespCode)
	}
	if response.XMLName.Local != "PaymentResponse" {
		t.Errorf("Expected XML root element 'PaymentResponse', got %q", response.XMLName.Local)
	}

	// Test with invalid form value
	invalidForm := mockFormValuer{
		values: map[string]string{
			"paymentResponse": "invalid base64",
		},
	}
	_, _, err = DecryptPaymentResponseBackend(invalidForm, []byte("test key"))
	if err == nil {
		t.Error("Expected error with invalid payment response")
	}

	// Test with empty form value
	emptyForm := mockFormValuer{
		values: map[string]string{},
	}
	_, _, err = DecryptPaymentResponseBackend(emptyForm, []byte("test key"))
	if err == nil {
		t.Error("Expected error with empty payment response")
	}
}

// mockFormValuer implements FormValuer for testing
type mockFormValuer struct {
	values map[string]string
}

func (m mockFormValuer) PostFormValue(key string) string {
	return m.values[key]
}

func TestCreatePaymentPayload(t *testing.T) {
	// Test inputs
	merchantID := "MERCHANT123"
	secretKey := "SECRET456"
	timestamp := "1707210770"
	invoiceNo := "INV1707210770"
	paymentDetails := SecureFieldsPaymentDetails{
		AmountCents:  9910,
		CurrencyCode: "702",
		Description:  "1 room for 2 nights",
		CustomerName: "John Doe",
		CountryCode:  "SG",
		StoreCard:    "Y",
		UserDefined1: "1",
		UserDefined2: "2",
		UserDefined3: "3",
		UserDefined4: "4",
		UserDefined5: "5",
	}
	form := mockFormValuer{
		values: map[string]string{
			"encryptedCardInfo": "ENCRYPTED_CARD_DATA",
		},
	}

	// Call function
	payload := CreateSecureFieldsPaymentPayload("http://localhost:8080", merchantID, secretKey, timestamp, invoiceNo, paymentDetails, form)

	// Decode base64
	xmlBytes, err := base64.StdEncoding.DecodeString(payload.FormFields["paymentRequest"])
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}
	xmlStr := string(xmlBytes)

	// Check XML structure
	expected := []string{
		"<PaymentRequest>",
		"<version>9.4</version>",
		"<timeStamp>" + timestamp + "</timeStamp>",
		"<merchantID>" + merchantID + "</merchantID>",
		"<uniqueTransactionCode>" + invoiceNo + "</uniqueTransactionCode>",
		"<desc>" + paymentDetails.Description + "</desc>",
		"<amt>000000009910</amt>",
		"<currencyCode>" + paymentDetails.CurrencyCode + "</currencyCode>",
		"<paymentChannel></paymentChannel>",
		"<panCountry>" + paymentDetails.CountryCode + "</panCountry>",
		"<cardholderName>" + paymentDetails.CustomerName + "</cardholderName>",
		"<request3DS>Y</request3DS>",
		"<storeCard>" + paymentDetails.StoreCard + "</storeCard>",
		"<encCardData>" + form.PostFormValue("encryptedCardInfo") + "</encCardData>",
		"<userDefined1>" + paymentDetails.UserDefined1 + "</userDefined1>",
		"<userDefined2>" + paymentDetails.UserDefined2 + "</userDefined2>",
		"<userDefined3>" + paymentDetails.UserDefined3 + "</userDefined3>",
		"<userDefined4>" + paymentDetails.UserDefined4 + "</userDefined4>",
		"<userDefined5>" + paymentDetails.UserDefined5 + "</userDefined5>",
		"</PaymentRequest>",
	}

	// Check that each expected element is present in the XML
	for _, exp := range expected {
		if !strings.Contains(xmlStr, exp) {
			t.Errorf("Expected XML to contain %q, but it didn't\nXML: %s", exp, xmlStr)
		}
	}

	// Verify HMAC signature
	// Extract secureHash from XML using regex since it's dynamically generated
	re := regexp.MustCompile(`<secureHash>([^<]+)</secureHash>`)
	matches := re.FindStringSubmatch(xmlStr)
	if len(matches) != 2 {
		t.Fatal("Could not find secureHash in XML")
	}
	secureHash := matches[1]

	// Calculate expected HMAC
	strToHash := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s",
		"9.4",                                   // version
		timestamp,                               // timestamp
		merchantID,                              // merchantID
		invoiceNo,                               // uniqueTransactionCode
		paymentDetails.Description,              // desc
		"000000009910",                          // amt
		paymentDetails.CurrencyCode,             // currencyCode
		"",                                      // paymentChannel
		"",                                      // storeCardUniqueID
		"",                                      // panBank
		paymentDetails.CountryCode,              // country
		paymentDetails.CustomerName,             // cardholderName
		"",                                      // cardholderEmail
		"",                                      // payCategoryID
		paymentDetails.UserDefined1,             // userDefined1
		paymentDetails.UserDefined2,             // userDefined2
		paymentDetails.UserDefined3,             // userDefined3
		paymentDetails.UserDefined4,             // userDefined4
		paymentDetails.UserDefined5,             // userDefined5
		paymentDetails.StoreCard,                // storeCard
		"",                                      // ippTransaction
		"",                                      // installmentPeriod
		"",                                      // interestType
		"",                                      // recurring
		"",                                      // invoicePrefix
		"",                                      // recurringAmount
		"",                                      // allowAccumulate
		"",                                      // maxAccumulateAmt
		"",                                      // recurringInterval
		"",                                      // recurringCount
		"",                                      // chargeNextDate
		"",                                      // promotion
		"Y",                                     // request3DS
		"",                                      // statementDescriptor
		"",                                      // agentCode
		"",                                      // channelCode
		"",                                      // paymentExpiry
		"",                                      // mobileNo
		"",                                      // tokenizeWithoutAuthorization
		form.PostFormValue("encryptedCardInfo"), // encryptedCardInfo
	)
	h := hmac.New(sha1.New, []byte(secretKey))
	h.Write([]byte(strToHash))
	expectedHash := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	if secureHash != expectedHash {
		t.Errorf("Expected secureHash %q, got %q", expectedHash, secureHash)
	}
}
