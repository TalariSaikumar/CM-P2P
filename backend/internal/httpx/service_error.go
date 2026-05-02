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
)

// WrapValidation returns a validation error with a specific message.
func WrapValidation(msg string) *ServiceError {
	return NewError(http.StatusBadRequest, "VALIDATION_ERROR", msg)
}

// Wrapf formats a generic bad request.
func Wrapf(format string, args ...any) *ServiceError {
	return NewError(http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf(format, args...))
}
