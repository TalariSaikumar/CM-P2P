package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HandoverPhotoStep groups optional condition photos by handover phase.
type HandoverPhotoStep string

const (
	HandoverPhotoOwnerPickup    HandoverPhotoStep = "owner_pickup"
	HandoverPhotoCustomerPickup HandoverPhotoStep = "customer_pickup"
	HandoverPhotoCustomerReturn HandoverPhotoStep = "customer_return"
	HandoverPhotoOwnerReturn    HandoverPhotoStep = "owner_return"
)

// BookingHandoverPhoto stores metadata for an optional handover image in blob storage.
type BookingHandoverPhoto struct {
	ID         uuid.UUID         `gorm:"type:uuid;primaryKey"`
	BookingID  uuid.UUID         `gorm:"type:uuid;not null;index:idx_booking_handover_photos_booking_step,priority:1"`
	UploaderID uuid.UUID         `gorm:"type:uuid;not null"`
	Step       HandoverPhotoStep `gorm:"type:varchar(32);not null;index:idx_booking_handover_photos_booking_step,priority:2"`
	BlobPath   string            `gorm:"size:512;not null"`
	BlobURL    string            `gorm:"size:1024;not null"`
	CreatedAt  time.Time

	Booking  Booking `gorm:"foreignKey:BookingID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Uploader User    `gorm:"foreignKey:UploaderID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (p *BookingHandoverPhoto) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
