package api2c2p

// `Server-to-server API - Frontend return URL` must be set in the 2c2p portal
import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/fullsailor/pkcs7"
)

// SecureFieldsResponse represents the encrypted card data response from 2C2P Secure Fields
type SecureFieldsResponse struct {
	EncryptedCardInfo string // used to make card payment
	MaskedCardInfo    string // first 6 and last 4 masked PAN
	ExpMonthCardInfo  string // card expiry month
	ExpYearCardInfo   string // card expiry year
}

// SecureFieldsErrorResponse represents error details from 2C2P Secure Fields
type SecureFieldsErrorResponse struct {
	ErrorCode        int    `json:"errCode"`
	ErrorDescription string `json:"errDesc"`
}

// SecureFieldsPaymentResponse represents the decoded response from 2C2P Secure Fields payment
type SecureFieldsPaymentResponse struct {
	InvoiceNo   string `json:"invoiceNo"`   // Invoice number, unique merchant order number
	ChannelCode string `json:"channelCode"` // Payment channel code
	RespCode    string `json:"respCode"`    // Response code
	RespDesc    string `json:"respDesc"`    // Response description
}

// GetSecureFieldsJSURLs returns the URLs for required JavaScript files
// Set sandbox to true for testing environment
func GetSecureFieldsJSURLs(sandbox bool) (secureFieldsJS, securePay string) {
	if sandbox {
		return "https://2c2p-uat-cloudfront.s3-ap-southeast-1.amazonaws.com/2C2PPGW/secureField/my2c2p-secureFields.1.0.0.min.js",
			"https://demo2.2c2p.com/2C2PFrontEnd/SecurePayment/api/my2c2p-sandbox.1.7.3.min.js"
	}
	// Production URLs - TODO: confirm with 2C2P for production URLs
	return "https://2c2p-cloudfront.s3-ap-southeast-1.amazonaws.com/2C2PPGW/secureField/my2c2p-secureFields.1.0.0.min.js",
		"https://2c2p.com/2C2PFrontEnd/SecurePayment/api/my2c2p.1.7.3.min.js"
}

// GenerateSecureFieldsHTML generates the HTML template for secure fields form
func GenerateSecureFieldsHTML(formAction string, sandbox bool) string {
	secureFieldsJS, securePayJS := GetSecureFieldsJSURLs(sandbox)
	return `<!DOCTYPE html>
<html>
<head>
    <title>2C2P SecureField</title>
    <script type="text/javascript" src="` + secureFieldsJS + `"></script>
    <script type="text/javascript" src="` + securePayJS + `"></script>
    <style>
        ._2c2pPaymentField { margin: 5px; }
        ._2c2pCard { color: blue; }
        ._2c2pMonth { color: brown; }
        ._2c2pYear { color: red; }
        ._2c2pCvv { color: green; }
        ._2c2pPaymentFieldError { color: red; font-style: italic; }
    </style>
</head>
<body>
    <form id="2c2p-payment-form" action="` + formAction + `" method="POST"></form>
    <input type="button" value="Checkout" onclick="Checkout()" />

    <script type="text/javascript">
        function Checkout() {
            ClearFormErrorMessage();
            My2c2p.getEncrypted("2c2p-payment-form", function(encryptedData, errCode, errDesc) {
                DisplayFormErrorMessage(errCode, errDesc);
                if (errCode != 0) {
                    DisplayFormErrorMessage(errCode, errDesc);
                } else {
                    var form = document.getElementById("2c2p-payment-form");
                    if (form != undefined) {
                        // Send encryptedData to your backend:
                        // encryptedData.encryptedCardInfo - used to make card payment
                        // encryptedData.maskedCardInfo   - first 6 and last 4 masked PAN
                        // encryptedData.expMonthCardInfo - card expiry month
                        // encryptedData.expYearCardInfo  - card expiry year
                        form.submit();
                    }
                }
            });
        }

        function DisplayFormErrorMessage(errCode, errDesc) {
            var errControl;
            switch (errCode) {
                case 1:
                case 2:
                    errControl = document.getElementById('2c2pError-cardnumber');
                    break;
                case 3:
                case 8:
                case 9:
                    errControl = document.getElementById('2c2pError-month');
                    break;
                case 4:
                case 5:
                case 6:
                case 7:
                    errControl = document.getElementById('2c2pError-year');
                    break;
            }

            if (errControl != undefined) {
                errControl.innerHTML = errDesc;
                errControl.focus();
            } else {
                console.log(errDesc + '(' + errCode + ')');
            }
        }

        function ClearFormErrorMessage() {
            var errSpans = Array.prototype.slice.call(document.getElementsByClassName('_2c2pPaymentFieldError'));
            if (errSpans.length > 0) {
                errSpans.forEach(function(errSpan) { errSpan.innerHTML = ""; });
            }
        }
    </script>
</body>
</html>`
}

// DecodePaymentResponse decodes the payment response from 2C2P
func DecodePaymentResponse(paymentResponse string) (*SecureFieldsPaymentResponse, error) {
	// Decode base64URL string
	decoded, err := base64.RawURLEncoding.DecodeString(paymentResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64URL: %w", err)
	}

	// Parse JSON
	var response SecureFieldsPaymentResponse
	if err := json.Unmarshal(decoded, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return &response, nil
}

// Decrypt decrypts PKCS7 enveloped data using an RSA private key
func Decrypt(encryptedData []byte, combinedPEM []byte) ([]byte, error) {
	var privKey interface{}
	var cert *x509.Certificate

	// Read all PEM blocks
	for block, rest := pem.Decode(combinedPEM); block != nil; block, rest = pem.Decode(rest) {
		switch block.Type {
		case "PRIVATE KEY":
			var err error
			privKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key: %v", err)
			}
		case "CERTIFICATE":
			var err error
			cert, err = x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %v", err)
			}
		}
	}

	if privKey == nil {
		return nil, fmt.Errorf("no private key found in PEM data")
	}
	if cert == nil {
		return nil, fmt.Errorf("no certificate found in PEM data")
	}

	// Decode base64 data
	decodedData, err := base64.StdEncoding.DecodeString(string(encryptedData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %v", err)
	}

	// Parse the PKCS7 data
	p7, err := pkcs7.Parse(decodedData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS7 data: %v", err)
	}

	// Decrypt the data
	decrypted, err := p7.Decrypt(cert, privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %v", err)
	}

	return decrypted, nil
}
