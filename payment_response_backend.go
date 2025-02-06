package api2c2p

import (
	"encoding/xml"
)

type PaymentResponseBackEnd struct {
	XMLName               xml.Name `xml:"PaymentResponse"`
	Version               string   `xml:"version"`
	TimeStamp             string   `xml:"timeStamp"`
	MerchantID            string   `xml:"merchantID"`
	RespCode              string   `xml:"respCode"`
	PAN                   string   `xml:"pan"`
	Amount                string   `xml:"amt"`
	UniqueTransactionCode string   `xml:"uniqueTransactionCode"`
	TranRef               string   `xml:"tranRef"`
	ApprovalCode          string   `xml:"approvalCode"`
	RefNumber             string   `xml:"refNumber"`
	ECI                   string   `xml:"eci"`
	DateTime              string   `xml:"dateTime"`
	Status                string   `xml:"status"`
	FailReason            string   `xml:"failReason"` // can contain successful reason too
	UserDefined1          string   `xml:"userDefined1"`
	UserDefined2          string   `xml:"userDefined2"`
	UserDefined3          string   `xml:"userDefined3"`
	UserDefined4          string   `xml:"userDefined4"`
	UserDefined5          string   `xml:"userDefined5"`
	IPPPeriod             string   `xml:"ippPeriod"`
	IPPInterestType       string   `xml:"ippInterestType"`
	IPPInterestRate       string   `xml:"ippInterestRate"`
	IPPMerchantAbsorbRate string   `xml:"ippMerchantAbsorbRate"`
	PaidChannel           string   `xml:"paidChannel"`
	PaidAgent             string   `xml:"paidAgent"`
	PaymentChannel        string   `xml:"paymentChannel"`
	BackendInvoice        string   `xml:"backendInvoice"`
	IssuerCountry         string   `xml:"issuerCountry"`
	IssuerCountryA3       string   `xml:"issuerCountryA3"`
	BankName              string   `xml:"bankName"`
	CardType              string   `xml:"cardType"`
	ProcessBy             string   `xml:"processBy"`
	PaymentScheme         string   `xml:"paymentScheme"`
	PaymentID             string   `xml:"paymentID"`
	AcquirerResponseCode  string   `xml:"acquirerResponseCode"`
	SchemePaymentID       string   `xml:"schemePaymentID"`
	HashValue             string   `xml:"hashValue"`
}
