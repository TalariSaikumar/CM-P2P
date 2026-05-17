package razorpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyPaymentSignature_valid(t *testing.T) {
	secret := "test_secret"
	orderID := "order_abc"
	paymentID := "pay_xyz"
	payload := orderID + "|" + paymentID
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))

	if !VerifyPaymentSignature(orderID, paymentID, sig, secret) {
		t.Fatal("expected valid signature")
	}
}

func TestVerifyPaymentSignature_invalid(t *testing.T) {
	if VerifyPaymentSignature("order_abc", "pay_xyz", "bad", "test_secret") {
		t.Fatal("expected invalid signature")
	}
}
