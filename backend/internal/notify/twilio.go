package notify

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TwilioSMS sends SMS via Twilio REST API (Messages resource).
type TwilioSMS struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	HTTPClient *http.Client
}

func (t *TwilioSMS) client() *http.Client {
	if t.HTTPClient != nil {
		return t.HTTPClient
	}
	return &http.Client{Timeout: 15 * time.Second}
}

// Send posts a new outbound SMS. Body should stay within carrier limits.
func (t *TwilioSMS) Send(ctx context.Context, toE164, body string) error {
	if t.AccountSID == "" || t.AuthToken == "" || t.FromNumber == "" {
		return fmt.Errorf("twilio: not configured")
	}
	if toE164 == "" {
		return fmt.Errorf("twilio: empty destination")
	}

	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.AccountSID)
	form := url.Values{}
	form.Set("To", toE164)
	form.Set("From", t.FromNumber)
	form.Set("Body", body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.AccountSID, t.AuthToken)

	resp, err := t.client().Do(req)
	if err != nil {
		return fmt.Errorf("twilio: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("twilio: status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return nil
}

// BookingConfirmationBody formats SMS per product requirements.
func BookingConfirmationBody(bookingID, carName, carPlate, price string) string {
	return fmt.Sprintf(
		"Your booking is confirmed. Booking ID: %s. Car: %s (%s). Final agreed price: %s.",
		bookingID, carName, carPlate, price,
	)
}
