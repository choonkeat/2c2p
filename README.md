# 2C2P Payment Gateway API Client

A Go client library for integrating with the 2C2P Payment Gateway API (v4.3.1).

## Features

- JWT-based authentication
- Payment Inquiry API support
- Payment Token API support
- Comprehensive test coverage
- CLI tools for testing each API endpoint

## API Documentation

Always refer to the official 2C2P API documentation for the most up-to-date information:

- [API Documentation Portal](https://developer.2c2p.com/docs)
- [Payment Token API v4.3.1](https://developer.2c2p.com/v4.3.1/docs/api-payment-token)
  - Check the "Payment Token Request Parameter" section for all available request fields
  - Check the "Payment Token Response Parameter" section for response fields
- [Payment Inquiry API v4.3.1](https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry)
  - Check the "Payment Inquiry Request Parameter" section for all available request fields
  - Check the "Payment Inquiry Response Parameter" section for response fields

## Field Naming Conventions

When defining request/response types in Go, follow these naming conventions:

1. **Preserve Original JSON Names**: Always keep the original JSON field name in the struct tag
   ```go
   // Good - preserves JSON name and adds unit suffix to Go field
   RecurringIntervalDays int `json:"recurringInterval"`
   
   // Bad - changed JSON name
   RecurringDays int `json:"recurringDays"`
   
   // Also Bad - preserved JSON name but Go field name doesn't indicate unit
   RecurringDays int `json:"recurringInterval"`
   ```

2. **Add Format/Standard Suffixes**: For fields with specific formats or standards, append the format to the Go field name
   ```go
   // Currency codes
   CurrencyCodeISO4217 string `json:"currencyCode"`
   
   // Country codes
   CountryCodeISO3166 string `json:"countryCode"`
   
   // Date formats
   ChargeNextDateYYYYMMDD string `json:"chargeNextDate"`
   PaymentExpiryYYYYMMDDHHMMSS string `json:"paymentExpiry"`
   ```

3. **Add Unit Suffixes**: For fields representing quantities, append the unit to the Go field name
   ```go
   // Time durations
   RecurringIntervalDays int `json:"recurringInterval"`
   InstallmentPeriodFilterMonths []int `json:"installmentPeriodFilter"`
   ```

4. **Don't Add Type Suffixes**: Don't append Go type names to fields
   ```go
   // Good
   Amount float64 `json:"amount"`
   
   // Bad - includes type name
   AmountFloat64 float64 `json:"amount"`
   ```

These conventions help developers understand:
- The exact format required for date strings (YYYYMMDD vs YYYYMMDDHHMMSS)
- Which standards are being used (ISO4217 vs ISO3166)
- What units are expected (days vs months)
- The original API field names (via JSON tags)

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

#### Testing CLI Help

Always test the CLI tools first with the `-h` flag to see available options:

```bash
go run cli/payment-inquiry/main.go -h
go run cli/payment-token/main.go -h
```

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

#### Payment Token CLI

```bash
# Basic usage
go run cli/payment-token/main.go \
    -merchantID your_merchant_id \
    -secretKey your_secret_key \
    -currencyCode THB \
    -amount 100.00 \
    -invoiceNo INV123 \
    -description "Test payment"

# Advanced usage with optional parameters
go run cli/payment-token/main.go \
    -merchantID your_merchant_id \
    -secretKey your_secret_key \
    -currencyCode THB \
    -amount 100.00 \
    -invoiceNo INV123 \
    -description "Test payment" \
    -paymentChannel "CC,IPP,APM" \
    -request3DS "Y" \
    -tokenize \
    -cardTokens "token1,token2" \
    -installmentPeriods "3,6,9" \
    -recurring \
    -recurringAmount 100.00 \
    -recurringInterval 30 \
    -recurringCount 12 \
    -paymentExpiry "2024-12-31 23:59:59" \
    -frontendURL "https://your-site.com/return" \
    -backendURL "https://your-site.com/notify" \
    -userName "John Doe" \
    -userEmail "john@example.com" \
    -userMobile "1234567890" \
    -userCountry "SG" \
    -userMobilePrefix "65" \
    -locale "en"
```

## Project Structure

```
.
├── client.go                 # Core client implementation with JWT handling
├── payment_inquiry.go        # Payment Inquiry API implementation
├── payment_inquiry_test.go   # Tests for Payment Inquiry
├── cli/                      # CLI tools for testing each API
│   └── payment-inquiry/      # Payment Inquiry CLI tool
│       └── main.go
│   └── payment-token/        # Payment Token CLI tool
│       └── main.go
├── logs/                     # Development conversation logs
│   └── YYYY-MM-DD.md         # Daily conversation logs
└── Makefile                  # Build and test automation
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
Please append our raw conversation to logs/YYYY-MM-DD.md
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
