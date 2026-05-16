package httpx

import (
	"fmt"
	"net/http"
)

// ServiceError is a domain error mapped to HTTP responses.
type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

// NewError builds a ServiceError with a status and stable code for clients.
func NewError(status int, code, message string) *ServiceError {
	return &ServiceError{Status: status, Code: code, Message: message}
}

// Common constructors
var (
	ErrUnauthorized     = NewError(http.StatusUnauthorized, "UNAUTHORIZED", "You need to sign in to continue.")
	ErrForbidden        = NewError(http.StatusForbidden, "FORBIDDEN", "You do not have access to this resource.")
	ErrNotFound         = NewError(http.StatusNotFound, "NOT_FOUND", "We could not find what you were looking for.")
	ErrConflict         = NewError(http.StatusConflict, "CONFLICT", "This action conflicts with the current state.")
	ErrValidation       = NewError(http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
	ErrStorage          = NewError(http.StatusServiceUnavailable, "STORAGE_UNAVAILABLE", "File storage is not available right now. Please try again later.")
	ErrKYCRequired      = NewError(http.StatusForbidden, "KYC_REQUIRED", "Complete identity verification before using this feature.")
	ErrDrivingLicense   = NewError(http.StatusForbidden, "DRIVING_LICENSE_REQUIRED", "Add your verified driving license number to your profile before booking.")
	ErrCarAlreadyBooked = NewError(http.StatusConflict, "CAR_ALREADY_BOOKED", "Already booked for this period. Try different dates or another vehicle.")
	ErrCarBookedToday   = NewError(http.StatusConflict, "CAR_BOOKED_TODAY", "This vehicle is booked for today’s date (UTC). Details cannot be changed until that rental window no longer includes today.")
	ErrPaymentNotReady  = NewError(http.StatusBadRequest, "PAYMENT_NOT_READY", "Payment is available only after the owner confirms the booking and an agreed price is set.")
	ErrPaymentAlreadyPaid = NewError(http.StatusConflict, "ALREADY_PAID", "This booking is already marked as paid.")
	ErrInvalidPaymentMethod = NewError(http.StatusBadRequest, "INVALID_PAYMENT_METHOD", "Choose a valid payment method: UPI, CARD, NET_BANKING, or WALLET.")
	ErrBookingPaidNoCancel  = NewError(http.StatusConflict, "BOOKING_PAID", "Bookings with a deposit or full payment cannot be cancelled in the app. Contact support if you need to change or refund a trip.")
	ErrReviewExists           = NewError(http.StatusConflict, "REVIEW_EXISTS", "You have already submitted a review for this booking.")
	ErrHandoverExists         = NewError(http.StatusConflict, "HANDOVER_EXISTS", "That pickup or return handover was already recorded.")
	ErrBookingTermsNotAck     = NewError(http.StatusBadRequest, "BOOKING_TERMS", "You must acknowledge the deposit and post-trip billing terms before creating a booking.")
	ErrSettlementNotReady     = NewError(http.StatusBadRequest, "SETTLEMENT_NOT_READY", "The owner has not yet submitted post-trip charges. You can pay the final balance only after they do.")
	ErrSettlementTooEarly     = NewError(http.StatusBadRequest, "SETTLEMENT_TOO_EARLY", "Post-trip charges can be submitted once the customer has recorded return check-in.")
	ErrDepositRequiredForSettlement = NewError(http.StatusBadRequest, "DEPOSIT_REQUIRED", "The customer must pay the trip deposit before you can submit post-trip charges.")
	ErrSettlementLockedPaid   = NewError(http.StatusConflict, "SETTLEMENT_LOCKED", "This booking is already fully paid; post-trip charges cannot be changed.")
	ErrReturnHandoverRequired   = NewError(http.StatusBadRequest, "RETURN_HANDOVER_REQUIRED", "Record return handover before submitting post-trip charges.")
	ErrDepositRequiredForHandover = NewError(http.StatusBadRequest, "DEPOSIT_REQUIRED", "Handover opens after the customer pays the trip deposit.")
	ErrOwnerHandoverRequired    = NewError(http.StatusBadRequest, "OWNER_HANDOVER_REQUIRED", "The owner must record vehicle handover before you can complete pickup check-in.")
	ErrOwnerHandoverExists      = NewError(http.StatusConflict, "OWNER_HANDOVER_EXISTS", "Owner vehicle handover was already recorded.")
	ErrCustomerPickupExists     = NewError(http.StatusConflict, "CUSTOMER_PICKUP_EXISTS", "Customer pickup check-in was already completed.")
	ErrOwnerOnlyHandover        = NewError(http.StatusForbidden, "OWNER_ONLY", "Only the owner can record vehicle handover at pickup.")
	ErrCustomerOnlyPickup       = NewError(http.StatusForbidden, "CUSTOMER_ONLY", "Only the customer can complete pickup check-in.")
	ErrCustomerReturnRequired   = NewError(http.StatusBadRequest, "CUSTOMER_RETURN_REQUIRED", "The customer must record return check-in before you can continue.")
	ErrCustomerReturnExists     = NewError(http.StatusConflict, "CUSTOMER_RETURN_EXISTS", "Customer return check-in was already recorded.")
	ErrFinalPaymentRequiredForReturnAccept = NewError(http.StatusBadRequest, "FINAL_PAYMENT_REQUIRED", "The customer must pay the final balance before you can accept the return.")
	ErrOwnerReturnAlreadyAccepted = NewError(http.StatusConflict, "OWNER_RETURN_ACCEPTED", "Return was already accepted for this booking.")
	ErrCustomerPriceNotAccepted = NewError(http.StatusConflict, "CUSTOMER_PRICE_NOT_ACCEPTED", "The customer must accept the quoted price before you can confirm this booking.")
	ErrPriceNotQuoted           = NewError(http.StatusBadRequest, "PRICE_NOT_QUOTED", "The owner has not set a quoted price to accept yet.")
)

// WrapValidation returns a validation error with a specific message.
func WrapValidation(msg string) *ServiceError {
	return NewError(http.StatusBadRequest, "VALIDATION_ERROR", msg)
}

// Wrapf formats a generic bad request.
func Wrapf(format string, args ...any) *ServiceError {
	return NewError(http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf(format, args...))
}
