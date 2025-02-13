package api2c2p

// `Server-to-server API - Frontend return URL` must be set in the 2c2p portal
import (
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"log"
	"strings"

	"github.com/fullsailor/pkcs7"
	"github.com/google/uuid"
)

// SecureFieldsResponse represents a response from the secure fields API
// Documentation (PHP Code): https://developer.2c2p.com/v4.3.1/docs/using-securefields
type SecureFieldsResponse struct {
	// EncryptedCardInfo contains the encrypted card data from 2C2P Secure Fields
	EncryptedCardInfo string // used to make card payment

	// MaskedCardInfo contains first 6 and last 4 masked PAN
	MaskedCardInfo string // first 6 and last 4 masked PAN

	// ExpMonthCardInfo contains the card expiry month
	ExpMonthCardInfo string // card expiry month

	// ExpYearCardInfo contains the card expiry year
	ExpYearCardInfo string // card expiry year

	// RespCode contains the response code
	RespCode PaymentResponseCode `json:"respCode"` // Response code
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

// FormValuer is an interface for getting form values
type FormValuer interface {
	PostFormValue(string) string
}

// SecureFieldsScriptURLs returns the URLs for required JavaScript files
// Set sandbox to true for testing environment
func SecureFieldsScriptURLs(sandbox bool) (secureFieldsJS, securePay string) {
	if sandbox {
		return "https://2c2p-uat-cloudfront.s3-ap-southeast-1.amazonaws.com/2C2PPGW/secureField/my2c2p-secureFields.1.0.0.min.js",
			"https://demo2.2c2p.com/2C2PFrontEnd/SecurePayment/api/my2c2p-sandbox.1.7.3.min.js"
	}
	// Production URLs - TODO: confirm with 2C2P for production URLs
	return "https://2c2p-cloudfront.s3-ap-southeast-1.amazonaws.com/2C2PPGW/secureField/my2c2p-secureFields.1.0.0.min.js",
		"https://2c2p.com/2C2PFrontEnd/SecurePayment/api/my2c2p.1.7.3.min.js"
}

// SecureFieldsFormHTML generates the HTML template for secure fields form
func SecureFieldsFormHTML(merchantID, secretKey, formAction string, sandbox bool) string {
	secureFieldsJS, securePayJS := SecureFieldsScriptURLs(sandbox)
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

func createSignatureString(apiVersion, timestamp, merchantID, invoiceNo string, details SecureFieldsPaymentDetails, encryptedCardInfo string) string {
	// Construct signature string with all fields in the same order as PHP
	return fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s",
		apiVersion,          // version
		timestamp,           // timestamp
		merchantID,          // merchantID
		invoiceNo,           // uniqueTransactionCode
		details.Description, // desc
		details.AmountCents.ZeroPrefixed12DCents(), // amt
		details.CurrencyCode,                       // currencyCode
		"",                                         // paymentChannel
		"",                                         // storeCardUniqueID
		"",                                         // panBank
		details.CountryCode,                        // country
		details.CustomerName,                       // cardholderName
		"",                                         // cardholderEmail
		"",                                         // payCategoryID
		details.UserDefined1,                       // userDefined1
		details.UserDefined2,                       // userDefined2
		details.UserDefined3,                       // userDefined3
		details.UserDefined4,                       // userDefined4
		details.UserDefined5,                       // userDefined5
		details.StoreCard,                          // storeCard
		"",                                         // ippTransaction
		"",                                         // installmentPeriod
		"",                                         // interestType
		"",                                         // recurring
		"",                                         // invoicePrefix
		"",                                         // recurringAmount
		"",                                         // allowAccumulate
		"",                                         // maxAccumulateAmt
		"",                                         // recurringInterval
		"",                                         // recurringCount
		"",                                         // chargeNextDate
		"",                                         // promotion
		"Y",                                        // request3DS
		"",                                         // statementDescriptor
		"",                                         // agentCode
		"",                                         // channelCode
		"",                                         // paymentExpiry
		"",                                         // mobileNo
		"",                                         // tokenizeWithoutAuthorization
		encryptedCardInfo,                          // encryptedCardInfo
	)
}

func createHMAC(data, key string) string {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(data))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

// SecureFieldsPaymentPayload contains all required fields to render a 2C2P payment form
type SecureFieldsPaymentPayload struct {
	FormURL    string
	FormFields map[string]string
}

type SecureFieldsPaymentDetails struct {
	AmountCents      Cents
	CurrencyCode     string
	IsLoyaltyPayment bool
	Description      string
	CustomerName     string
	CountryCode      string
	StoreCard        string
	UserDefined1     string
	UserDefined2     string
	UserDefined3     string
	UserDefined4     string
	UserDefined5     string
}

// PaymentRequest represents the XML structure for a payment request
type PaymentRequest struct {
	XMLName               xml.Name         `xml:"PaymentRequest"`
	Version               string           `xml:"version"`
	TimeStamp             string           `xml:"timeStamp"`
	MerchantID            string           `xml:"merchantID"`
	UniqueTransactionCode string           `xml:"uniqueTransactionCode"`
	Description           string           `xml:"desc"`
	Amount                string           `xml:"amt"`
	CurrencyCode          string           `xml:"currencyCode"`
	PaymentChannel        string           `xml:"paymentChannel"`
	PanCountry            string           `xml:"panCountry"`
	CardholderName        string           `xml:"cardholderName"`
	Request3DS            string           `xml:"request3DS"`
	SecureHash            string           `xml:"secureHash"`
	StoreCard             string           `xml:"storeCard"`
	EncCardData           string           `xml:"encCardData"`
	UserDefined1          string           `xml:"userDefined1"`
	UserDefined2          string           `xml:"userDefined2"`
	UserDefined3          string           `xml:"userDefined3"`
	UserDefined4          string           `xml:"userDefined4"`
	UserDefined5          string           `xml:"userDefined5"`
	IsLoyaltyPayment      YesNo            `xml:"isLoyaltyPayment,omitempty"` // Y or N
	LoyaltyPayments       *LoyaltyPayments `xml:"loyaltyPayments,omitempty"`
}

type LoyaltyPayments struct {
	LoyaltyPayment []LoyaltyPayment `xml:"loyaltyPayment"`
}

type LoyaltyPayment struct {
	LoyaltyProvider string     `xml:"loyaltyProvider,omitempty"`
	RedeemAmt       Dollars    `xml:"redeemAmt"`
	RedeemCurrency  string     `xml:"redeemCurrency"`
	Redemption      Redemption `xml:"redemption"`
}

type Redemption struct {
	Reward Reward `xml:"reward"`
}

type YesNo bool

const (
	Yes YesNo = true
	No  YesNo = false
)

// MarshalXML implements xml.Marshaler
func (yn YesNo) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if yn {
		return e.EncodeElement("Y", start)
	}
	return e.EncodeElement("N", start)
}

// UnmarshalXML implements xml.Unmarshaler
func (yn *YesNo) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := dec.DecodeElement(&s, &start); err != nil {
		return err
	}
	switch s {
	case "Y":
		*yn = true
	case "N":
		*yn = false
	default:
		return fmt.Errorf("invalid value: %s", s)
	}
	return nil
}

func CreateSecureFieldsPaymentPayload(c2pURL, merchantID, secretKey, timestamp, invoiceNo string, paymentDetails SecureFieldsPaymentDetails, form FormValuer) SecureFieldsPaymentPayload {
	encryptedCardInfo := form.PostFormValue("encryptedCardInfo")

	// Create HMAC signature string
	strToHash := createSignatureString(
		"9.4", // API version
		timestamp,
		merchantID,
		invoiceNo,
		paymentDetails,
		encryptedCardInfo,
	)

	// Create HMAC hash
	hmacHash := createHMAC(strToHash, secretKey)

	// Create payment request XML
	paymentRequest := PaymentRequest{
		Version:               "9.4",
		TimeStamp:             timestamp,
		MerchantID:            merchantID,
		UniqueTransactionCode: invoiceNo,
		Description:           paymentDetails.Description,
		Amount:                paymentDetails.AmountCents.ZeroPrefixed12DCents(),
		CurrencyCode:          paymentDetails.CurrencyCode,
		PanCountry:            paymentDetails.CountryCode,
		CardholderName:        paymentDetails.CustomerName,
		Request3DS:            "Y",
		SecureHash:            hmacHash,
		StoreCard:             paymentDetails.StoreCard,
		EncCardData:           encryptedCardInfo,
		UserDefined1:          paymentDetails.UserDefined1,
		UserDefined2:          paymentDetails.UserDefined2,
		UserDefined3:          paymentDetails.UserDefined3,
		UserDefined4:          paymentDetails.UserDefined4,
		UserDefined5:          paymentDetails.UserDefined5,
	}

	if paymentDetails.IsLoyaltyPayment {
		paymentRequest.IsLoyaltyPayment = Yes
		paymentRequest.LoyaltyPayments = &LoyaltyPayments{
			LoyaltyPayment: []LoyaltyPayment{
				{
					LoyaltyProvider: "MCCY",
					RedeemAmt:       paymentDetails.AmountCents.ToDollars(),
					RedeemCurrency:  "SGD", // paymentDetails.CurrencyCode,
					Redemption: Redemption{
						Reward: Reward{
							ID:       uuid.New().String(), // generate random UUID
							Quantity: paymentDetails.AmountCents.ToDollars(),
						},
					},
				},
			},
		}
	}
	log.Printf("Payment request: %v", paymentRequest)

	// Marshal the payment request to XML
	xmlBytes, err := xml.Marshal(paymentRequest)
	if err != nil {
		log.Printf("Error marshaling payment request to XML: %v", err)
		return SecureFieldsPaymentPayload{}
	}
	log.Printf("Payment request XML: %s", string(xmlBytes))

	// Base64 encode the XML
	return SecureFieldsPaymentPayload{
		FormURL: c2pURL + "/2C2PFrontEnd/SecurePayment/PaymentAuth.aspx",
		FormFields: map[string]string{
			"paymentRequest": base64.StdEncoding.EncodeToString(xmlBytes),
		},
	}
}

type PaymentResponseBackEnd struct {
	XMLName               xml.Name            `xml:"PaymentResponse"`
	Version               string              `xml:"version"`
	TimeStamp             string              `xml:"timeStamp"`
	MerchantID            string              `xml:"merchantID"`
	RespCode              PaymentResponseCode `xml:"respCode"`
	PAN                   string              `xml:"pan"`
	Amount                string              `xml:"amt"`
	UniqueTransactionCode string              `xml:"uniqueTransactionCode"`
	TranRef               string              `xml:"tranRef"`
	ApprovalCode          string              `xml:"approvalCode"`
	RefNumber             string              `xml:"refNumber"`
	ECI                   string              `xml:"eci"`
	DateTime              string              `xml:"dateTime"`
	Status                string              `xml:"status"`
	FailReason            string              `xml:"failReason"` // can contain successful reason too
	UserDefined1          string              `xml:"userDefined1"`
	UserDefined2          string              `xml:"userDefined2"`
	UserDefined3          string              `xml:"userDefined3"`
	UserDefined4          string              `xml:"userDefined4"`
	UserDefined5          string              `xml:"userDefined5"`
	IPPPeriod             string              `xml:"ippPeriod"`
	IPPInterestType       string              `xml:"ippInterestType"`
	IPPInterestRate       string              `xml:"ippInterestRate"`
	IPPMerchantAbsorbRate string              `xml:"ippMerchantAbsorbRate"`
	PaidChannel           string              `xml:"paidChannel"`
	PaidAgent             string              `xml:"paidAgent"`
	PaymentChannel        string              `xml:"paymentChannel"`
	BackendInvoice        string              `xml:"backendInvoice"`
	IssuerCountry         string              `xml:"issuerCountry"`
	IssuerCountryA3       string              `xml:"issuerCountryA3"`
	BankName              string              `xml:"bankName"`
	CardType              string              `xml:"cardType"`
	ProcessBy             string              `xml:"processBy"`
	PaymentScheme         string              `xml:"paymentScheme"`
	PaymentID             string              `xml:"paymentID"`
	AcquirerResponseCode  string              `xml:"acquirerResponseCode"`
	SchemePaymentID       string              `xml:"schemePaymentID"`
	HashValue             string              `xml:"hashValue"`
}

// DecryptPaymentResponseBackend decrypts and parses the payment response from 2C2P
func (c *Client) DecryptPaymentResponseBackend(r FormValuer) (PaymentResponseBackEnd, []byte, error) {
	encryptedResponse := r.PostFormValue("paymentResponse")

	// Decrypt the response
	decrypted, err := decryptPKCS7([]byte(encryptedResponse), c.PrivateKey, c.PublicCert)
	if err != nil {
		return PaymentResponseBackEnd{}, nil, fmt.Errorf("error decrypting response: %w", err)
	}

	// Parse XML response
	var response PaymentResponseBackEnd
	err = xml.Unmarshal(decrypted, &response)
	if err != nil {
		return PaymentResponseBackEnd{}, nil, fmt.Errorf("error parsing XML response: %w", err)
	}

	return response, decrypted, nil
}

// decryptPKCS7 decrypts base64-encoded PKCS7 enveloped data using certificate and private key from PEM data.
// The combinedPEM must contain both a private key (PKCS8) and certificate in PEM format.
func decryptPKCS7(encryptedData []byte, privateKey *rsa.PrivateKey, publicCert *x509.Certificate) ([]byte, error) {
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
	decrypted, err := p7.Decrypt(publicCert, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %v", err)
	}

	return decrypted, nil
}
