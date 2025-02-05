package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	api2c2p "github.com/choonkeat/2c2p"
)

var (
	port           = flag.Int("port", 8080, "Port to run the server on")
	sandbox        = flag.Bool("sandbox", true, "Use sandbox environment")
	outputFile     = flag.String("output", "", "Output HTML file path (optional)")
	formAction     = flag.String("formAction", "/process-payment", "Form action URL")
	merchantID     = flag.String("merchantID", "", "2C2P Merchant ID")
	secretKey      = flag.String("secretKey", "", "2C2P Secret Key")
	serverURL      = flag.String("serverURL", "http://localhost:8080", "Your server URL prefix (e.g., https://your-domain.com)")
	c2cpURL        = flag.String("c2cpURL", "https://demo2.2c2p.com", "2C2P server URL")
	privateKeyFile = flag.String("privateKey", "", "Path to ED25519 private key file")
)

func main() {
	flag.Parse()

	html := api2c2p.GenerateSecureFieldsHTML(*formAction, *sandbox)

	// If output file is specified, write the HTML to file
	if *outputFile != "" {
		dir := filepath.Dir(*outputFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Error creating directory: %v", err)
		}
		if err := os.WriteFile(*outputFile, []byte(html), 0644); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
		fmt.Printf("HTML written to %s\n", *outputFile)
		return
	}

	// Otherwise, start a web server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("securefields").Parse(html)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/process-payment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Get encrypted card data from form
		encryptedCardInfo := r.PostFormValue("encryptedCardInfo")

		// Create payment request XML
		timestamp := fmt.Sprintf("%d", time.Now().Unix())
		invoiceNo := fmt.Sprintf("INV%s", timestamp)
		amount := "000000010010" // $100.10 formatted as 12 digits
		currencyCode := "702"
		apiVersion := "9.4"
		desc := "1 room for 2 nights"
		cardholderName := "John Doe"
		country := "SG"
		storeCard := "Y"

		// Construct signature string with all fields in the same order as PHP
		strToHash := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s",
			apiVersion,        // version
			timestamp,         // timestamp
			*merchantID,       // merchantID
			invoiceNo,         // uniqueTransactionCode
			desc,              // desc
			amount,            // amt
			currencyCode,      // currencyCode
			"",                // paymentChannel
			"",                // storeCardUniqueID
			"",                // panBank
			country,           // country
			cardholderName,    // cardholderName
			"",                // cardholderEmail
			"",                // payCategoryID
			"",                // userDefined1
			"",                // userDefined2
			"",                // userDefined3
			"",                // userDefined4
			"",                // userDefined5
			storeCard,         // storeCard
			"",                // ippTransaction
			"",                // installmentPeriod
			"",                // interestType
			"",                // recurring
			"",                // invoicePrefix
			"",                // recurringAmount
			"",                // allowAccumulate
			"",                // maxAccumulateAmt
			"",                // recurringInterval
			"",                // recurringCount
			"",                // chargeNextDate
			"",                // promotion
			"Y",               // request3DS
			"",                // statementDescriptor
			"",                // agentCode
			"",                // channelCode
			"",                // paymentExpiry
			"",                // mobileNo
			"",                // tokenizeWithoutAuthorization
			encryptedCardInfo, // encryptedCardInfo
		)
		h := hmac.New(sha1.New, []byte(*secretKey))
		h.Write([]byte(strToHash))
		hash := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

		// Create payment request XML
		xmlStr := fmt.Sprintf(`<PaymentRequest>
		<version>%s</version>
		<timeStamp>%s</timeStamp>
		<merchantID>%s</merchantID>
		<uniqueTransactionCode>%s</uniqueTransactionCode>
		<desc>%s</desc>
		<amt>%s</amt>
		<currencyCode>%s</currencyCode>
		<paymentChannel></paymentChannel>
		<panCountry>%s</panCountry>
		<cardholderName>%s</cardholderName>
		<request3DS>Y</request3DS>
		<secureHash>%s</secureHash>
		<storeCard>%s</storeCard>
		<encCardData>%s</encCardData>
	</PaymentRequest>`,
			apiVersion,
			timestamp,
			*merchantID,
			invoiceNo,
			desc,
			amount,
			currencyCode,
			country,
			cardholderName,
			hash,
			storeCard,
			encryptedCardInfo,
		)

		// Base64 encode the XML
		payload := base64.StdEncoding.EncodeToString([]byte(xmlStr))

		// Render auto-submitting form
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<body>
	<form action="%s/2C2PFrontEnd/SecurePayment/PaymentAuth.aspx" method="POST" name="paymentRequestForm">
		Processing payment request, Do not close the browser, press back or refresh the page.
		<input type="hidden" name="paymentRequest" value="%s">
	</form>
	<script>
		document.paymentRequestForm.submit();
	</script>
</body>
</html>`, *c2cpURL, payload)
	})

	http.HandleFunc("/payment-return", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Get payment response
		paymentResponse := r.FormValue("paymentResponse")
		fmt.Printf("Payment response: %s\n", paymentResponse)

		// Read private key file
		privateKey, err := os.ReadFile(*privateKeyFile)
		if err != nil {
			http.Error(w, "Error reading private key: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Decrypt using ED25519 private key
		decrypted, err := api2c2p.Decrypt([]byte(paymentResponse), privateKey)
		if err != nil {
			http.Error(w, "Error decrypting response: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("Decrypted response: %s\n", string(decrypted))

		// Parse the XML response
		var response api2c2p.PaymentResponseBackEnd
		err = xml.Unmarshal(decrypted, &response)
		if err != nil {
			http.Error(w, "Error parsing XML response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Handle frontend return - display payment result to user
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
	<style>
		.payment-details { margin: 20px; }
		.payment-details dt { font-weight: bold; margin-top: 10px; }
		.success { color: green; }
		.failure { color: red; }
	</style>
</head>
<body>
	<h1>Payment Result</h1>
	<div class="payment-details">
		<h2>Payment Status: <span class="%s">%s</span></h2>
		%s
		<dl>
			<dt>Transaction Code:</dt><dd>%s</dd>
			<dt>Amount:</dt><dd>%s</dd>
			<dt>Card:</dt><dd>%s (%s)</dd>
			<dt>Bank:</dt><dd>%s</dd>
			<dt>Response Code:</dt><dd>%s</dd>
			<dt>DateTime:</dt><dd>%s</dd>
		</dl>
	</div>
	<h3>Raw Response</h3>
	<pre>%s</pre>
</body>
</html>`,
			func() string {
				if response.Status == "S" && response.RespCode == "00" && response.FailReason == "" {
					return "success"
				}
				return "failure"
			}(),
			func() string {
				if response.Status == "S" && response.RespCode == "00" && response.FailReason == "" {
					return "Success"
				}
				return "Failed"
			}(),
			func() string {
				if response.FailReason != "" {
					return "<p class=\"failure\">Reason: " + response.FailReason + "</p>"
				}
				return ""
			}(),
			response.UniqueTransactionCode,
			response.Amount,
			response.CardType,
			response.PAN,
			response.BankName,
			response.RespCode,
			response.DateTime,
			string(decrypted),
		)
	})

	http.HandleFunc("/payment-notify", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
			return
		}
		// Handle backend notification - process payment result
		log.Printf("Payment notification received: %v", r.PostForm)
		w.WriteHeader(http.StatusOK)
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Starting server at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
