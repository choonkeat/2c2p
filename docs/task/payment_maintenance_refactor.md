# Payment Maintenance Operations Refactor

## Background

Currently, the refund operation is implemented in `refund.go`. We need to implement additional payment maintenance operations (VoidCancel, RefundStatusInquiry, and SettlePayment) that share similar request/response patterns and encryption requirements.

## Objective

Refactor the payment maintenance operations to:
1. Share common implementation logic while maintaining distinct type safety
2. Implement new operations: VoidCancel, RefundStatusInquiry, and SettlePayment
3. Keep code DRY (Don't Repeat Yourself) while ensuring type safety

## Implementation Plan

### Phase 1: Move Existing Refund Implementation

- [ ] Create new `payment_maintenance.go` file in root package
- [ ] Move existing code from `refund.go` without changes:
  - [ ] Move `PaymentProcessRequest` and `RefundResponse` types
  - [ ] Move `StringClaims` and its methods
  - [ ] Move encryption helper functions
  - [ ] Move `NewRefundRequest` function
  - [ ] Move `Refund` function
- [ ] Run `make test` to ensure everything still works
- [ ] Clean up old `refund.go` file

### Phase 2: Refactor for Common Implementation

- [ ] Rename types to be more generic:
  - [ ] `PaymentProcessRequest` -> `PaymentMaintenanceRequest`
  - [ ] `RefundResponse` -> `PaymentMaintenanceResponse`
- [ ] Extract common fields into base types
- [ ] Create operation-specific request/response types extending base types
- [ ] Implement common helper function:
```go
func processPaymentMaintenance(
    ctx context.Context,
    req interface{},
    processType string,
) (interface{}, error)
```
- [ ] Update existing Refund operation to use new common implementation
- [ ] Run `make test` to verify refactoring

### Phase 3: Implement New Operations

- [ ] Implement VoidCancel:
  - [ ] Define request/response types
  - [ ] Implement `VoidCancel` method
  - [ ] Add tests
- [ ] Implement RefundStatusInquiry:
  - [ ] Define request/response types
  - [ ] Implement `RefundStatusInquiry` method
  - [ ] Add tests
- [ ] Implement SettlePayment:
  - [ ] Define request/response types
  - [ ] Implement `SettlePayment` method
  - [ ] Add tests
- [ ] Run `make test` after each operation implementation

### Phase 4: Documentation and Examples

- [ ] Update API documentation
- [ ] Add usage examples for each operation
- [ ] Update README.md with new operations
- [ ] Add integration test examples

## Success Criteria

- [ ] All existing tests pass after moving refund implementation
- [ ] All payment maintenance operations implemented and tested
- [ ] Code reuse through common helper functions
- [ ] Type safety maintained for each operation
- [ ] Documentation complete and up-to-date

## Notes

- Each phase should complete with passing tests before moving to next phase
- Process types:
  - "R" for Refund
  - "V" for VoidCancel
  - "S" for Settle
- Error handling consistent across all operations
- Maintain backward compatibility throughout refactoring

## Important Implementation Guidelines

### Code Organization Principles

1. **Core SDK Functions**
   - [ ] Place all core business logic in dedicated files
   - [ ] Keep functions pure and return testable values
   - [ ] Handle encryption, decryption, and data transformation
   - [ ] Expose clean interfaces that hide implementation complexity

2. **Request Preparation Pattern**
   - [ ] Implement `new*Request` helper functions for HTTP requests
   - [ ] Helper functions handle all request preparation (URL, headers, body)
   - [ ] Separate request construction from execution for better testing

3. **Code Style**
   - [ ] Keep public API surface minimal and focused
   - [ ] Follow existing naming conventions

### Testing Guidelines

1. **Test Verification**
   - [ ] Run `make test` after each significant change
   - [ ] Compiler errors indicate incomplete code changes
   - [ ] Test failures likely indicate implementation bugs (DO NOT modify test expectations)

2. **Test Coverage**
   - [ ] Each API has its own test file
   - [ ] Test both success and error scenarios
   - [ ] Include JWT token validation tests
   - [ ] Use sample request/response values from 2C2P documentation

3. **Documentation**
   - [ ] Add package documentation with links to 2C2P API docs
   - [ ] Document struct fields with type and required/optional status
   - [ ] Add usage examples in package documentation
