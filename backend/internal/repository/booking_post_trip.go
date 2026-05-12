package repository

import (
	"context"
	"time"

	"carmanage/backend/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SavePostTripSettlement replaces charge rows and updates scalar settlement fields on the booking.
func (d *DB) SavePostTripSettlement(ctx context.Context, bookingID uuid.UUID, rows []models.BookingPostTripCharge, postTripTotal decimal.Decimal, paymentStatus string, settlementAt time.Time) error {
	return d.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("booking_id = ?", bookingID).Delete(&models.BookingPostTripCharge{}).Error; err != nil {
			return err
		}
		for i := range rows {
			r := rows[i]
			r.BookingID = bookingID
			if err := tx.Create(&r).Error; err != nil {
				return err
			}
		}
		updates := map[string]interface{}{
			"post_trip_charges_total":   postTripTotal,
			"payment_status":            paymentStatus,
			"settlement_submitted_at":   settlementAt,
			"updated_at":                time.Now().UTC(),
		}
		return tx.Model(&models.Booking{}).Where("id = ?", bookingID).Updates(updates).Error
	})
}
