# 2C2P Payment Gateway API Client

A Go client library for integrating with the 2C2P Payment Gateway API (v4.3.1).

## Features

- JWT-based authentication
- Payment Inquiry API support
- Comprehensive test coverage
- CLI tool for easy testing

## Installation

```bash
go get github.com/choonkeat/2c2p
```

## Usage

### Creating a Client

```go
client := api2c2p.NewClient(
    "your_merchant_id",
    "your_secret_key",
    "https://sandbox-pgw.2c2p.com", // or https://pgw.2c2p.com for production
)
```

### Making a Payment Inquiry

```go
request := &api2c2p.PaymentInquiryRequest{
    MerchantID:   "your_merchant_id",
    InvoiceNo:    "your_invoice_number",  // Either InvoiceNo or PaymentToken is required
    PaymentToken: "payment_token",        // Optional, alternative to InvoiceNo
    Locale:       "en",                   // Optional
}

response, err := client.PaymentInquiry(request)
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Payment status: %s - %s\n", response.RespCode, response.RespDesc)
```

### Using the CLI Tool

The package includes a CLI tool for testing the API:

```bash
go run example/main.go \
    -function PaymentInquiry \
    -merchantID your_merchant_id \
    -invoiceNo your_invoice_number \
    -secretKey your_secret_key \
    [-locale en] \
    [-baseURL https://sandbox-pgw.2c2p.com]
```

## Project Structure

```
.
├── client.go              # Core client implementation with JWT handling
├── payment_inquiry.go     # Payment Inquiry API implementation
├── payment_inquiry_test.go # Tests for Payment Inquiry
├── example/
│   └── main.go           # CLI tool implementation
└── Makefile              # Build and test automation
```

## Testing

Run the test suite:

```bash
make test
```

## Documentation

- [2C2P API Documentation](https://developer.2c2p.com/v4.3.1/docs)
- Specific endpoints:
  - [Payment Inquiry](https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry)
  - [Payment Inquiry Request Parameters](https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry-request-parameter)
  - [Payment Inquiry Response Parameters](https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry-response-parameter)
  - [JWT Token Guide](https://developer.2c2p.com/v4.3.1/docs/json-web-tokens-jwt)

## Response Codes

For a complete list of response codes and their meanings, refer to the [Response Code List](https://developer.2c2p.com/v4.3.1/docs/response-code-payment) in the 2C2P documentation.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin feature/my-new-feature`)
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
