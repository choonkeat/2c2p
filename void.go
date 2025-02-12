package api2c2p

import (
	"context"
	"encoding/xml"
	"fmt"
)

// VoidCancelRequest represents a void/cancel request
type VoidCancelRequest struct {
	InvoiceNo       string  `xml:"invoiceNo"`
	MerchantID      string  `xml:"merchantID,omitempty"`
	ActionAmount    Dollars `xml:"actionAmount"`
	ProcessType     string  `xml:"processType"` // Always "V" for void/cancel
	IdempotencyID   *string `xml:"idempotencyID,omitempty"`
	ChildMerchantID *string `xml:"childMerchantID,omitempty"`
}

// VoidCancelResponse represents the response from a void/cancel request
type VoidCancelResponse struct {
	XMLName        xml.Name `xml:"PaymentProcessResponse"`
	Version        string   `xml:"version"`
	TimeStamp      string   `xml:"timeStamp"`
	MerchantID     string   `xml:"merchantID"`
	InvoiceNo      string   `xml:"invoiceNo,omitempty"`
	ActionAmount   string   `xml:"actionAmount,omitempty"`
	ProcessType    string   `xml:"processType"`
	RespCode       string   `xml:"respCode"`
	RespDesc       string   `xml:"respDesc"`
	ApprovalCode   string   `xml:"approvalCode,omitempty"`
	ReferenceNo    string   `xml:"referenceNo,omitempty"`
	TransactionID  string   `xml:"transactionID,omitempty"`
	TransactionRef string   `xml:"transactionRef,omitempty"`
}

// VoidCancel processes a void/cancel request for a previously successful payment
func (c *Client) VoidCancel(ctx context.Context, req *VoidCancelRequest) (*VoidCancelResponse, error) {
	if req.InvoiceNo == "" {
		return nil, fmt.Errorf("invoice number is required")
	}
	if req.ActionAmount.ToCents() <= 0 {
		return nil, fmt.Errorf("action amount must be greater than 0")
	}
	if req.MerchantID == "" {
		req.MerchantID = c.MerchantID
	}

	// Always set process type to "V" for void/cancel
	req.ProcessType = "V"

	// Create payment process request
	processReq := &PaymentProcessRequest{
		Version:         "3.8",
		MerchantID:      req.MerchantID,
		InvoiceNo:       req.InvoiceNo,
		ActionAmount:    req.ActionAmount,
		ProcessType:     req.ProcessType,
		IdempotencyID:   req.IdempotencyID,
		ChildMerchantID: req.ChildMerchantID,
	}

	// Process the request
	var resp VoidCancelResponse
	if err := c.PerformPaymentProcess(ctx, processReq, &resp); err != nil {
		return nil, fmt.Errorf("failed to process void/cancel request: %w", err)
	}

	return &resp, nil
}
