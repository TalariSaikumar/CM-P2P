package repository

import (
	"context"

	"carmanage/backend/internal/models"

	"github.com/google/uuid"
)

// CreateBookingHandoverPhoto persists handover photo metadata.
func (d *DB) CreateBookingHandoverPhoto(ctx context.Context, p *models.BookingHandoverPhoto) error {
	return d.WithContext(ctx).Create(p).Error
}

// CountHandoverPhotosForStep returns how many photos exist for a booking step.
func (d *DB) CountHandoverPhotosForStep(ctx context.Context, bookingID uuid.UUID, step models.HandoverPhotoStep) (int64, error) {
	var n int64
	err := d.WithContext(ctx).Model(&models.BookingHandoverPhoto{}).
		Where("booking_id = ? AND step = ?", bookingID, step).
		Count(&n).Error
	return n, err
}
