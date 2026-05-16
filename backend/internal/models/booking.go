package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// BookingStatus drives customer/owner dashboards and negotiation flow.
type BookingStatus string

const (
	BookingPending     BookingStatus = "PENDING"
	BookingNegotiating BookingStatus = "NEGOTIATING"
	BookingConfirmed   BookingStatus = "CONFIRMED"
	BookingCompleted   BookingStatus = "COMPLETED"
	BookingCancelled   BookingStatus = "CANCELLED"
)

// Booking payment lifecycle (simulated gateway): 75% deposit, then post-trip charges, then final balance.
const (
	BookingPaymentUnpaid      = "UNPAID"
	BookingPaymentDepositPaid = "DEPOSIT_PAID"
	BookingPaymentFinalDue    = "FINAL_DUE"
	BookingPaymentPaid        = "PAID"
)

// Booking ties a customer inquiry to a car; owner sets FinalBookingPrice via PATCH.
type Booking struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	CarID      uuid.UUID `gorm:"type:uuid;not null;index"`
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index"`
	OwnerID    uuid.UUID `gorm:"type:uuid;not null;index"`

	Status BookingStatus `gorm:"type:varchar(24);not null;default:PENDING;index"`

	// FinalBookingPrice is the owner-negotiated rate per calendar day (nullable until agreed).
	FinalBookingPrice *decimal.Decimal `gorm:"type:numeric(14,2)"`

	// Customer accepts the owner's quoted price (cleared when owner changes the price).
	CustomerAcceptedPriceAt     *time.Time       `gorm:"type:timestamptz"`
	CustomerAcceptedPriceAmount *decimal.Decimal `gorm:"type:numeric(14,2)"`

	// CustomerNote optional text from the initial booking inquiry.
	CustomerNote string `gorm:"type:text"`

	// Rental window and handover points (required on create for new rows).
	// DB defaults satisfy AutoMigrate for existing rows before backfill; API still validates on create.
	RentalFrom  time.Time `gorm:"type:timestamptz;not null;default:now()"`
	RentalTo    time.Time `gorm:"type:timestamptz;not null;default:now()"`
	PickupPoint string    `gorm:"type:text;not null;default:''"`
	DropPoint   string    `gorm:"type:text;not null;default:''"`

	PaymentStatus            string           `gorm:"size:16;not null;default:UNPAID;index"`
	PaymentMethod            string           `gorm:"size:24"`
	PaidAt                   *time.Time       `gorm:"type:timestamptz"` // set when the booking is fully settled (final payment)
	DepositPaidAt            *time.Time       `gorm:"type:timestamptz"`
	DepositCustomerTotal     *decimal.Decimal `gorm:"type:numeric(14,2)"`
	PostTripChargesTotal     decimal.Decimal  `gorm:"type:numeric(14,2);not null;default:0"`
	SettlementSubmittedAt    *time.Time       `gorm:"type:timestamptz"`
	CustomerAcknowledgedTermsAt *time.Time    `gorm:"type:timestamptz"`
	CustomerCommissionRate   *decimal.Decimal `gorm:"type:numeric(6,3)"`
	OwnerCommissionRate      *decimal.Decimal `gorm:"type:numeric(6,3)"`
	CustomerCommissionAmount *decimal.Decimal `gorm:"type:numeric(14,2)"`
	OwnerCommissionAmount    *decimal.Decimal `gorm:"type:numeric(14,2)"`
	CustomerTotalPaid        *decimal.Decimal `gorm:"type:numeric(14,2)"`
	OwnerNetPayout           *decimal.Decimal `gorm:"type:numeric(14,2)"`

	GstPercentOnCommission *decimal.Decimal `gorm:"type:numeric(5,2)"`
	CustomerGSTAmount      *decimal.Decimal `gorm:"type:numeric(14,2)"`
	OwnerGSTAmount         *decimal.Decimal `gorm:"type:numeric(14,2)"`

	CancellationReason string     `gorm:"type:text"`
	CancelledAt        *time.Time `gorm:"type:timestamptz"`
	CancelledByUserID  *uuid.UUID `gorm:"type:uuid"`

	// Owner records vehicle handover when keys are given (before customer drives).
	OwnerPickupOdometerKM    *int       `gorm:"type:integer"`
	OwnerPickupFuelPercent   *int       `gorm:"type:smallint"`
	OwnerPickupHandoverNotes string     `gorm:"type:text"`
	OwnerPickupHandoverAt    *time.Time `gorm:"type:timestamptz"`

	// Customer pickup check-in after accepting the owner's handover record.
	CustomerPickupAcceptedAt *time.Time `gorm:"type:timestamptz"`
	PickupOdometerKM         *int       `gorm:"type:integer"`
	PickupFuelPercent        *int       `gorm:"type:smallint"`
	PickupHandoverNotes      string     `gorm:"type:text"`
	PickupHandoverAt         *time.Time `gorm:"type:timestamptz"`
	// Customer return check-in when handing the vehicle back.
	ReturnOdometerKM    *int       `gorm:"type:integer"`
	ReturnFuelPercent   *int       `gorm:"type:smallint"`
	ReturnHandoverNotes string     `gorm:"type:text"`
	ReturnHandoverAt    *time.Time `gorm:"type:timestamptz"`

	// Owner confirms return after the customer has paid the final balance.
	OwnerReturnAcceptedAt *time.Time `gorm:"type:timestamptz"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Car      Car  `gorm:"foreignKey:CarID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Customer User `gorm:"foreignKey:CustomerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Owner    User `gorm:"foreignKey:OwnerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Messages          []Message
	Reviews           []BookingReview           `gorm:"foreignKey:BookingID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	PostTripCharges   []BookingPostTripCharge `gorm:"foreignKey:BookingID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	HandoverPhotos    []BookingHandoverPhoto  `gorm:"foreignKey:BookingID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (b *Booking) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	if b.PaymentStatus == "" {
		b.PaymentStatus = BookingPaymentUnpaid
	}
	return nil
}
