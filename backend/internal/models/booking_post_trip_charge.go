package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// BookingPostTripCharge is an owner-declared line (tolls, damage, fines, etc.) added after the trip.
type BookingPostTripCharge struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey"`
	BookingID uuid.UUID       `gorm:"type:uuid;not null;index"`
	Label     string          `gorm:"type:text;not null"`
	AmountInr decimal.Decimal `gorm:"type:numeric(14,2);not null"`
	CreatedAt time.Time       `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (BookingPostTripCharge) TableName() string {
	return "booking_post_trip_charges"
}

func (c *BookingPostTripCharge) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
