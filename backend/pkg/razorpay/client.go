package razorpay

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const apiBase = "https://api.razorpay.com/v1"

// Client talks to Razorpay Orders and verifies payment signatures.
type Client struct {
	KeyID     string
	KeySecret string
	HTTP      *http.Client
}

// NewClient returns a Razorpay API client.
func NewClient(keyID, keySecret string) *Client {
	return &Client{
		KeyID:     strings.TrimSpace(keyID),
		KeySecret: strings.TrimSpace(keySecret),
		HTTP:      &http.Client{Timeout: 30 * time.Second},
	}
}

type createOrderRequest struct {
	Amount   int64             `json:"amount"`
	Currency string            `json:"currency"`
	Receipt  string            `json:"receipt"`
	Notes    map[string]string `json:"notes,omitempty"`
}

type createOrderResponse struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

// CreateOrder creates a Razorpay order. amountPaise must be at least 100 (₹1).
func (c *Client) CreateOrder(amountPaise int64, receipt string, notes map[string]string) (*createOrderResponse, error) {
	if amountPaise < 100 {
		return nil, fmt.Errorf("razorpay: amount must be at least 100 paise")
	}
	receipt = strings.TrimSpace(receipt)
	if receipt == "" {
		return nil, fmt.Errorf("razorpay: receipt is required")
	}
	if len(receipt) > 40 {
		receipt = receipt[:40]
	}

	body, err := json.Marshal(createOrderRequest{
		Amount:   amountPaise,
		Currency: "INR",
		Receipt:  receipt,
		Notes:    notes,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, apiBase+"/orders", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.KeyID, c.KeySecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("razorpay create order: %s", strings.TrimSpace(string(raw)))
	}

	var out createOrderResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if strings.TrimSpace(out.ID) == "" {
		return nil, fmt.Errorf("razorpay create order: empty order id")
	}
	return &out, nil
}

// VerifyPaymentSignature checks the HMAC for order_id|payment_id.
func VerifyPaymentSignature(orderID, paymentID, signature, secret string) bool {
	orderID = strings.TrimSpace(orderID)
	paymentID = strings.TrimSpace(paymentID)
	signature = strings.TrimSpace(signature)
	secret = strings.TrimSpace(secret)
	if orderID == "" || paymentID == "" || signature == "" || secret == "" {
		return false
	}
	payload := orderID + "|" + paymentID
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
