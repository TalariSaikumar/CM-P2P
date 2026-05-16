package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"carmanage/backend/internal/httpx"
	"carmanage/backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxHandoverPhotosPerStep = 10

var allowedHandoverImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// UploadHandoverPhoto stores an optional condition photo for a handover step.
func (s *BookingService) UploadHandoverPhoto(ctx context.Context, userID, bookingID uuid.UUID, step models.HandoverPhotoStep, filename, contentType string, body io.Reader) (*models.BookingHandoverPhoto, error) {
	if s.Blob == nil {
		return nil, httpx.ErrStorage
	}
	if body == nil {
		return nil, httpx.WrapValidation("Image file is required.")
	}
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if ct == "" {
		ct = "application/octet-stream"
	}
	if !allowedHandoverImageTypes[ct] {
		return nil, httpx.WrapValidation("Only JPEG, PNG, or WebP images are allowed.")
	}

	b, err := s.Repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	if err := validateHandoverPhotoUpload(b, userID, step); err != nil {
		return nil, err
	}

	n, err := s.Repo.CountHandoverPhotosForStep(ctx, bookingID, step)
	if err != nil {
		return nil, err
	}
	if n >= maxHandoverPhotosPerStep {
		return nil, httpx.WrapValidation(fmt.Sprintf("At most %d photos per handover step.", maxHandoverPhotosPerStep))
	}

	safe := strings.ReplaceAll(strings.ToLower(filename), "..", "")
	if safe == "" {
		safe = "photo"
	}
	blobPath := fmt.Sprintf("bookings/%s/handover/%s/%s-%d", bookingID.String(), step, safe, time.Now().UnixNano())
	url, err := s.Blob.Upload(ctx, blobPath, body, ct)
	if err != nil {
		return nil, fmt.Errorf("upload handover photo: %w", err)
	}

	p := &models.BookingHandoverPhoto{
		BookingID:  bookingID,
		UploaderID: userID,
		Step:       step,
		BlobPath:   blobPath,
		BlobURL:    url,
	}
	if err := s.Repo.CreateBookingHandoverPhoto(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func validateHandoverPhotoUpload(b *models.Booking, userID uuid.UUID, step models.HandoverPhotoStep) error {
	if b.CustomerID != userID && b.OwnerID != userID {
		return httpx.ErrForbidden
	}
	if b.Status != models.BookingConfirmed {
		return httpx.NewError(409, "BOOKING_STATE", "Photos can only be added on confirmed bookings.")
	}
	if !DepositPaidForTrip(b) {
		return httpx.ErrDepositRequiredForHandover
	}
	isOwner := b.OwnerID == userID
	isCustomer := b.CustomerID == userID

	switch step {
	case models.HandoverPhotoOwnerPickup:
		if !isOwner {
			return httpx.ErrOwnerOnlyHandover
		}
		if CustomerPickupComplete(b) {
			return httpx.NewError(409, "HANDOVER_PHASE_CLOSED", "Pickup phase is closed; photos cannot be added for owner handover.")
		}
	case models.HandoverPhotoCustomerPickup:
		if !isCustomer {
			return httpx.ErrCustomerOnlyPickup
		}
		if !OwnerPickupRecorded(b) {
			return httpx.ErrOwnerHandoverRequired
		}
		if CustomerReturnRecorded(b) {
			return httpx.NewError(409, "HANDOVER_PHASE_CLOSED", "Pickup phase is closed.")
		}
	case models.HandoverPhotoCustomerReturn:
		if !isCustomer {
			return httpx.ErrCustomerOnlyPickup
		}
		if !CustomerPickupComplete(b) {
			return httpx.NewError(400, "PICKUP_INCOMPLETE", "Complete pickup check-in before return photos.")
		}
		if OwnerReturnAccepted(b) {
			return httpx.NewError(409, "HANDOVER_PHASE_CLOSED", "Return phase is closed.")
		}
	case models.HandoverPhotoOwnerReturn:
		if !isOwner {
			return httpx.ErrForbidden
		}
		if !CustomerReturnRecorded(b) {
			return httpx.ErrCustomerReturnRequired
		}
		if b.PaymentStatus != models.BookingPaymentPaid {
			return httpx.ErrFinalPaymentRequiredForReturnAccept
		}
		if OwnerReturnAccepted(b) {
			return httpx.ErrOwnerReturnAlreadyAccepted
		}
	default:
		return httpx.WrapValidation("Unknown handover photo step.")
	}
	return nil
}

// ParseHandoverPhotoStep validates the step query/form value.
func ParseHandoverPhotoStep(raw string) (models.HandoverPhotoStep, error) {
	switch models.HandoverPhotoStep(strings.TrimSpace(raw)) {
	case models.HandoverPhotoOwnerPickup, models.HandoverPhotoCustomerPickup,
		models.HandoverPhotoCustomerReturn, models.HandoverPhotoOwnerReturn:
		return models.HandoverPhotoStep(strings.TrimSpace(raw)), nil
	default:
		return "", httpx.WrapValidation(`Step must be one of: owner_pickup, customer_pickup, customer_return, owner_return.`)
	}
}
