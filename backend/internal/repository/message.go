package repository

import (
	"context"

	"carmanage/backend/internal/models"

	"github.com/google/uuid"
)

// ListMessagesForBooking returns chat messages oldest-first.
func (d *DB) ListMessagesForBooking(ctx context.Context, bookingID uuid.UUID) ([]models.Message, error) {
	var rows []models.Message
	if err := d.WithContext(ctx).Where("booking_id = ?", bookingID).Preload("Sender").Order("created_at asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// CreateMessage persists a chat line.
func (d *DB) CreateMessage(ctx context.Context, m *models.Message) error {
	return d.WithContext(ctx).Create(m).Error
}

// CountMessages returns how many messages exist for a booking.
func (d *DB) CountMessages(ctx context.Context, bookingID uuid.UUID) (int64, error) {
	var n int64
	if err := d.WithContext(ctx).Model(&models.Message{}).Where("booking_id = ?", bookingID).Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}
