package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"carmanage/backend/internal/httpx"
	"carmanage/backend/internal/models"
	"carmanage/backend/pkg/razorpay"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PaymentCheckoutInfo tells the client whether to use Razorpay Checkout or simulated pay.
type PaymentCheckoutInfo struct {
	Provider      string `json:"provider"` // "razorpay" or "simulated"
	RazorpayKeyID string `json:"razorpay_key_id,omitempty"`
}

// PaymentCheckoutInfo returns checkout mode from config.
func (s *BookingService) PaymentCheckoutInfo() PaymentCheckoutInfo {
	if s.Config != nil && s.Config.RazorpayEnabled() {
		return PaymentCheckoutInfo{
			Provider:      "razorpay",
			RazorpayKeyID: s.Config.RazorpayKeyID,
		}
	}
	return PaymentCheckoutInfo{Provider: "simulated"}
}

// RazorpayOrderResult is returned when creating a Razorpay order for checkout.
type RazorpayOrderResult struct {
	OrderID     string
	AmountPaise int64
	Currency    string
	KeyID       string
}

// RazorpayPaymentProof is sent by the client after successful Checkout.
type RazorpayPaymentProof struct {
	OrderID   string
	PaymentID string
	Signature string
}

func (s *BookingService) razorpayClient() *razorpay.Client {
	return razorpay.NewClient(s.Config.RazorpayKeyID, s.Config.RazorpayKeySecret)
}

func dueAmountInrForPayment(b *models.Booking, sv PaymentSettlementView) (decimal.Decimal, error) {
	switch b.PaymentStatus {
	case models.BookingPaymentUnpaid:
		if sv.DepositDueInr.IsZero() {
			return decimal.Zero, httpx.WrapValidation("Deposit amount is invalid.")
		}
		return sv.DepositDueInr, nil
	case models.BookingPaymentFinalDue:
		if sv.FinalDueInr.IsZero() {
			return decimal.Zero, httpx.WrapValidation("Final balance is invalid.")
		}
		return sv.FinalDueInr, nil
	default:
		return decimal.Zero, httpx.ErrPaymentNotReady
	}
}

func inrToPaise(d decimal.Decimal) int64 {
	return d.Mul(decimal.NewFromInt(100)).Round(0).IntPart()
}

func paymentReceipt(bookingID uuid.UUID, phase string) string {
	r := fmt.Sprintf("%s-%s", bookingID.String(), phase)
	if len(r) > 40 {
		return r[:40]
	}
	return r
}

// CustomerCreateRazorpayOrder creates a Razorpay order for the amount due now (deposit or final).
func (s *BookingService) CustomerCreateRazorpayOrder(ctx context.Context, customerID, bookingID uuid.UUID) (*RazorpayOrderResult, error) {
	if s.Config == nil || !s.Config.RazorpayEnabled() {
		return nil, httpx.NewError(503, "PAYMENTS_UNAVAILABLE", "Online payments are not configured. Use demo checkout or contact support.")
	}

	b, bd, err := s.customerBookingReadyForPayment(ctx, customerID, bookingID)
	if err != nil {
		return nil, err
	}
	sv := s.SettlementView(b, bd)
	due, err := dueAmountInrForPayment(b, sv)
	if err != nil {
		return nil, err
	}
	paise := inrToPaise(due)
	if paise < 100 {
		return nil, httpx.WrapValidation("Payment amount is too small.")
	}

	phase := sv.Phase
	if phase == "" {
		phase = "pay"
	}
	notes := map[string]string{
		"booking_id": bookingID.String(),
		"phase":      phase,
	}
	order, err := s.razorpayClient().CreateOrder(paise, paymentReceipt(bookingID, phase), notes)
	if err != nil {
		return nil, httpx.NewError(502, "PAYMENT_GATEWAY_ERROR", "Could not start payment. Please try again.")
	}

	return &RazorpayOrderResult{
		OrderID:     order.ID,
		AmountPaise: order.Amount,
		Currency:    order.Currency,
		KeyID:       s.Config.RazorpayKeyID,
	}, nil
}

func (s *BookingService) customerBookingReadyForPayment(ctx context.Context, customerID, bookingID uuid.UUID) (*models.Booking, PaymentBreakdown, error) {
	b, err := s.Repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, PaymentBreakdown{}, httpx.ErrNotFound
		}
		return nil, PaymentBreakdown{}, err
	}
	if b.CustomerID != customerID {
		return nil, PaymentBreakdown{}, httpx.ErrForbidden
	}
	if b.Status != models.BookingConfirmed {
		return nil, PaymentBreakdown{}, httpx.ErrPaymentNotReady
	}
	if b.FinalBookingPrice == nil {
		return nil, PaymentBreakdown{}, httpx.ErrPaymentNotReady
	}
	if b.PaymentStatus == models.BookingPaymentPaid {
		return nil, PaymentBreakdown{}, httpx.ErrPaymentAlreadyPaid
	}
	if b.PaymentStatus == models.BookingPaymentDepositPaid {
		return nil, PaymentBreakdown{}, httpx.ErrSettlementNotReady
	}
	bd, err := s.BreakdownForBooking(b)
	if err != nil {
		return nil, PaymentBreakdown{}, err
	}
	return b, bd, nil
}

func (s *BookingService) verifyRazorpayProof(proof RazorpayPaymentProof) error {
	proof.OrderID = strings.TrimSpace(proof.OrderID)
	proof.PaymentID = strings.TrimSpace(proof.PaymentID)
	proof.Signature = strings.TrimSpace(proof.Signature)
	if proof.OrderID == "" || proof.PaymentID == "" || proof.Signature == "" {
		return httpx.ErrPaymentVerificationFailed
	}
	if !razorpay.VerifyPaymentSignature(proof.OrderID, proof.PaymentID, proof.Signature, s.Config.RazorpayKeySecret) {
		return httpx.ErrPaymentVerificationFailed
	}
	return nil
}
