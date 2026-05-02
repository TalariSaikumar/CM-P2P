package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// AirbagDetail groups airbags by type (e.g. front, side curtain) with a count.
type AirbagDetail struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// Car is an owner-listed vehicle with required identity and tiered pricing.
type Car struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	OwnerID uuid.UUID `gorm:"type:uuid;not null;index"`

	CarName            string `gorm:"size:255;not null"`
	CarModel           string `gorm:"size:255;not null;index"`
	CarNumber          string `gorm:"size:32;not null;uniqueIndex"` // plate
	RegistrationNumber string `gorm:"size:64;not null"`             // RC
	EngineNumber       string `gorm:"size:64;not null"`

	PricePerHour decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	PricePerDay  decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	PricePerKm   decimal.Decimal `gorm:"type:numeric(12,2);not null"`

	// Location supports customer search by area/city.
	Location string `gorm:"size:255;not null;index"`

	// Defaults satisfy AutoMigrate when adding NOT NULL columns to existing rows (SQL migration 000008 does the same).
	ModelYear       int    `gorm:"not null;default:2020"`
	Color           string `gorm:"size:64;not null;default:''"`
	FuelType        string `gorm:"size:16;not null;default:petrol"` // petrol, diesel, cng, ev
	Transmission    string `gorm:"size:16;not null;default:manual"` // auto, manual
	MileageKm       int    `gorm:"not null;default:0"`
	NumSeats        int    `gorm:"not null;default:5"`
	AirbagCount     int    `gorm:"not null;default:0"`
	AirbagDetails   []byte `gorm:"type:jsonb"` // nullable for legacy AutoMigrate rows; service sets JSON on create/update
	CameraType      string `gorm:"size:64;not null;default:''"`
	AirConditioning bool   `gorm:"not null;default:false"`
	Airbags         bool   `gorm:"not null;default:false"`
	CruiseControl   bool   `gorm:"not null;default:false"`
	OpenRoof        bool   `gorm:"not null;default:false"`
	Navigation      bool   `gorm:"not null;default:false"`
	Speakers        bool   `gorm:"not null;default:false"`

	IsActive bool `gorm:"not null;default:true;index"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Owner  User       `gorm:"foreignKey:OwnerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Images []CarImage `gorm:"foreignKey:CarID"`
}

func (c *Car) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// CarImage references car media in Azure Blob Storage.
type CarImage struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	CarID uuid.UUID `gorm:"type:uuid;not null;index"`

	BlobPath  string `gorm:"size:512;not null"`
	BlobURL   string `gorm:"size:1024;not null"`
	SortOrder int    `gorm:"not null;default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Car Car `gorm:"foreignKey:CarID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ci *CarImage) BeforeCreate(tx *gorm.DB) error {
	if ci.ID == uuid.Nil {
		ci.ID = uuid.New()
	}
	return nil
}
