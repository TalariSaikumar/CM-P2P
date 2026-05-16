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
	CarID                     uuid.UUID
	CustomerNote              string
	RentalFrom                time.Time
	RentalTo                  time.Time
	PickupPoint               string
	DropPoint                 string
	AcknowledgedDepositTerms bool
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

	if !in.AcknowledgedDepositTerms {
		return nil, httpx.ErrBookingTermsNotAck
	}

	nowAck := time.Now().UTC()
	b := &models.Booking{
		CarID:                       car.ID,
		CustomerID:                  customerID,
		OwnerID:                     car.OwnerID,
		Status:                      models.BookingPending,
		CustomerNote:                strings.TrimSpace(in.CustomerNote),
		RentalFrom:                  in.RentalFrom.UTC(),
		RentalTo:                    in.RentalTo.UTC(),
		PickupPoint:                 pickup,
		DropPoint:                   drop,
		CustomerAcknowledgedTermsAt: &nowAck,
	}
	if err := s.Repo.CreateBooking(ctx, b); err != nil {
		return nil, err
	}

	note := strings.TrimSpace(in.CustomerNote)
	if note != "" {
		m := &models.Message{
			BookingID: b.ID,
			SenderID:  customerID,
			Body:      note,
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
	b.CustomerAcceptedPriceAt = nil
	b.CustomerAcceptedPriceAmount = nil
	if err := s.Repo.UpdateBooking(ctx, b); err != nil {
		return nil, err
	}
	return s.Repo.GetBookingByID(ctx, bookingID)
}

// CustomerHasAcceptedQuotedPrice is true when the customer accepted the owner's current quoted amount.
func CustomerHasAcceptedQuotedPrice(b *models.Booking) bool {
	if b.FinalBookingPrice == nil || b.CustomerAcceptedPriceAt == nil || b.CustomerAcceptedPriceAmount == nil {
		return false
	}
	return b.FinalBookingPrice.Equal(*b.CustomerAcceptedPriceAmount)
}

// CustomerAcceptQuotedPrice records that the customer is satisfied with the owner's current quoted price.
func (s *BookingService) CustomerAcceptQuotedPrice(ctx context.Context, customerID, bookingID uuid.UUID) (*models.Booking, error) {
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
		return nil, httpx.NewError(409, "BOOKING_CLOSED", "This booking can no longer be updated.")
	}
	if b.FinalBookingPrice == nil {
		return nil, httpx.ErrPriceNotQuoted
	}
	if CustomerHasAcceptedQuotedPrice(b) {
		return b, nil
	}

	now := time.Now().UTC()
	amt := b.FinalBookingPrice.Copy()
	b.CustomerAcceptedPriceAt = &now
	b.CustomerAcceptedPriceAmount = &amt
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
	if !CustomerHasAcceptedQuotedPrice(b) {
		return nil, httpx.ErrCustomerPriceNotAccepted
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
	now := time.Now().UTC()
	b.CancellationReason = "Withdrawn by customer before a final price was set."
	b.CancelledAt = &now
	b.CancelledByUserID = &customerID
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
	if b.Status == models.BookingCancelled || b.Status == models.BookingCompleted {
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

// BreakdownForBooking returns commission math on trip rental (per-day negotiated rate × trip days).
func (s *BookingService) BreakdownForBooking(b *models.Booking) (PaymentBreakdown, error) {
	if b.FinalBookingPrice == nil {
		return PaymentBreakdown{}, httpx.WrapValidation("Agreed price is not set.")
	}
	tripBase := AgreedRentalBaseForBooking(b)

	cPct, oPct, gstPct := s.commissionRates()
	if b.PaymentStatus == models.BookingPaymentPaid {
		if b.CustomerCommissionRate != nil {
			cPct, _ = b.CustomerCommissionRate.Float64()
		}
		if b.OwnerCommissionRate != nil {
			oPct, _ = b.OwnerCommissionRate.Float64()
		}
		if b.GstPercentOnCommission != nil {
			gstPct, _ = b.GstPercentOnCommission.Float64()
		}
	}

	bd := BuildPaymentBreakdown(tripBase, cPct, oPct, gstPct)
	return bd, nil
}

// CustomerPaymentPreview returns the booking (with relations) and INR breakdown for a confirmed booking (customer only).
func (s *BookingService) CustomerPaymentPreview(ctx context.Context, customerID, bookingID uuid.UUID) (*models.Booking, PaymentBreakdown, error) {
	b, err := s.Repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, PaymentBreakdown{}, httpx.ErrNotFound
		}
		return nil, PaymentBreakdown{}, err
	}
	if b.CustomerID != customerID {
		return nil, PaymentBreakdown{}, httpx.ErrForbidden
	}
	if b.Status != models.BookingConfirmed {
		return nil, PaymentBreakdown{}, httpx.ErrPaymentNotReady
	}
	if b.FinalBookingPrice == nil {
		return nil, PaymentBreakdown{}, httpx.ErrPaymentNotReady
	}
	bd, err := s.BreakdownForBooking(b)
	if err != nil {
		return nil, PaymentBreakdown{}, err
	}
	return b, bd, nil
}

// CustomerRecordPayment records the 75% deposit first, then the final balance after the owner submits post-trip charges.
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
	sv := s.SettlementView(b, bd)
	now := time.Now().UTC()
	cr := decimal.NewFromFloat(bd.CustomerCommissionPct).Round(3)
	or := decimal.NewFromFloat(bd.OwnerCommissionPct).Round(3)
	cc := bd.CustomerCommissionAmount
	oc := bd.OwnerCommissionAmount
	cgst := bd.CustomerGSTAmount
	ogst := bd.OwnerGSTAmount
	gstPctDec := decimal.NewFromFloat(bd.GstPercentOnCommission).Round(2)

	switch b.PaymentStatus {
	case models.BookingPaymentUnpaid:
		dep := sv.DepositDueInr
		if dep.IsZero() {
			return nil, httpx.WrapValidation("Deposit amount is invalid.")
		}
		b.PaymentStatus = models.BookingPaymentDepositPaid
		b.DepositPaidAt = &now
		b.DepositCustomerTotal = &dep
		b.PaymentMethod = method
		if err := s.Repo.UpdateBooking(ctx, b); err != nil {
			return nil, err
		}
		return s.Repo.GetBookingByID(ctx, bookingID)

	case models.BookingPaymentDepositPaid:
		return nil, httpx.ErrSettlementNotReady

	case models.BookingPaymentFinalDue:
		ctFull := bd.CustomerTotal.Add(b.PostTripChargesTotal).Round(2)

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
		b.CustomerTotalPaid = &ctFull
		ownerRentalNet := bd.OwnerNet.Round(2)
		b.OwnerNetPayout = &ownerRentalNet

		if err := s.Repo.UpdateBooking(ctx, b); err != nil {
			return nil, err
		}
		return s.Repo.GetBookingByID(ctx, bookingID)

	default:
		return nil, httpx.ErrPaymentNotReady
	}
}

// CancelBooking ends an inquiry or an unpaid confirmed booking. Paid trips cannot be self-cancelled.
func (s *BookingService) CancelBooking(ctx context.Context, userID, bookingID uuid.UUID, reason string) (*models.Booking, error) {
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
	if b.Status == models.BookingCancelled {
		return nil, httpx.ErrConflict
	}
	switch b.Status {
	case models.BookingPending, models.BookingNegotiating:
		// ok
	case models.BookingConfirmed:
		if b.PaymentStatus != models.BookingPaymentUnpaid {
			return nil, httpx.ErrBookingPaidNoCancel
		}
	default:
		return nil, httpx.NewError(409, "BOOKING_STATE", "This booking cannot be cancelled.")
	}

	reason = strings.TrimSpace(reason)
	if len(reason) > 2000 {
		reason = reason[:2000]
	}
	if reason == "" {
		reason = "Cancelled by user."
	}
	now := time.Now().UTC()
	b.Status = models.BookingCancelled
	b.CancellationReason = reason
	b.CancelledAt = &now
	b.CancelledByUserID = &userID
	if err := s.Repo.UpdateBooking(ctx, b); err != nil {
		return nil, err
	}
	return s.Repo.GetBookingByID(ctx, bookingID)
}

type PatchHandoverInput struct {
	Phase       string
	OdometerKM  int
	FuelPercent *int
	Notes       string
}

// PatchHandover records owner vehicle handover, customer pickup check-in, or return handover.
func (s *BookingService) PatchHandover(ctx context.Context, userID, bookingID uuid.UUID, in PatchHandoverInput) (*models.Booking, error) {
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
	if b.Status != models.BookingConfirmed {
		return nil, httpx.NewError(409, "BOOKING_STATE", "Handover can only be recorded after the booking is confirmed.")
	}
	if !DepositPaidForTrip(b) {
		return nil, httpx.ErrDepositRequiredForHandover
	}
	phase := strings.ToLower(strings.TrimSpace(in.Phase))
	if phase != "pickup" && phase != "return" {
		return nil, httpx.WrapValidation(`Phase must be "pickup" or "return".`)
	}
	if in.OdometerKM <= 0 || in.OdometerKM > 2000000 {
		return nil, httpx.WrapValidation("Odometer must be a positive distance in km.")
	}
	if in.FuelPercent != nil {
		fp := *in.FuelPercent
		if fp < 0 || fp > 100 {
			return nil, httpx.WrapValidation("Fuel percent must be between 0 and 100.")
		}
	}
	notes := strings.TrimSpace(in.Notes)
	if len(notes) > 5000 {
		notes = notes[:5000]
	}
	now := time.Now().UTC()
	od := in.OdometerKM
	isOwner := b.OwnerID == userID
	isCustomer := b.CustomerID == userID

	if phase == "pickup" {
		switch {
		case isOwner:
			if b.OwnerPickupHandoverAt != nil {
				return nil, httpx.ErrOwnerHandoverExists
			}
			b.OwnerPickupOdometerKM = &od
			b.OwnerPickupFuelPercent = in.FuelPercent
			b.OwnerPickupHandoverNotes = notes
			b.OwnerPickupHandoverAt = &now
		case isCustomer:
			if !OwnerPickupRecorded(b) {
				return nil, httpx.ErrOwnerHandoverRequired
			}
			if CustomerPickupComplete(b) {
				return nil, httpx.ErrCustomerPickupExists
			}
			b.PickupOdometerKM = &od
			b.PickupFuelPercent = in.FuelPercent
			b.PickupHandoverNotes = notes
			b.PickupHandoverAt = &now
			b.CustomerPickupAcceptedAt = &now
		default:
			return nil, httpx.ErrForbidden
		}
	} else {
		if !CustomerPickupComplete(b) {
			return nil, httpx.ErrOwnerHandoverRequired
		}
		switch {
		case isCustomer:
			if CustomerReturnRecorded(b) {
				return nil, httpx.ErrCustomerReturnExists
			}
			if po := EffectivePickupOdometerKM(b); po != nil && od < *po {
				return nil, httpx.WrapValidation("Return odometer must be greater than or equal to pickup odometer.")
			}
			b.ReturnOdometerKM = &od
			b.ReturnFuelPercent = in.FuelPercent
			b.ReturnHandoverNotes = notes
			b.ReturnHandoverAt = &now
		case isOwner:
			if !CustomerReturnRecorded(b) {
				return nil, httpx.ErrCustomerReturnRequired
			}
			if b.PaymentStatus != models.BookingPaymentPaid {
				return nil, httpx.ErrFinalPaymentRequiredForReturnAccept
			}
			if OwnerReturnAccepted(b) {
				return nil, httpx.ErrOwnerReturnAlreadyAccepted
			}
			b.OwnerReturnAcceptedAt = &now
			b.Status = models.BookingCompleted
		default:
			return nil, httpx.ErrForbidden
		}
	}
	if err := s.Repo.UpdateBooking(ctx, b); err != nil {
		return nil, err
	}
	return s.Repo.GetBookingByID(ctx, bookingID)
}

// SubmitReview adds a 1–5 star review after rental_to (UTC) once payment is PAID.
func (s *BookingService) SubmitReview(ctx context.Context, reviewerID, bookingID uuid.UUID, rating int, comment string) (*models.BookingReview, error) {
	b, err := s.Repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	if b.Status != models.BookingConfirmed && b.Status != models.BookingCompleted {
		return nil, httpx.NewError(409, "BOOKING_STATE", "Reviews are only available for confirmed or completed bookings.")
	}
	if b.PaymentStatus != models.BookingPaymentPaid {
		return nil, httpx.NewError(400, "REVIEW_NOT_READY", "Payment must be completed before reviews open.")
	}
	if !time.Now().UTC().After(b.RentalTo.UTC()) {
		return nil, httpx.WrapValidation("Reviews open after the rental end date.")
	}
	var party string
	switch reviewerID {
	case b.CustomerID:
		party = "CUSTOMER"
	case b.OwnerID:
		party = "OWNER"
	default:
		return nil, httpx.ErrForbidden
	}
	for i := range b.Reviews {
		if b.Reviews[i].ReviewerParty == party {
			return nil, httpx.ErrReviewExists
		}
	}
	if rating < 1 || rating > 5 {
		return nil, httpx.WrapValidation("Rating must be between 1 and 5.")
	}
	comment = strings.TrimSpace(comment)
	if len(comment) > 5000 {
		comment = comment[:5000]
	}
	r := &models.BookingReview{
		BookingID:     bookingID,
		ReviewerParty: party,
		ReviewerID:    reviewerID,
		Rating:        rating,
		Comment:       comment,
	}
	if err := s.Repo.CreateBookingReview(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}
