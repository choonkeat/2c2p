package api2c2p

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
)

// PaymentProcessRequest represents a refund request
type PaymentProcessRequest struct {
	XMLName         xml.Name `xml:"PaymentProcessRequest"`
	Version         string   `xml:"version"`
	TimeStamp       *string  `xml:"timeStamp,omitempty"`
	MerchantID      string   `xml:"merchantID"`
	InvoiceNo       string   `xml:"invoiceNo"`
	ChildMerchantID *string  `xml:"childMerchantID,omitempty"`
	ActionAmount    Dollars  `xml:"actionAmount"`
	ProcessType     string   `xml:"processType"`
	BankCode        *string  `xml:"bankCode,omitempty"`
	AccountName     *string  `xml:"accountName,omitempty"`
	AccountNumber   *string  `xml:"accountNumber,omitempty"`
	SubMerchantList *struct {
		SubMerchant []struct {
			SubMID          string  `xml:"subMID,attr"`
			SubAmount       float64 `xml:"subAmount,attr"`
			LoyaltyPayments *struct {
				LoyaltyRefund []LoyaltyRefund `xml:"loyaltyRefund"`
			} `xml:"loyaltyPayments,omitempty"`
		} `xml:"subMerchant"`
	} `xml:"subMerchantList,omitempty"`
	NotifyURL       *string `xml:"notifyURL,omitempty"`
	IdempotencyID   *string `xml:"idempotencyID,omitempty"`
	LoyaltyPayments *struct {
		LoyaltyRefund []LoyaltyRefund `xml:"loyaltyRefund"`
	} `xml:"loyaltyPayments,omitempty"`
}

type LoyaltyRefund struct {
	LoyaltyProvider         string  `xml:"loyaltyProvider,omitempty"`
	ExternalMerchantID      string  `xml:"externalMerchantId,omitempty"`
	TotalRefundRewardAmount Dollars `xml:"totalRefundRewardAmount,omitempty"`
	RefundRewards           *struct {
		Reward []RefundReward `xml:"reward"`
	} `xml:"refundRewards,omitempty"`
}

type RefundReward struct {
	Type     string  `xml:"type,omitempty"`
	Quantity float64 `xml:"quantity,omitempty"`
}

// RefundResponse represents the response from a refund request
type RefundResponse struct {
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

// Refund processes a refund request for a previously successful payment
func (c *Client) Refund(ctx context.Context, invoiceNo string, amount Cents) (*RefundResponse, error) {
	// Create refund request
	req := &PaymentProcessRequest{
		Version:      "4.3",
		TimeStamp:    nil, // No timestamp as requested
		MerchantID:   c.MerchantID,
		InvoiceNo:    invoiceNo,
		ActionAmount: amount.ToDollars(),
		ProcessType:  "R",
		// LoyaltyPayments: &struct {
		// 	LoyaltyRefund []LoyaltyRefund `xml:"loyaltyRefund"`
		// }{
		// 	LoyaltyRefund: []LoyaltyRefund{
		// 		{
		// 			TotalRefundRewardAmount: amount.ToDollars(),
		// 			RefundRewards:           nil,
		// 		},
		// 	},
		// },
	}

	// Create HTTP request
	httpReq, err := c.NewRefundRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	decrypted, err := c.verifyJWSAndDecryptJWE(string(body))
	if err != nil {
		return nil, fmt.Errorf("verify and decrypt JWS JWE: %w", err)
	}

	var refundResp RefundResponse
	if err := xml.NewDecoder(bytes.NewReader(decrypted)).Decode(&refundResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &refundResp, nil
}

func jwsWithRawPayload(privateKey *rsa.PrivateKey, token *jwt.Token, payload []byte) (string, error) {
	h, err := json.Marshal(token.Header)
	if err != nil {
		return "", err
	}

	sstr := token.EncodeSegment(h) + "." + token.EncodeSegment(payload)

	sig, err := token.Method.Sign(sstr, privateKey)
	if err != nil {
		return "", err
	}

	return sstr + "." + token.EncodeSegment(sig), nil
}

// verifyJWSAndDecryptJWE verifies a JWS token using the public key and decrypts the JWE payload using the private key.
// The inputToken string should be a JWS token containing a JWE payload.
func (c *Client) verifyJWSAndDecryptJWE(inputToken string) ([]byte, error) {
	publicKey, ok := c.ServerJWTPublicCert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("convert public key to RSA public key")
	}

	// Parse and verify JWS
	jws, err := jose.ParseSigned(inputToken, []jose.SignatureAlgorithm{jose.PS256})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWS: %w", err)
	}

	// Verify JWS signature and get payload
	jweTokenBytes, err := jws.Verify(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to verify JWS signature: %w", err)
	}

	// Parse JWE token
	object, err := jose.ParseEncrypted(string(jweTokenBytes), []jose.KeyAlgorithm{jose.RSA_OAEP}, []jose.ContentEncryption{jose.A256GCM})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE token: %w", err)
	}

	// Decrypt JWE token
	decrypted, err := object.Decrypt(c.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt JWE token: %w", err)
	}

	return decrypted, nil
}

func (c *Client) encryptJWEAndSignJWS(xmlData []byte) (string, error) {
	// Encrypt with JWE
	// Create encrypter
	encrypter, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{
			Algorithm: jose.RSA_OAEP,
			Key:       c.ServerJWTPublicCert.PublicKey,
		},
		// this option means to include `"typ": "JWE"` in header
		// but sample request in https://developer.2c2p.com/v4.3.1/docs/payment-maintenance-refund-guide
		// only has { "alg": "RSA-OAEP", "enc": "A256GCM" } without `"typ"`
		(&jose.EncrypterOptions{}).WithType("JWE"),
	)
	if err != nil {
		return "", fmt.Errorf("create encrypter: %w", err)
	}

	// Encrypt data
	log.Printf("[DEBUG] Encrypting data: %s", string(xmlData))
	jwe, err := encrypter.Encrypt(xmlData)
	if err != nil {
		return "", fmt.Errorf("encrypt data: %w", err)
	}

	// Serialize to compact form
	jweToken, err := jwe.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("serialize JWE: %w", err)
	}
	log.Printf("jweToken: %s", jweToken)

	// Then sign with JWS PS256
	// https://developer.2c2p.com/v4.3.1/recipes/prepare-request-payload-with-jwt-jws-with-keys
	// https://developer.2c2p.com/v4.3.1/docs/payment-maintenance-refund-guide
	token := jwt.New(jwt.SigningMethodPS256)

	// Sign the token
	signedJWE, err := jwsWithRawPayload(c.PrivateKey, token, []byte(jweToken))
	if err != nil {
		return "", fmt.Errorf("jwsWithRawPayload: %w", err)
	}
	return signedJWE, nil
}

// NewRefundRequest creates a new HTTP request for refunding a payment
func (c *Client) NewRefundRequest(ctx context.Context, req *PaymentProcessRequest) (*http.Request, error) {
	// Marshal request to XML
	xmlData, err := xml.MarshalIndent(req, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	log.Printf("[DEBUG] Request XML: %s", string(xmlData))

	// Sign the token
	signedJWE, err := c.encryptJWEAndSignJWS(xmlData)
	if err != nil {
		return nil, fmt.Errorf("jwsWithRawPayload: %w", err)
	}

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.frontendEndpoint("2C2PFrontend/PaymentAction/2.0/action"), strings.NewReader(signedJWE))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "text/plain")
	return httpReq, nil
}
