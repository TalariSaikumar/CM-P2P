package service

import (
	"time"

	"carmanage/backend/internal/models"

	"github.com/shopspring/decimal"
)

// TripDaysInclusive counts inclusive UTC calendar days from rental_from through rental_to (minimum 1).
func TripDaysInclusive(rentalFrom, rentalTo time.Time) int {
	rf := rentalFrom.UTC()
	rt := rentalTo.UTC()
	y1, m1, d1 := rf.Date()
	y2, m2, d2 := rt.Date()
	start := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
	end := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)
	if end.Before(start) {
		return 1
	}
	days := int(end.Sub(start) / (24 * time.Hour))
	if days < 0 {
		return 1
	}
	return days + 1
}

// AgreedRentalBaseForBooking is negotiated per-day rate × inclusive trip days (basis for fees and GST).
func AgreedRentalBaseForBooking(b *models.Booking) decimal.Decimal {
	if b.FinalBookingPrice == nil {
		return decimal.Zero
	}
	days := TripDaysInclusive(b.RentalFrom, b.RentalTo)
	return b.FinalBookingPrice.Mul(decimal.NewFromInt(int64(days))).Round(2)
}
