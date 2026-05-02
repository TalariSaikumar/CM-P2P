package service

import (
	"context"
	"errors"
	"fmt"
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
	RentalFrom   time.Time
	RentalTo     time.Time
	PickupPoint  string
	DropPoint    string
}

// ParseBookingDateTime accepts RFC3339, date-only YYYY-MM-DD (UTC midnight), or YYYY-MM-DDTHH:MM:SS (UTC).
func ParseBookingDateTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("empty datetime")
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.UTC); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04:05", s, time.UTC); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04", s, time.UTC); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid datetime %q", s)
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

	pickup := strings.TrimSpace(in.PickupPoint)
	drop := strings.TrimSpace(in.DropPoint)
	if in.RentalTo.Before(in.RentalFrom) || in.RentalTo.Equal(in.RentalFrom) {
		return nil, httpx.WrapValidation("Rental end must be after rental start.")
	}

	overlap, err := s.Repo.HasCarBookingDateOverlap(ctx, car.ID, in.RentalFrom, in.RentalTo, nil)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, httpx.ErrCarAlreadyBooked
	}

	b := &models.Booking{
		CarID:        car.ID,
		CustomerID:   customerID,
		OwnerID:      car.OwnerID,
		Status:       models.BookingPending,
		CustomerNote: strings.TrimSpace(in.CustomerNote),
		RentalFrom:   in.RentalFrom.UTC(),
		RentalTo:     in.RentalTo.UTC(),
		PickupPoint:  pickup,
		DropPoint:    drop,
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

// ListForCustomerPaged returns bookings for the customer with total count.
func (s *BookingService) ListForCustomerPaged(ctx context.Context, customerID uuid.UUID, offset, limit int) ([]models.Booking, int64, error) {
	return s.Repo.ListBookingsForCustomerPaged(ctx, customerID, offset, limit)
}

// ListForOwnerPaged returns bookings for the owner with total count.
func (s *BookingService) ListForOwnerPaged(ctx context.Context, ownerID uuid.UUID, offset, limit int) ([]models.Booking, int64, error) {
	return s.Repo.ListBookingsForOwnerPaged(ctx, ownerID, offset, limit)
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

func (s *BookingService) Confirm(ctx context.Context, ownerID, bookingID uuid.UUID) (*models.Booking, error) {
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
	if b.Status == models.BookingConfirmed {
		return b, nil
	}
	if b.Status == models.BookingCancelled {
		return nil, httpx.NewError(409, "BOOKING_CANCELLED", "This booking is no longer active.")
	}
	if b.FinalBookingPrice == nil {
		return nil, httpx.NewError(400, "PRICE_NOT_SET", "Set the final agreed price before confirming.")
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

// Withdraw lets the customer cancel before the owner has set a final price (no negotiation locked in yet).
func (s *BookingService) Withdraw(ctx context.Context, customerID, bookingID uuid.UUID) (*models.Booking, error) {
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
	if b.Status == models.BookingConfirmed || b.Status == models.BookingCancelled {
		return nil, httpx.NewError(409, "BOOKING_CLOSED", "This booking can no longer be withdrawn.")
	}
	if b.FinalBookingPrice != nil {
		return nil, httpx.NewError(400, "PRICE_SET", "The owner has set a final price; you can no longer withdraw this inquiry.")
	}
	if b.Status != models.BookingPending && b.Status != models.BookingNegotiating {
		return nil, httpx.NewError(409, "BOOKING_STATE", "This booking cannot be withdrawn.")
	}

	b.Status = models.BookingCancelled
	if err := s.Repo.UpdateBooking(ctx, b); err != nil {
		return nil, err
	}
	return s.Repo.GetBookingByID(ctx, bookingID)
}

type UpdateTripDetailsInput struct {
	RentalFrom  time.Time
	RentalTo    time.Time
	PickupPoint string
	DropPoint   string
}

// UpdateTripDetails lets the customer change rental window and handover points while PENDING or NEGOTIATING.
func (s *BookingService) UpdateTripDetails(ctx context.Context, customerID, bookingID uuid.UUID, in UpdateTripDetailsInput) (*models.Booking, error) {
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
	if b.Status != models.BookingPending && b.Status != models.BookingNegotiating {
		return nil, httpx.NewError(409, "BOOKING_STATE", "Trip details can only be updated while the booking is pending or negotiating.")
	}
	if b.Status == models.BookingConfirmed || b.Status == models.BookingCancelled {
		return nil, httpx.NewError(409, "BOOKING_CLOSED", "This booking can no longer be updated.")
	}

	pickup := strings.TrimSpace(in.PickupPoint)
	drop := strings.TrimSpace(in.DropPoint)
	if pickup == "" || drop == "" {
		return nil, httpx.WrapValidation("Pickup point and drop-off point are required.")
	}
	if in.RentalTo.Before(in.RentalFrom) || in.RentalTo.Equal(in.RentalFrom) {
		return nil, httpx.WrapValidation("Rental end must be after rental start.")
	}

	overlap, err := s.Repo.HasCarBookingDateOverlap(ctx, b.CarID, in.RentalFrom, in.RentalTo, &b.ID)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, httpx.ErrCarAlreadyBooked
	}

	b.RentalFrom = in.RentalFrom.UTC()
	b.RentalTo = in.RentalTo.UTC()
	b.PickupPoint = pickup
	b.DropPoint = drop
	if err := s.Repo.UpdateBooking(ctx, b); err != nil {
		return nil, err
	}
	return s.Repo.GetBookingByID(ctx, bookingID)
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

func (s *BookingService) commissionRates() (customerPct, ownerPct, gstOnCommissionPct float64) {
	if s.Config == nil {
		return 2, 1.5, 18
	}
	c := s.Config.CustomerCommissionPercent
	o := s.Config.OwnerCommissionPercent
	g := s.Config.GstPercentOnCommission
	if c < 0 {
		c = 0
	}
	if o < 0 {
		o = 0
	}
	if g < 0 {
		g = 0
	}
	if g == 0 {
		g = 18
	}
	return c, o, g
}

// BreakdownForBooking returns commission math for the agreed price (uses stored amounts when already paid).
func (s *BookingService) BreakdownForBooking(b *models.Booking) (PaymentBreakdown, error) {
	if b.FinalBookingPrice == nil {
		return PaymentBreakdown{}, httpx.WrapValidation("Agreed price is not set.")
	}
	base := *b.FinalBookingPrice

	if b.PaymentStatus == models.BookingPaymentPaid &&
		b.CustomerCommissionRate != nil && b.OwnerCommissionRate != nil &&
		b.CustomerCommissionAmount != nil && b.OwnerCommissionAmount != nil &&
		b.CustomerTotalPaid != nil && b.OwnerNetPayout != nil {
		cPct, _ := b.CustomerCommissionRate.Float64()
		oPct, _ := b.OwnerCommissionRate.Float64()
		gstPct := 0.0
		if b.GstPercentOnCommission != nil {
			gstPct, _ = b.GstPercentOnCommission.Float64()
		}
		cgst := decimal.Zero
		if b.CustomerGSTAmount != nil {
			cgst = *b.CustomerGSTAmount
		}
		ogst := decimal.Zero
		if b.OwnerGSTAmount != nil {
			ogst = *b.OwnerGSTAmount
		}
		platform := b.CustomerCommissionAmount.Add(*b.OwnerCommissionAmount).Add(cgst).Add(ogst).Round(2)
		return PaymentBreakdown{
			AgreedBase:               base,
			CustomerCommissionPct:    cPct,
			OwnerCommissionPct:       oPct,
			GstPercentOnCommission:   gstPct,
			CustomerCommissionAmount: *b.CustomerCommissionAmount,
			OwnerCommissionAmount:    *b.OwnerCommissionAmount,
			CustomerGSTAmount:        cgst,
			OwnerGSTAmount:           ogst,
			CustomerTotal:            *b.CustomerTotalPaid,
			OwnerNet:                 *b.OwnerNetPayout,
			PlatformTotal:            platform,
		}, nil
	}

	c, o, g := s.commissionRates()
	return BuildPaymentBreakdown(base, c, o, g), nil
}

// CustomerPaymentPreview returns the INR breakdown for a confirmed booking (customer only).
func (s *BookingService) CustomerPaymentPreview(ctx context.Context, customerID, bookingID uuid.UUID) (PaymentBreakdown, error) {
	b, err := s.Repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return PaymentBreakdown{}, httpx.ErrNotFound
		}
		return PaymentBreakdown{}, err
	}
	if b.CustomerID != customerID {
		return PaymentBreakdown{}, httpx.ErrForbidden
	}
	if b.Status != models.BookingConfirmed {
		return PaymentBreakdown{}, httpx.ErrPaymentNotReady
	}
	if b.FinalBookingPrice == nil {
		return PaymentBreakdown{}, httpx.ErrPaymentNotReady
	}
	return s.BreakdownForBooking(b)
}

// CustomerRecordPayment simulates a successful payment and stores commission breakdown on the booking.
func (s *BookingService) CustomerRecordPayment(ctx context.Context, customerID, bookingID uuid.UUID, methodRaw string) (*models.Booking, error) {
	method := NormalizePaymentMethod(methodRaw)
	if method == "" {
		return nil, httpx.ErrInvalidPaymentMethod
	}

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
	if b.Status != models.BookingConfirmed {
		return nil, httpx.ErrPaymentNotReady
	}
	if b.FinalBookingPrice == nil {
		return nil, httpx.ErrPaymentNotReady
	}
	if b.PaymentStatus == models.BookingPaymentPaid {
		return s.Repo.GetBookingByID(ctx, bookingID)
	}

	bd, err := s.BreakdownForBooking(b)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	cr := decimal.NewFromFloat(bd.CustomerCommissionPct).Round(3)
	or := decimal.NewFromFloat(bd.OwnerCommissionPct).Round(3)
	cc := bd.CustomerCommissionAmount
	oc := bd.OwnerCommissionAmount
	cgst := bd.CustomerGSTAmount
	ogst := bd.OwnerGSTAmount
	ct := bd.CustomerTotal
	on := bd.OwnerNet
	gstPctDec := decimal.NewFromFloat(bd.GstPercentOnCommission).Round(2)

	b.PaymentStatus = models.BookingPaymentPaid
	b.PaymentMethod = method
	b.PaidAt = &now
	b.CustomerCommissionRate = &cr
	b.OwnerCommissionRate = &or
	b.CustomerCommissionAmount = &cc
	b.OwnerCommissionAmount = &oc
	b.CustomerGSTAmount = &cgst
	b.OwnerGSTAmount = &ogst
	b.GstPercentOnCommission = &gstPctDec
	b.CustomerTotalPaid = &ct
	b.OwnerNetPayout = &on

	if err := s.Repo.UpdateBooking(ctx, b); err != nil {
		return nil, err
	}
	return s.Repo.GetBookingByID(ctx, bookingID)
}
