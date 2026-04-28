package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole distinguishes platform participants.
type UserRole string

const (
	RoleCustomer UserRole = "CUSTOMER"
	RoleOwner    UserRole = "OWNER"
)

// User is the core identity with dual-side KYC fields.
// CUSTOMER: must supply DrivingLicenseNumber for booking (enforced in services/middleware).
// OWNER: must be KYC verified before listing vehicles.
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     string    `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string `gorm:"size:255;not null"`
	Role      UserRole  `gorm:"type:varchar(20);not null;index"`

	FullName      string `gorm:"size:255;not null"`
	AadhaarNumber string `gorm:"size:32;not null"`
	PhoneNumber   string `gorm:"size:32;not null;index"`
	Address       string `gorm:"type:text;not null"`

	// DrivingLicenseNumber is required for customers at booking time (nullable in DB until provided).
	DrivingLicenseNumber *string `gorm:"size:64"`

	IsKYCVerified bool `gorm:"not null;default:false;index"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
