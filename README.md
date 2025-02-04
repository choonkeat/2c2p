# 2C2P Payment Gateway API Client

A Go client library for integrating with the 2C2P Payment Gateway API (v4.3.1).

## Features

- JWT-based authentication
- Payment Inquiry API support
- Comprehensive test coverage
- CLI tools for testing each API endpoint

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

### Using the CLI Tools

Each API endpoint has its own CLI tool for testing. The tools are located in the `cli` directory:

#### Payment Inquiry CLI

```bash
go run cli/payment-inquiry/main.go \
    -merchantID your_merchant_id \
    -secretKey your_secret_key \
    -invoiceNo your_invoice_number \
    [-paymentToken payment_token] \
    [-locale en] \
    [-baseURL https://sandbox-pgw.2c2p.com]
```

## Project Structure

```
.
├── client.go                 # Core client implementation with JWT handling
├── payment_inquiry.go        # Payment Inquiry API implementation
├── payment_inquiry_test.go   # Tests for Payment Inquiry
├── cli/                     # CLI tools for testing each API
│   └── payment-inquiry/     # Payment Inquiry CLI tool
│       └── main.go
├── logs/                    # Development conversation logs
│   └── YYYY-MM-DD.md       # Daily conversation logs
└── Makefile                 # Build and test automation
```

## Development

### Running Tests

```bash
make test
```

### Viewing Documentation

```bash
make docs-view
```

Then visit http://localhost:6060/pkg/github.com/choonkeat/2c2p

### Logging Development Conversations

Development conversations with AI assistants are automatically logged in the `logs` directory. To append the current conversation to today's log:

```
Please append our conversation to logs/YYYY-MM-DD.md
```

This helps maintain a record of design decisions and implementation details.

## Contributing New API Implementations

When implementing a new 2C2P API endpoint:

1. **File Organization**
   - Create a new file named after the API (e.g., `payment_inquiry.go` for Payment Inquiry API)
   - Create a corresponding test file (e.g., `payment_inquiry_test.go`)
   - Create a CLI tool in `cli/api-name/main.go`

2. **Code Structure**
   - Add package documentation at the top of the file with links to relevant 2C2P API documentation
   - Define request and response structs with proper JSON tags and field comments
   - Implement the API method on the `Client` struct

3. **Documentation**
   - Include links to the API documentation in the file header
   - Document each struct field with its type and whether it's required
   - Add usage examples in the package documentation

4. **Testing**
   - Use sample request/response values from the 2C2P documentation
   - Test both success and error cases
   - Include JWT token validation tests

5. **CLI Tool**
   - Create a new directory under `cli` for your API (e.g., `cli/payment-inquiry/`)
   - Implement a focused CLI tool with flags specific to your API
   - Include proper validation and help text for all flags
   - Format the output to be human-readable

For example, see how the Payment Inquiry API is implemented:
1. API implementation in `payment_inquiry.go`
2. Tests in `payment_inquiry_test.go`
3. CLI tool in `cli/payment-inquiry/main.go`

## Documentation References

- [2C2P API Documentation](https://developer.2c2p.com/v4.3.1/docs)
- Specific endpoints:
  - [Payment Inquiry](https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry)
  - [Payment Inquiry Request Parameters](https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry-request-parameter)
  - [Payment Inquiry Response Parameters](https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry-response-parameter)
  - [JWT Token Guide](https://developer.2c2p.com/v4.3.1/docs/json-web-tokens-jwt)

## Response Codes

For a complete list of response codes and their meanings, refer to the [Response Code List](https://developer.2c2p.com/v4.3.1/docs/response-code-payment) in the 2C2P documentation.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
