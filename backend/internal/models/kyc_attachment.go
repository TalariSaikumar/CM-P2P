package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KYCAttachmentKind categorizes uploaded verification documents in blob storage.
type KYCAttachmentKind string

const (
	KYCKindAadhaarFront KYCAttachmentKind = "AADHAAR_FRONT"
	KYCKindAadhaarBack  KYCAttachmentKind = "AADHAAR_BACK"
	KYCKindDrivingLicense KYCAttachmentKind = "DRIVING_LICENSE"
	KYCKindAddressProof   KYCAttachmentKind = "ADDRESS_PROOF"
	KYCKindOther          KYCAttachmentKind = "OTHER"
)

// KYCAttachment stores metadata for a document stored in Azure Blob Storage.
type KYCAttachment struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index"`

	Kind     KYCAttachmentKind `gorm:"type:varchar(32);not null"`
	BlobPath string            `gorm:"size:512;not null"` // container-relative path or full URL depending on app convention
	BlobURL  string            `gorm:"size:1024;not null"` // public or SAS URL for display/download

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (k *KYCAttachment) BeforeCreate(tx *gorm.DB) error {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	return nil
}
