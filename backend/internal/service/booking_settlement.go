package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"carmanage/backend/internal/httpx"
	"carmanage/backend/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// DepositPercentOfCustomerTotal is the upfront share of the agreed trip customer total (inclusive of platform fees in that total).
const DepositPercentOfCustomerTotal = 75

// PaymentSettlementView is derived state for APIs (deposit, post-trip, final balance).
type PaymentSettlementView struct {
	Phase                   string
	DepositPercent          int
	DepositDueInr           decimal.Decimal
	DepositPaidInr          decimal.Decimal
	TripBalanceInr          decimal.Decimal
	PostTripChargesInr      decimal.Decimal
	FinalDueInr             decimal.Decimal
	OwnerProjectedPayoutInr decimal.Decimal
}

// SettlementView computes UI amounts for the two-phase payment flow.
func (s *BookingService) SettlementView(b *models.Booking, bd PaymentBreakdown) PaymentSettlementView {
	pct := decimal.NewFromInt(DepositPercentOfCustomerTotal).Div(decimal.NewFromInt(100))
	trip := bd.CustomerTotal.Round(2)
	post := b.PostTripChargesTotal
	if post.IsNegative() {
		post = decimal.Zero
	}
	post = post.Round(2)
	out := PaymentSettlementView{
		DepositPercent:          DepositPercentOfCustomerTotal,
		PostTripChargesInr:      post,
		OwnerProjectedPayoutInr: bd.OwnerNet.Add(post).Round(2),
	}

	switch b.PaymentStatus {
	case models.BookingPaymentUnpaid:
		out.Phase = "unpaid_deposit"
		out.DepositDueInr = trip.Mul(pct).Round(2)
		out.TripBalanceInr = trip.Sub(out.DepositDueInr).Round(2)
	case models.BookingPaymentDepositPaid:
		out.Phase = "awaiting_settlement"
		if b.DepositCustomerTotal != nil {
			out.DepositPaidInr = b.DepositCustomerTotal.Round(2)
		}
		out.TripBalanceInr = trip.Sub(out.DepositPaidInr).Round(2)
	case models.BookingPaymentFinalDue:
		out.Phase = "final_due"
		if b.DepositCustomerTotal != nil {
			out.DepositPaidInr = b.DepositCustomerTotal.Round(2)
		}
		out.TripBalanceInr = trip.Sub(out.DepositPaidInr).Round(2)
		out.FinalDueInr = trip.Add(post).Sub(out.DepositPaidInr).Round(2)
		if out.FinalDueInr.IsNegative() {
			out.FinalDueInr = decimal.Zero
		}
	case models.BookingPaymentPaid:
		out.Phase = "paid"
		if b.DepositCustomerTotal != nil {
			out.DepositPaidInr = b.DepositCustomerTotal.Round(2)
		} else if b.CustomerTotalPaid != nil {
			out.DepositPaidInr = b.CustomerTotalPaid.Round(2)
		}
		out.TripBalanceInr = decimal.Zero
		out.FinalDueInr = decimal.Zero
		out.OwnerProjectedPayoutInr = bd.OwnerNet.Round(2)
	default:
		out.Phase = "unpaid_deposit"
		out.DepositDueInr = trip.Mul(pct).Round(2)
		out.TripBalanceInr = trip.Sub(out.DepositDueInr).Round(2)
	}
	return out
}

// PostTripChargeLine is owner input for post-trip settlement.
type PostTripChargeLine struct {
	Label     string
	AmountInr decimal.Decimal
}

// OwnerPutPostTripCharges replaces post-trip lines and sets payment to FINAL_DUE (deposit must already be paid).
func (s *BookingService) OwnerPutPostTripCharges(ctx context.Context, ownerID, bookingID uuid.UUID, lines []PostTripChargeLine) (*models.Booking, error) {
	b, err := s.Repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	if b.OwnerID != ownerID {
		return nil, httpx.ErrForbidden
	}
	if b.Status != models.BookingConfirmed {
		return nil, httpx.NewError(http.StatusConflict, "BOOKING_STATE", "Settlement applies only to confirmed bookings.")
	}
	switch b.PaymentStatus {
	case models.BookingPaymentDepositPaid, models.BookingPaymentFinalDue:
		// ok — owner may revise lines until the customer pays the final balance
	case models.BookingPaymentUnpaid:
		return nil, httpx.ErrDepositRequiredForSettlement
	case models.BookingPaymentPaid:
		return nil, httpx.ErrSettlementLockedPaid
	default:
		return nil, httpx.ErrConflict
	}
	if !CustomerReturnRecorded(b) {
		return nil, httpx.ErrCustomerReturnRequired
	}

	const maxLines = 30
	maxLine := decimal.NewFromInt(500_000)
	maxSum := decimal.NewFromInt(2_000_000)
	if len(lines) > maxLines {
		return nil, httpx.WrapValidation("Too many line items (max 30).")
	}
	var rows []models.BookingPostTripCharge
	sum := decimal.Zero
	for _, ln := range lines {
		label := strings.TrimSpace(ln.Label)
		amt := ln.AmountInr.Round(2)
		if label == "" {
			return nil, httpx.WrapValidation("Each charge needs a label.")
		}
		if len(label) > 240 {
			return nil, httpx.WrapValidation("Each label must be at most 240 characters.")
		}
		if amt.IsNegative() {
			return nil, httpx.WrapValidation("Amounts cannot be negative.")
		}
		if amt.GreaterThan(maxLine) {
			return nil, httpx.WrapValidation("Single line amount is too large.")
		}
		if amt.IsZero() {
			continue
		}
		sum = sum.Add(amt)
		rows = append(rows, models.BookingPostTripCharge{
			BookingID: bookingID,
			Label:     label,
			AmountInr: amt,
		})
	}
	sum = sum.Round(2)
	if sum.GreaterThan(maxSum) {
		return nil, httpx.WrapValidation("Total post-trip charges exceed the allowed limit.")
	}

	now := time.Now().UTC()
	if err := s.Repo.SavePostTripSettlement(ctx, bookingID, rows, sum, models.BookingPaymentFinalDue, now); err != nil {
		return nil, err
	}
	return s.Repo.GetBookingByID(ctx, bookingID)
}

// OwnerPutPostTripChargesInput is parsed from JSON (amounts as decimal strings).
type OwnerPutPostTripChargesInput struct {
	Items []struct {
		Label     string `json:"label"`
		AmountInr string `json:"amount_inr"`
	} `json:"items"`
}

// ParsePostTripChargeLines validates owner JSON into decimal lines.
func ParsePostTripChargeLines(in OwnerPutPostTripChargesInput) ([]PostTripChargeLine, error) {
	if in.Items == nil {
		return nil, httpx.WrapValidation(`Send an "items" array (use [] for no extra charges).`)
	}
	out := make([]PostTripChargeLine, 0, len(in.Items))
	for _, it := range in.Items {
		raw := strings.TrimSpace(it.AmountInr)
		if raw == "" {
			raw = "0"
		}
		d, err := decimal.NewFromString(raw)
		if err != nil {
			return nil, httpx.WrapValidation("Each amount_inr must be a valid decimal.")
		}
		if d.IsZero() {
			continue
		}
		label := strings.TrimSpace(it.Label)
		if label == "" {
			return nil, httpx.WrapValidation("Each non-zero charge needs a label.")
		}
		out = append(out, PostTripChargeLine{Label: label, AmountInr: d})
	}
	return out, nil
}
