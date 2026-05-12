package repository

import (
	"context"

	"carmanage/backend/internal/models"
)

func (d *DB) CreateBookingReview(ctx context.Context, r *models.BookingReview) error {
	return d.WithContext(ctx).Create(r).Error
}
