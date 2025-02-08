# 2C2P Payment Gateway API Client

A Go client library for integrating with the 2C2P Payment Gateway API (v4.3.1).

## Features

- JWT-based authentication
- Payment Inquiry API support
- Payment Token API support
- SecureFields integration for PCI-compliant card data collection
- QR Payment support (VISA QR, Master Card QR, UPI QR)
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
- [Response Codes](https://developer.2c2p.com/v4.3.1/docs/response-code-payment)
- [Flow Response Codes](https://developer.2c2p.com/v4.3.1/docs/response-code-payment-flow)
- [QR Payment API](https://developer.2c2p.com/v4.3.1/docs/direct-api-method-qr-payment)
- [Refund API](https://developer.2c2p.com/v4.3.1/docs/payment-maintenance-refund-guide)

### API Encryption

The following APIs use encryption to secure sensitive data:

1. **SecureFields API**
   - Backend responses are encrypted using PKCS7
   - Responses are decrypted using the merchant's private key
   - Card data is transmitted in encrypted format

2. **JWT-based APIs**
   - Payment Token API: Request payload is signed using JWT
   - Payment Inquiry API: Request payload is signed using JWT
   - Tokens are generated using the merchant's secret key

3. **Standard HTTPS APIs**
   - QR Payment API uses standard HTTPS without additional encryption layers

## Usage

### Creating a Client

```go
client := api2c2p.NewClient(api2c2p.Config{
    SecretKey:  "your_secret_key",
    MerchantID: "your_merchant_id",
    BaseURL:    "https://sandbox-pgw.2c2p.com", // or https://pgw.2c2p.com for production
})
```

### SecureFields Integration

Before running SecureFields integration:

1. Generate server-to-server key pair:
```bash
go run cmd/server-to-server-key/main.go
```

2. Configure 2C2P merchant portal:
   - Go to Options > Server-to-server API
   - Upload the generated `dist/public_cert.pem` as your Public key
   - Set Frontend return URL to `$serverURL/payment-return`
   - Set Backend return URL to `$serverURL/payment-notify`

3. Start the SecureFields server:
```bash
go run cmd/secure_fields/main.go -merchantID your_merchant_id -secretKey your_secret_key
```

For implementation details, refer to:
- Frontend response handling: See `handlePaymentResponse` in `cmd/secure_fields/main.go`
- Backend notification handling: See `handlePaymentNotification` in `cmd/secure_fields/main.go`
- Response field definitions: See `PaymentResponseBackEnd` in `payment_response_backend.go`

### Processing a Refund

To refund a settled transaction:

```go
refundReq := &api2c2p.RefundRequest{
    InvoiceNo:    "your_invoice_number",
    ActionAmount: 25.00, // Amount to refund
}

refundResp, err := client.Refund(context.Background(), refundReq)
if err != nil {
    log.Fatalf("Failed to process refund: %v", err)
}

// Check response
if refundResp.RespCode == "0000" {
    fmt.Println("Refund successful")
} else {
    fmt.Printf("Refund failed: %s\n", refundResp.RespDesc)
}
```

Note: Refunds can only be processed for settled transactions.

## Code Organization and Implementation Principles

The codebase follows a clear separation of concerns that makes it both testable and maintainable. These principles guide both the existing codebase structure and how new API implementations should be added:

1. **Core SDK Functions**
   - Contains all core business logic and data structures in dedicated files (e.g., `payment_inquiry.go`)
   - Functions are pure and return testable values
   - Handles encryption, decryption, and data transformation
   - Exposes clean interfaces that hide implementation complexity

2. **Comprehensive Testing**
   - Each API has its own test file (e.g., `payment_inquiry_test.go`)
   - Uses mock implementations where needed (e.g., `mockFormValuer`)
   - Tests both success and error scenarios
   - Includes test data files for consistent verification
   - Uses sample request/response values from 2C2P documentation
   - Includes JWT token validation tests

3. **CLI Implementation**
   - Each API has a focused CLI tool in `cmd/api-name/main.go`
   - Remains implementation-agnostic by importing the SDK
   - Includes proper validation and help text
   - Formats output to be human-readable
   - Delegates all business logic to the SDK

4. **Documentation and Structure**
   - Add package documentation with links to relevant 2C2P API docs
   - Document struct fields with type and required/optional status
   - Add usage examples in package documentation
   - Define request/response structs with proper JSON tags and field comments

5. **Request Preparation and Testing**
   - Only applies to methods that make HTTP requests in their function body
   - Each such method has a corresponding `new*Request` helper function
   - Helper functions handle all request preparation (URL, headers, body)
   - This separation allows for comprehensive testing of request construction
   - Tests verify HTTP method, headers, and request body consistency
   - Example:
   ```go
   // PaymentInquiry makes an HTTP request to check payment status
   func (c *Client) PaymentInquiry(ctx context.Context, req *PaymentInquiryRequest) (*PaymentInquiryResponse, error) {
       // Create and make request
       httpReq, err := c.newPaymentInquiryRequest(ctx, req)
       if err != nil {
           return nil, err
       }
       resp, debug, err := c.doRequestWithDebug(httpReq)
       // ... handle response and errors
   }

   // Test that request is constructed correctly
   func TestNewPaymentInquiryRequest(t *testing.T) {
       // Test that given the same input:
       // - HTTP method is always POST
       // - Content-Type is application/json
       // - Request body matches expected JSON
   }
   ```

   Note: This pattern is NOT used for methods that:
   - Don't make HTTP requests
   - Only prepare data for other systems (e.g., SecureFields form data)
   - Handle responses from other systems

This organization ensures:
- The SDK is easy to test in isolation
- Implementation details are encapsulated
- New features can be added without modifying client code
- The codebase remains maintainable and extensible

For example, see the Payment Inquiry API implementation:
- API implementation in `payment_inquiry.go`
- Tests in `payment_inquiry_test.go`
- CLI tool in `cmd/payment-inquiry/main.go`

## Code Style and Practices

### Code Organization

1. **Unexport Unused Types and Functions**:
   - Types, functions, and methods that are not used in `cmd/*.go` should be unexported
   - This keeps the public API surface minimal and focused on actual usage
   ```go
   // Good - unexported since only used internally
   type debugInfo struct {
       Request  *debugRequest
       Response *debugResponse
   }

   // Bad - exported but not used in cmd/*.go
   type DebugInfo struct {
       Request  *DebugRequest
       Response *DebugResponse
   }
   ```

2. **Export Struct Fields for JSON**:
   - Prefer exported struct fields over custom `MarshalJSON`/`UnmarshalJSON` methods
   - This reduces code complexity and improves maintainability
   ```go
   // Good - uses exported field
   type UIParams struct {
       UserInfo *UserInfo `json:"userInfo,omitempty"`
   }

   // Bad - requires custom marshal/unmarshal
   type UIParams struct {
       userInfo *userInfo `json:"userInfo"`
   }
   ```

### Field Naming Conventions

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
