package api2c2p

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
