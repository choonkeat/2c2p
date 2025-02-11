package api2c2p

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	publicKey, err := c.serverPublicKey()
	if err != nil {
		return nil, fmt.Errorf("read server public key file: %#v: %w", c.ServerPublicKeyFile, err)
	}
	privateKey, _, err := c.loadPrivateKeyAndCert()
	if err != nil {
		return nil, fmt.Errorf("load private key and cert: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	decrypted, err := verifyAndDecryptJWSJWE(string(body), publicKey, privateKey)
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

// verifyAndDecryptJWSJWE verifies a JWS token using the public key and decrypts the JWE payload using the private key.
// The input string should be a JWS token containing a JWE payload.
func verifyAndDecryptJWSJWE(input string, publicKey any, privateKey any) ([]byte, error) {
	// Parse and verify JWS
	jws, err := jose.ParseSigned(input, []jose.SignatureAlgorithm{jose.PS256})
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
	decrypted, err := object.Decrypt(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt JWE token: %w", err)
	}

	return decrypted, nil
}

// NewRefundRequest creates a new HTTP request for refunding a payment
func (c *Client) NewRefundRequest(ctx context.Context, req *PaymentProcessRequest) (*http.Request, error) {
	// Marshal request to XML
	xmlData, err := xml.MarshalIndent(req, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Encrypt with JWE
	jweToken, err := c.encryptWithJWE(xmlData)
	if err != nil {
		return nil, fmt.Errorf("encrypt token: %w", err)
	}
	log.Printf("jweToken: %s", jweToken)

	// Then sign with JWS PS256
	// https://developer.2c2p.com/v4.3.1/recipes/prepare-request-payload-with-jwt-jws-with-keys
	// https://developer.2c2p.com/v4.3.1/docs/payment-maintenance-refund-guide
	token := jwt.New(jwt.SigningMethodPS256)
	token.Header["kid"] = "choonkeat-dist-public-cert"

	// Load private key for signing
	privateKey, _, err := c.loadPrivateKeyAndCert()
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}

	// Sign the token
	signedToken, err := jwsWithRawPayload(privateKey, token, []byte(jweToken))
	if err != nil {
		return nil, fmt.Errorf("jwsWithRawPayload: %w", err)
	}

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.frontendEndpoint("2C2PFrontend/PaymentAction/2.0/action"), strings.NewReader(signedToken))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "text/plain")
	return httpReq, nil
}

func (c *Client) serverPublicKey() (*rsa.PublicKey, error) {
	// Read and parse 2C2P's public key certificate
	certPEM, err := os.ReadFile(c.ServerPublicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read server public key file: %#v: %w", c.ServerPublicKeyFile, err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode server public key PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse server certificate: %w", err)
	}
	return cert.PublicKey.(*rsa.PublicKey), nil
}

func (c *Client) encryptWithJWE(data []byte) (string, error) {
	log.Printf("[DEBUG] Encrypting with %s", c.ServerPublicKeyFile)

	// Read and parse 2C2P's public key certificate
	publicKey, err := c.serverPublicKey()
	if err != nil {
		return "", fmt.Errorf("read server public key file: %#v: %w", c.ServerPublicKeyFile, err)
	}

	// Create encrypter
	encrypter, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{
			Algorithm: jose.RSA_OAEP,
			Key:       publicKey,
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
	log.Printf("[DEBUG] Encrypting data: %s", string(data))
	jwe, err := encrypter.Encrypt(data)
	if err != nil {
		return "", fmt.Errorf("encrypt data: %w", err)
	}

	// Serialize to compact form
	serialized, err := jwe.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("serialize JWE: %w", err)
	}

	return serialized, nil
}

func (c *Client) loadPrivateKeyAndCert() (*rsa.PrivateKey, *x509.Certificate, error) {
	// Read the combined PEM file
	pemData, err := os.ReadFile(c.PrivateKeyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("read private key file: %w", err)
	}

	// Parse private key
	var privateKey *rsa.PrivateKey
	var cert *x509.Certificate
	for {
		block, rest := pem.Decode(pemData)
		if block == nil {
			break
		}
		switch block.Type {
		case "RSA PRIVATE KEY", "PRIVATE KEY":
			if privateKey == nil {
				if block.Type == "RSA PRIVATE KEY" {
					privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
				} else {
					var key interface{}
					key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
					if err == nil {
						var ok bool
						privateKey, ok = key.(*rsa.PrivateKey)
						if !ok {
							err = fmt.Errorf("not an RSA private key")
						}
					}
				}
				if err != nil {
					return nil, nil, fmt.Errorf("parse private key: %w", err)
				}
			}
		case "CERTIFICATE":
			if cert == nil {
				cert, err = x509.ParseCertificate(block.Bytes)
				if err != nil {
					return nil, nil, fmt.Errorf("parse certificate: %w", err)
				}
			}
		}
		pemData = rest
	}

	if privateKey == nil {
		return nil, nil, fmt.Errorf("no private key found")
	}
	if cert == nil {
		return nil, nil, fmt.Errorf("no certificate found")
	}

	return privateKey, cert, nil
}
