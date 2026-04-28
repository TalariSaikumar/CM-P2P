package repository

import (
	"context"

	"carmanage/backend/internal/models"

	"github.com/google/uuid"
)

// GetUserByID loads a user by primary key.
func (d *DB) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	if err := d.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByEmail loads a user by email (case-insensitive).
func (d *DB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	if err := d.WithContext(ctx).Where("LOWER(email) = LOWER(?)", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// CreateUser persists a new user.
func (d *DB) CreateUser(ctx context.Context, u *models.User) error {
	return d.WithContext(ctx).Create(u).Error
}

// UpdateUser saves mutable profile fields and verification flag.
func (d *DB) UpdateUser(ctx context.Context, u *models.User) error {
	return d.WithContext(ctx).Save(u).Error
}

// CreateKYCAttachment persists a KYC document row.
func (d *DB) CreateKYCAttachment(ctx context.Context, a *models.KYCAttachment) error {
	return d.WithContext(ctx).Create(a).Error
}

// ListKYCAttachments returns attachments for a user.
func (d *DB) ListKYCAttachments(ctx context.Context, userID uuid.UUID) ([]models.KYCAttachment, error) {
	var rows []models.KYCAttachment
	if err := d.WithContext(ctx).Where("user_id = ?", userID).Order("created_at asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
