# 2C2P Payment Gateway API Client

A Go client library for integrating with the 2C2P Payment Gateway API (v4.3.1).

## Features

- JWT-based authentication
- Payment Inquiry API support
- Payment Token API support
- SecureFields integration for PCI-compliant card data collection
- CLI tools for API testing and utilities
- Comprehensive test coverage

## Installation

```bash
go get github.com/choonkeat/2c2p
```

## API Documentation

Always refer to the official 2C2P API documentation:
- [API Documentation Portal](https://developer.2c2p.com/docs)
- [Payment Token API](https://developer.2c2p.com/v4.3.1/docs/api-payment-token)
- [Payment Inquiry API](https://developer.2c2p.com/v4.3.1/docs/api-payment-inquiry)
- [Using SecureFields](https://developer.2c2p.com/v4.3.1/docs/using-securefields)
- [Payment Response Parameters](https://developer.2c2p.com/v4.3.1/docs/api-payment-response-back-end-parameter)

## Usage

### Creating a Client

```go
client := api2c2p.NewClient(
    "your_merchant_id",
    "your_secret_key",
    "https://sandbox-pgw.2c2p.com", // or https://pgw.2c2p.com for production
)
```

### SecureFields Integration

Before running SecureFields integration:

1. Generate server-to-server key pair:
```bash
go run cli/server-to-server-key/main.go
```

2. Configure 2C2P merchant portal:
   - Go to Options > Server-to-server API
   - Upload the generated `dist/public_cert.pem` as your Public key
   - Set Frontend return URL to `$serverURL/payment-return`
   - Set Backend return URL to `$serverURL/payment-notify`

3. Start the SecureFields server:
```bash
go run cli/secure_fields/main.go -merchantID your_merchant_id -secretKey your_secret_key
```

For implementation details, refer to:
- Frontend response handling: See `handlePaymentResponse` in `cli/secure_fields/main.go`
- Backend notification handling: See `handlePaymentNotification` in `cli/secure_fields/main.go`
- Response field definitions: See `PaymentResponseBackEnd` in `payment_response_backend.go`

## Field Naming Conventions

When defining request/response types in Go, follow these conventions:

1. **Preserve Original JSON Names**:
```go
// Good - preserves JSON name and adds unit suffix to Go field
RecurringIntervalDays int `json:"recurringInterval"`

// Bad - changed JSON name
RecurringDays int `json:"recurringDays"`
```

2. **Add Format/Standard Suffixes**:
```go
// Currency codes
CurrencyCodeISO4217 string `json:"currencyCode"`

// Date formats
PaymentExpiryYYYYMMDDHHMMSS string `json:"paymentExpiry"`
```

3. **Add Unit Suffixes**:
```go
// Time durations
RecurringIntervalDays int `json:"recurringInterval"`
```

4. **Don't Add Type Suffixes**:
```go
// Good
Amount float64 `json:"amount"`

// Bad - includes type name
AmountFloat64 float64 `json:"amount"`
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

## Contributing New API Implementations

When implementing a new 2C2P API endpoint:

1. **File Organization**
   - Create API implementation file (e.g., `payment_inquiry.go`)
   - Create test file (e.g., `payment_inquiry_test.go`)
   - Create CLI tool in `cli/api-name/main.go`

2. **Code Structure**
   - Add package documentation with links to relevant 2C2P API docs
   - Define request/response structs with proper JSON tags and field comments
   - Implement the API method on the `Client` struct

3. **Documentation**
   - Include API documentation links in file header
   - Document struct fields with type and required/optional status
   - Add usage examples in package documentation

4. **Testing**
   - Use sample request/response values from 2C2P documentation
   - Test both success and error cases
   - Include JWT token validation tests

5. **CLI Tool**
   - Create focused CLI tool with flags specific to your API
   - Include proper validation and help text
   - Format output to be human-readable

For example, see the Payment Inquiry API implementation:
- API implementation in `payment_inquiry.go`
- Tests in `payment_inquiry_test.go`
- CLI tool in `cli/payment-inquiry/main.go`

