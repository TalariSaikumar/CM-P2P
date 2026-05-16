package service

import (
	"testing"
	"time"

	"carmanage/backend/internal/models"

	"github.com/shopspring/decimal"
)

func TestTripDaysInclusive(t *testing.T) {
	from := time.Date(2026, 8, 9, 10, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 10, 18, 0, 0, 0, time.UTC)
	if got := TripDaysInclusive(from, to); got != 2 {
		t.Fatalf("expected 2 days, got %d", got)
	}
	if TripDaysInclusive(from, from) != 1 {
		t.Fatal("same day should be 1")
	}
}

func TestTripDaysInclusiveEndOfDay(t *testing.T) {
	from := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 16, 23, 59, 59, 999000000, time.UTC)
	if got := TripDaysInclusive(from, to); got != 2 {
		t.Fatalf("May 15–16 (end 23:59) = 2 days, got %d", got)
	}
}

func TestAgreedRentalBaseForBooking(t *testing.T) {
	perDay := decimal.RequireFromString("1000")
	b := &models.Booking{
		FinalBookingPrice: &perDay,
		RentalFrom:        time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC),
		RentalTo:          time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
	}
	base := AgreedRentalBaseForBooking(b)
	want := decimal.RequireFromString("2000")
	if !base.Equal(want) {
		t.Fatalf("expected %s, got %s", want, base)
	}
}

func TestBuildPaymentBreakdownOnTripBase(t *testing.T) {
	base := decimal.RequireFromString("2000")
	bd := BuildPaymentBreakdown(base, 2, 1.5, 18)
	if bd.OwnerCommissionAmount.StringFixed(2) != "30.00" {
		t.Fatalf("owner fee: %s", bd.OwnerCommissionAmount)
	}
	if bd.OwnerGSTAmount.StringFixed(2) != "360.00" {
		t.Fatalf("owner gst: %s", bd.OwnerGSTAmount)
	}
	if bd.OwnerNet.StringFixed(2) != "1610.00" {
		t.Fatalf("owner net: %s", bd.OwnerNet)
	}
}
