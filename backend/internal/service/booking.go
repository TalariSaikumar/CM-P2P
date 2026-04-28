package service

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"carmanage/backend/internal/httpx"
	"carmanage/backend/internal/models"
	"carmanage/backend/internal/notify"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// BookingService implements inquiry, negotiation, and confirmation.
type BookingService struct {
	Deps
}

type CreateBookingInput struct {
	CarID        uuid.UUID
	CustomerNote string
}

func (s *BookingService) Create(ctx context.Context, customerID uuid.UUID, in CreateBookingInput) (*models.Booking, error) {
	u, err := s.Repo.GetUserByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	if u.Role != models.RoleCustomer {
		return nil, httpx.NewError(403, "CUSTOMER_ONLY", "Only customers can create booking inquiries.")
	}
	if !u.IsKYCVerified {
		return nil, httpx.ErrKYCRequired
	}
	if u.DrivingLicenseNumber == nil || strings.TrimSpace(*u.DrivingLicenseNumber) == "" {
		return nil, httpx.ErrDrivingLicense
	}

	car, err := s.Repo.GetCarByID(ctx, in.CarID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	if !car.IsActive {
		return nil, httpx.NewError(400, "CAR_INACTIVE", "This listing is not available for booking.")
	}
	if car.OwnerID == customerID {
		return nil, httpx.WrapValidation("You cannot book your own vehicle.")
	}

	b := &models.Booking{
		CarID:        car.ID,
		CustomerID:   customerID,
		OwnerID:      car.OwnerID,
		Status:       models.BookingPending,
		CustomerNote: strings.TrimSpace(in.CustomerNote),
	}
	if err := s.Repo.CreateBooking(ctx, b); err != nil {
		return nil, err
	}
	return s.Repo.GetBookingByID(ctx, b.ID)
}

func (s *BookingService) Get(ctx context.Context, userID, bookingID uuid.UUID) (*models.Booking, error) {
	b, err := s.Repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	if b.CustomerID != userID && b.OwnerID != userID {
		return nil, httpx.ErrForbidden
	}
	return b, nil
}

func (s *BookingService) ListForCustomer(ctx context.Context, customerID uuid.UUID) ([]models.Booking, error) {
	return s.Repo.ListBookingsForCustomer(ctx, customerID)
}

func (s *BookingService) ListForOwner(ctx context.Context, ownerID uuid.UUID) ([]models.Booking, error) {
	return s.Repo.ListBookingsForOwner(ctx, ownerID)
}

type PatchFinalPriceInput struct {
	FinalBookingPrice string
}

func (s *BookingService) PatchFinalPrice(ctx context.Context, ownerID, bookingID uuid.UUID, in PatchFinalPriceInput) (*models.Booking, error) {
	price, err := decimal.NewFromString(strings.TrimSpace(in.FinalBookingPrice))
	if err != nil {
		return nil, httpx.WrapValidation("Final booking price must be a valid decimal number.")
	}
	if price.IsNegative() {
		return nil, httpx.WrapValidation("Final booking price cannot be negative.")
	}

	b, err := s.Repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	if b.OwnerID != ownerID {
		return nil, httpx.ErrForbidden
	}
	if b.Status == models.BookingConfirmed || b.Status == models.BookingCancelled {
		return nil, httpx.NewError(409, "BOOKING_CLOSED", "This booking can no longer be updated.")
	}

	b.FinalBookingPrice = &price
	if err := s.Repo.UpdateBooking(ctx, b); err != nil {
		return nil, err
	}
	return s.Repo.GetBookingByID(ctx, bookingID)
}

func (s *BookingService) Confirm(ctx context.Context, customerID, bookingID uuid.UUID) (*models.Booking, error) {
	b, err := s.Repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	if b.CustomerID != customerID {
		return nil, httpx.ErrForbidden
	}
	if b.Status == models.BookingConfirmed {
		return b, nil
	}
	if b.Status == models.BookingCancelled {
		return nil, httpx.NewError(409, "BOOKING_CANCELLED", "This booking is no longer active.")
	}
	if b.FinalBookingPrice == nil {
		return nil, httpx.NewError(400, "PRICE_NOT_SET", "The owner has not set a final price yet.")
	}

	b.Status = models.BookingConfirmed
	if err := s.Repo.UpdateBooking(ctx, b); err != nil {
		return nil, err
	}

	b, err = s.Repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if s.SMS != nil && s.Config.TwilioFromNumber != "" {
		customer := b.Customer
		body := notify.BookingConfirmationBody(
			b.ID.String(),
			b.Car.CarName,
			b.Car.CarNumber,
			"INR "+b.FinalBookingPrice.StringFixed(2),
		)
		to := normalizePhone(customer.PhoneNumber)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := s.SMS.Send(ctx, to, body); err != nil {
				log.Printf("twilio async: %v", err)
			}
		}()
	}

	return b, nil
}

func normalizePhone(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "+") {
		return p
	}
	digits := strings.Builder{}
	for _, r := range p {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	ds := digits.String()
	if len(ds) == 10 {
		return "+91" + ds
	}
	if len(ds) > 0 {
		return "+" + ds
	}
	return p
}

// --- messages ---

func (s *BookingService) ListMessages(ctx context.Context, userID, bookingID uuid.UUID) ([]models.Message, error) {
	if _, err := s.Get(ctx, userID, bookingID); err != nil {
		return nil, err
	}
	return s.Repo.ListMessagesForBooking(ctx, bookingID)
}

type PostMessageInput struct {
	Body string
}

func (s *BookingService) PostMessage(ctx context.Context, senderID, bookingID uuid.UUID, in PostMessageInput) (*models.Message, error) {
	b, err := s.Repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	if b.CustomerID != senderID && b.OwnerID != senderID {
		return nil, httpx.ErrForbidden
	}
	if b.Status == models.BookingConfirmed || b.Status == models.BookingCancelled {
		return nil, httpx.NewError(409, "BOOKING_CLOSED", "Messaging is closed for this booking.")
	}

	body := strings.TrimSpace(in.Body)
	if body == "" {
		return nil, httpx.WrapValidation("Message cannot be empty.")
	}

	m := &models.Message{
		BookingID: bookingID,
		SenderID:  senderID,
		Body:      body,
	}
	if err := s.Repo.CreateMessage(ctx, m); err != nil {
		return nil, err
	}
	if b.Status == models.BookingPending {
		b.Status = models.BookingNegotiating
		if err := s.Repo.UpdateBooking(ctx, b); err != nil {
			return nil, err
		}
	}

	// reload message with sender
	rows, err := s.Repo.ListMessagesForBooking(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	for i := len(rows) - 1; i >= 0; i-- {
		if rows[i].ID == m.ID {
			return &rows[i], nil
		}
	}
	return m, nil
}
