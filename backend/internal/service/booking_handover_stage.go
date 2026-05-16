package service

import (
	"time"

	"carmanage/backend/internal/models"
)

// Trip stage codes for handover / rental lifecycle UI.
const (
	TripStageHidden                      = "hidden"
	TripStageAwaitingOwnerHandover       = "awaiting_owner_handover"
	TripStageAwaitingCustomerPickup      = "awaiting_customer_pickup"
	TripStageOnTrip                      = "on_trip"
	TripStageAwaitingCustomerReturn      = "awaiting_customer_return"
	TripStageAwaitingPostTripCharges     = "awaiting_post_trip_charges"
	TripStageAwaitingFinalPayment        = "awaiting_final_payment"
	TripStageAwaitingOwnerReturnAccept   = "awaiting_owner_return_acceptance"
	TripStageReturnComplete              = "return_complete"
)

// TripStageInfo is exposed on booking JSON for drawer titles and hints.
type TripStageInfo struct {
	Code  string
	Label string
}

// DepositPaidForTrip returns true once the customer has paid the trip deposit.
func DepositPaidForTrip(b *models.Booking) bool {
	switch b.PaymentStatus {
	case models.BookingPaymentDepositPaid, models.BookingPaymentFinalDue, models.BookingPaymentPaid:
		return true
	default:
		return false
	}
}

// OwnerPickupRecorded is true after the owner logs handing the vehicle to the customer.
func OwnerPickupRecorded(b *models.Booking) bool {
	return b.OwnerPickupHandoverAt != nil
}

// CustomerPickupComplete is true after the customer accepts and records their pickup check-in.
func CustomerPickupComplete(b *models.Booking) bool {
	return b.CustomerPickupAcceptedAt != nil && b.PickupHandoverAt != nil
}

// CustomerReturnRecorded is true after the customer logs returning the vehicle.
func CustomerReturnRecorded(b *models.Booking) bool {
	return b.ReturnHandoverAt != nil
}

// OwnerReturnAccepted is true after the owner accepts the return (final payment must be complete).
func OwnerReturnAccepted(b *models.Booking) bool {
	return b.OwnerReturnAcceptedAt != nil
}

// ReturnComplete is true when customer return is logged and the owner has accepted it.
func ReturnComplete(b *models.Booking) bool {
	return CustomerReturnRecorded(b) && OwnerReturnAccepted(b)
}

// EffectivePickupOdometerKM prefers the customer reading, then owner, for return validation.
func EffectivePickupOdometerKM(b *models.Booking) *int {
	if b.PickupOdometerKM != nil {
		return b.PickupOdometerKM
	}
	return b.OwnerPickupOdometerKM
}

// TripStageForBooking computes the current rental handover stage.
func TripStageForBooking(b *models.Booking) TripStageInfo {
	if b.Status == models.BookingCompleted {
		return TripStageInfo{Code: TripStageReturnComplete, Label: "Trip complete"}
	}
	if b.Status != models.BookingConfirmed || !DepositPaidForTrip(b) {
		return TripStageInfo{Code: TripStageHidden, Label: ""}
	}
	if !OwnerPickupRecorded(b) {
		return TripStageInfo{Code: TripStageAwaitingOwnerHandover, Label: "Awaiting vehicle handover"}
	}
	if !CustomerPickupComplete(b) {
		return TripStageInfo{Code: TripStageAwaitingCustomerPickup, Label: "Pickup verification"}
	}
	if OwnerReturnAccepted(b) {
		return TripStageInfo{Code: TripStageReturnComplete, Label: "Trip complete"}
	}
	if !CustomerReturnRecorded(b) {
		now := time.Now().UTC()
		if now.After(b.RentalTo.UTC()) {
			return TripStageInfo{Code: TripStageAwaitingCustomerReturn, Label: "Return check-in"}
		}
		return TripStageInfo{Code: TripStageOnTrip, Label: "Trip in progress"}
	}
	if b.PaymentStatus == models.BookingPaymentDepositPaid {
		return TripStageInfo{Code: TripStageAwaitingPostTripCharges, Label: "Add post-trip charges"}
	}
	if b.PaymentStatus == models.BookingPaymentFinalDue {
		return TripStageInfo{Code: TripStageAwaitingFinalPayment, Label: "Awaiting final payment"}
	}
	if b.PaymentStatus != models.BookingPaymentPaid {
		return TripStageInfo{Code: TripStageAwaitingFinalPayment, Label: "Awaiting final payment"}
	}
	return TripStageInfo{Code: TripStageAwaitingOwnerReturnAccept, Label: "Return verification"}
}
