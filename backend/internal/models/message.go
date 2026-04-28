package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Message is a REST-persisted chat line between owner and customer for a booking.
type Message struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	BookingID uuid.UUID `gorm:"type:uuid;not null;index"`
	SenderID  uuid.UUID `gorm:"type:uuid;not null;index"`

	Body string `gorm:"type:text;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Booking Booking `gorm:"foreignKey:BookingID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Sender  User    `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
