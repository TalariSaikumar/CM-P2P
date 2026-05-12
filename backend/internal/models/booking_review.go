package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BookingReview is one party's rating for a completed paid rental (unique per booking + party).
type BookingReview struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	BookingID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_booking_review_party"`
	ReviewerParty string    `gorm:"size:16;not null;uniqueIndex:idx_booking_review_party"` // CUSTOMER or OWNER (who wrote it)
	ReviewerID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Rating        int       `gorm:"not null"`
	Comment       string    `gorm:"type:text;not null;default:''"`
	CreatedAt     time.Time

	Reviewer User `gorm:"foreignKey:ReviewerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (BookingReview) TableName() string {
	return "booking_reviews"
}

func (r *BookingReview) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
