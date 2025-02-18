package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	api2c2p "github.com/choonkeat/2c2p"
)

var (
	// Server configuration
	port = flag.Int("port", 8080, "Port to run the server on")

	// 2C2P configuration
	sandbox                = flag.Bool("sandbox", true, "Use sandbox environment")
	merchantID             = flag.String("merchantID", "", "2C2P Merchant ID")
	secretKey              = flag.String("secretKey", "", "2C2P Secret Key")
	combinedPem            = flag.String("combinedPem", "dist/combined_private_public.pem", "Path to combined private key and certificate PEM file generated by cmd/server-to-server-key/main.go")
	serverJWTPublicKeyFile = flag.String("serverJWTPublicKey", "dist/sandbox-jwt-2c2p.demo.2.1(public).cer", "Path to 2C2P's public key certificate (.cer file)")
	serverPKCS7PublicKey   = flag.String("serverPKCS7PublicKey", "dist/sandbox-pkcs7-demo2.2c2p.com(public).cer", "Path to 2C2P's public key certificate (.cer file)")
	paymentGatewayURL      = flag.String("paymentGatewayURL", "https://sandbox-pgw.2c2p.com", "2C2P Payment Gateway URL")
	frontendURL            = flag.String("frontendURL", "https://demo2.2c2p.com", "2C2P Frontend URL")

	// Form configuration
	formAction       = flag.String("formAction", "/process-payment", "Form action URL")
	isLoyaltyPayment = flag.Bool("isLoyaltyPayment", false, "Is loyalty payment")
)

// main starts a web server that demonstrates the 2C2P payment flow:
// 1. Displays a payment form with secure card fields
// 2. Processes the payment request and redirects to 2C2P
// 3. Handles the payment response and displays the result
// 4. Receives backend notifications for payment status updates
func main() {
	flag.Parse()

	// Create 2C2P client
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
		log.Fatal(err)
	}

	// Generate the payment form HTML with secure fields
	secureFieldsHTML := api2c2p.SecureFieldsFormHTML(*merchantID, *secretKey, *formAction, *sandbox)

	// Handler for the payment form page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.String())
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, secureFieldsHTML)
	})

	// Handler for processing payment form submission
	http.HandleFunc("/process-payment", handlePaymentRequest)

	// Handler for payment response from 2C2P
	// Create a closure to pass the pre-loaded private key
	http.HandleFunc("/payment-return", func(w http.ResponseWriter, r *http.Request) {
		handlePaymentResponse(w, r, client)
	})

	// Handler for backend payment notifications
	http.HandleFunc("/payment-notify", func(w http.ResponseWriter, r *http.Request) {
		handlePaymentNotification(w, r, client)
	})

	// Start the server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

// handlePaymentRequest processes the payment form submission:
// 1. Validates the request
// 2. Creates a payment request XML
// 3. Signs the request with HMAC
// 4. Redirects to 2C2P payment page
func handlePaymentRequest(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.String())
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Prepare payment request parameters
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	invoiceNo := fmt.Sprintf("INV%s", timestamp)
	paymentDetails := api2c2p.SecureFieldsPaymentDetails{
		AmountCents:      1234,
		CurrencyCode:     "702", // SGD
		IsLoyaltyPayment: *isLoyaltyPayment,
		Description:      "1 room for 2 nights",
		CustomerName:     "John Doe",
		CountryCode:      "SG",
		StoreCard:        "Y",
		UserDefined1:     "1",
		UserDefined2:     "2",
		UserDefined3:     "3",
		UserDefined4:     "4",
		UserDefined5:     "5",
	}

	// Create HMAC signature string
	payload := api2c2p.CreateSecureFieldsPaymentPayload(*frontendURL, *merchantID, *secretKey, timestamp, invoiceNo, paymentDetails, r)
	log.Printf("Payment request FormURL: %s", payload.FormURL)
	log.Printf("Payment request FormFields: %#v", payload.FormFields)

	// Render auto-submitting form to 2C2P
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.New("result").Parse(`<!DOCTYPE html>
<html>
<body>
	<form action="{{.FormURL}}" method="POST" name="paymentRequestForm">
		<p>Processing payment request. Please do not close the browser, press back or refresh the page.</p>
		{{range $key, $value := .FormFields}}
			<input type="hidden" name="{{$key}}" value="{{$value}}">
		{{end}}
	</form>
	<script>document.paymentRequestForm.submit();</script>
</body>
</html>`))

	err := tmpl.Execute(w, payload)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
	}
}

// handlePaymentResponse processes the payment response from 2C2P:
// 1. Decrypts the response using our private key
// 2. Parses the payment result XML
// 3. Displays the payment result to the customer
func handlePaymentResponse(w http.ResponseWriter, r *http.Request, client *api2c2p.Client) {
	log.Println(r.Method, r.URL.String())
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Decrypt and parse the payment response
	response, decrypted, err := client.DecryptPaymentResponseBackend(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decrypting payment notification: %v", err), http.StatusBadRequest)
		return
	}

	// Display payment result to customer
	renderPaymentResult(w, response, string(decrypted))
}

// handlePaymentNotification processes backend notifications from 2C2P
// These notifications are used to update the payment status in your system
func handlePaymentNotification(w http.ResponseWriter, r *http.Request, client *api2c2p.Client) {
	log.Println(r.Method, r.URL.String())
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Decrypt and parse the payment response
	response, decrypted, err := client.DecryptPaymentResponseBackend(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decrypting payment notification: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Payment notification received: RespCode=%s XML=%s", string(response.RespCode), string(decrypted))

	inquiryResponse, err := client.PaymentInquiryByInvoice(r.Context(), &api2c2p.PaymentInquiryByInvoiceRequest{
		InvoiceNo: response.UniqueTransactionCode,
		Locale:    "en",
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inquiring payment: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Payment inquiry result: %#v", inquiryResponse)

	w.WriteHeader(http.StatusOK)
}

// Helper functions

func renderPaymentResult(w http.ResponseWriter, response api2c2p.PaymentResponseBackEnd, rawResponse string) {
	w.Header().Set("Content-Type", "text/html")

	type templateData struct {
		StatusClass string
		StatusText  string
		FailReason  template.HTML
		TransCode   string
		Amount      string
		CardNumber  string
		CardType    string
		Bank        string
		RespCode    string
		DateTime    string
		RawResponse string
	}

	isSuccess := response.Status == "A" && response.RespCode == "00"

	statusClass := "failure"
	if isSuccess {
		statusClass = "success"
	}

	statusText := "Failed"
	if isSuccess {
		statusText = "Success"
	}

	var failReasonHTML template.HTML
	if response.FailReason != "" {
		failReasonHTML = template.HTML(fmt.Sprintf(`<p class="failure">Reason: %s</p>`, template.HTMLEscapeString(response.FailReason)))
	}

	tmpl := template.Must(template.New("result").Parse(`<!DOCTYPE html>
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
		<h2>Payment Status: <span class="{{.StatusClass}}">{{.StatusText}}</span></h2>
		{{.FailReason}}
		<dl>
			<dt>Transaction Code:</dt><dd>{{.TransCode}}</dd>
			<dt>Amount:</dt><dd>{{.Amount}}</dd>
			<dt>Card:</dt><dd>{{.CardNumber}} ({{.CardType}})</dd>
			<dt>Bank:</dt><dd>{{.Bank}}</dd>
			<dt>Response Code:</dt><dd>{{.RespCode}}</dd>
			<dt>DateTime:</dt><dd>{{.DateTime}}</dd>
		</dl>
	</div>
	<h3>Raw Response</h3>
	<pre>{{.RawResponse}}</pre>
</body>
</html>`))

	data := templateData{
		StatusClass: statusClass,
		StatusText:  statusText,
		FailReason:  failReasonHTML,
		TransCode:   response.UniqueTransactionCode,
		Amount:      response.Amount,
		CardNumber:  response.PAN,
		CardType:    response.CardType,
		Bank:        response.BankName,
		RespCode:    string(response.RespCode),
		DateTime:    response.DateTime,
		RawResponse: rawResponse,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
