package handler

import (
	"encoding/json"
	"time"

	"carmanage/backend/internal/models"
	"carmanage/backend/internal/service"

	"github.com/shopspring/decimal"
)

type userPublic struct {
	ID                   string  `json:"id"`
	Email                string  `json:"email"`
	Role                 string  `json:"role"`
	FullName             string  `json:"full_name"`
	AadhaarNumber        string  `json:"aadhaar_number"`
	PhoneNumber          string  `json:"phone_number"`
	Address              string  `json:"address"`
	DrivingLicenseNumber *string `json:"driving_license_number,omitempty"`
	IsKYCVerified        bool    `json:"is_kyc_verified"`
	CreatedAt            string  `json:"created_at"`
}

// userSummary is a reduced profile for counterpart views (bookings, chat).
type userSummary struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
}

func toUserSummary(u *models.User) userSummary {
	return userSummary{
		ID:          u.ID.String(),
		Email:       u.Email,
		Role:        string(u.Role),
		FullName:    u.FullName,
		PhoneNumber: u.PhoneNumber,
	}
}

func toUserPublic(u *models.User) userPublic {
	return userPublic{
		ID:                   u.ID.String(),
		Email:                u.Email,
		Role:                 string(u.Role),
		FullName:             u.FullName,
		AadhaarNumber:        u.AadhaarNumber,
		PhoneNumber:          u.PhoneNumber,
		Address:              u.Address,
		DrivingLicenseNumber: u.DrivingLicenseNumber,
		IsKYCVerified:        u.IsKYCVerified,
		CreatedAt:            u.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type carImagePublic struct {
	ID        string `json:"id"`
	BlobURL   string `json:"url"`
	SortOrder int    `json:"sort_order"`
}

type airbagDetailPublic struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type carPublic struct {
	ID                 string           `json:"id"`
	OwnerID            string           `json:"owner_id"`
	CarName            string           `json:"car_name"`
	CarModel           string           `json:"car_model"`
	CarNumber          string           `json:"car_number"`
	RegistrationNumber string           `json:"registration_number"`
	EngineNumber       string           `json:"engine_number"`
	PricePerHour       string           `json:"price_per_hour"`
	PricePerDay        string           `json:"price_per_day"`
	PricePerKm         string           `json:"price_per_km"`
	Location           string           `json:"location"`
	IsActive           bool             `json:"is_active"`
	Images             []carImagePublic `json:"images,omitempty"`
	CreatedAt          string           `json:"created_at"`

	ModelYear       int                  `json:"model_year"`
	Color           string               `json:"color"`
	FuelType        string               `json:"fuel_type"`
	Transmission    string               `json:"transmission"`
	MileageKm       int                  `json:"mileage_km"`
	NumSeats        int                  `json:"num_seats"`
	Airbags         bool                 `json:"airbags"`
	AirbagCount     int                  `json:"airbag_count"`
	AirbagDetails   []airbagDetailPublic `json:"airbag_details,omitempty"`
	CameraType      string               `json:"camera_type"`
	AirConditioning bool                 `json:"air_conditioning"`
	CruiseControl   bool                 `json:"cruise_control"`
	OpenRoof        bool                 `json:"open_roof"`
	Navigation      bool                 `json:"navigation"`
	Speakers        bool                 `json:"speakers"`
}

func decStr(d decimal.Decimal) string {
	return d.StringFixed(2)
}

// CarMineEntry is a car row for the owner fleet list (includes booking flag for today UTC).
type CarMineEntry struct {
	carPublic
	BookedForCurrentDate bool `json:"booked_for_current_date"`
}

func toCarPublic(c *models.Car) carPublic {
	out := carPublic{
		ID:                 c.ID.String(),
		OwnerID:            c.OwnerID.String(),
		CarName:            c.CarName,
		CarModel:           c.CarModel,
		CarNumber:          c.CarNumber,
		RegistrationNumber: c.RegistrationNumber,
		EngineNumber:       c.EngineNumber,
		PricePerHour:       decStr(c.PricePerHour),
		PricePerDay:        decStr(c.PricePerDay),
		PricePerKm:         decStr(c.PricePerKm),
		Location:           c.Location,
		IsActive:           c.IsActive,
		CreatedAt:          c.CreatedAt.UTC().Format(time.RFC3339),

		ModelYear:       c.ModelYear,
		Color:           c.Color,
		FuelType:        c.FuelType,
		Transmission:    c.Transmission,
		MileageKm:       c.MileageKm,
		NumSeats:        c.NumSeats,
		Airbags:         c.Airbags,
		AirbagCount:     c.AirbagCount,
		CameraType:      c.CameraType,
		AirConditioning: c.AirConditioning,
		CruiseControl:   c.CruiseControl,
		OpenRoof:        c.OpenRoof,
		Navigation:      c.Navigation,
		Speakers:        c.Speakers,
	}
	var ab []models.AirbagDetail
	if len(c.AirbagDetails) > 0 {
		_ = json.Unmarshal(c.AirbagDetails, &ab)
	}
	for _, d := range ab {
		out.AirbagDetails = append(out.AirbagDetails, airbagDetailPublic{Type: d.Type, Count: d.Count})
	}
	for _, im := range c.Images {
		out.Images = append(out.Images, carImagePublic{
			ID:        im.ID.String(),
			BlobURL:   im.BlobURL,
			SortOrder: im.SortOrder,
		})
	}
	return out
}

type bookingPostTripItemPublic struct {
	Label     string `json:"label"`
	AmountInr string `json:"amount_inr"`
}

type bookingPaymentPublic struct {
	PaymentStatus              string  `json:"payment_status"`
	PaymentMethod              string  `json:"payment_method,omitempty"`
	PaidAt                     *string `json:"paid_at,omitempty"`
	AgreedBaseInr              string  `json:"agreed_base_inr"`
	CustomerCommissionPercent  float64 `json:"customer_commission_percent"`
	OwnerCommissionPercent     float64 `json:"owner_commission_percent"`
	CustomerCommissionInr      string  `json:"customer_commission_inr"`
	OwnerCommissionInr         string  `json:"owner_commission_inr"`
	GstPercentOnCommission     float64 `json:"gst_percent_on_commission"`
	CustomerGstInr             string  `json:"customer_gst_inr"`
	OwnerGstInr                string  `json:"owner_gst_inr"`
	CustomerTotalInr           string  `json:"customer_total_inr"`
	OwnerNetInr                string  `json:"owner_net_inr"`
	PlatformCommissionTotalInr string  `json:"platform_commission_total_inr"`
	PaymentPhase               string                      `json:"payment_phase,omitempty"`
	DepositPercent             int                         `json:"deposit_percent,omitempty"`
	DepositDueInr              string                      `json:"deposit_due_inr,omitempty"`
	DepositPaidInr             string                      `json:"deposit_paid_inr,omitempty"`
	DepositPaidAt              *string                     `json:"deposit_paid_at,omitempty"`
	TripBalanceInr             string                      `json:"trip_balance_inr,omitempty"`
	PostTripChargesInr         string                      `json:"post_trip_charges_inr,omitempty"`
	FinalDueInr                string                      `json:"final_due_inr,omitempty"`
	OwnerProjectedPayoutInr    string                      `json:"owner_projected_payout_inr,omitempty"`
	PostTripItems              []bookingPostTripItemPublic `json:"post_trip_items,omitempty"`
}

type bookingCancellationPublic struct {
	Reason          string `json:"reason"`
	CancelledAt     string `json:"cancelled_at"`
	CancelledByRole string `json:"cancelled_by_role"`
}

type bookingHandoverPublic struct {
	PickupOdometerKm  *int    `json:"pickup_odometer_km,omitempty"`
	PickupFuelPercent *int    `json:"pickup_fuel_percent,omitempty"`
	PickupNotes       string  `json:"pickup_notes,omitempty"`
	PickupRecordedAt  *string `json:"pickup_recorded_at,omitempty"`
	ReturnOdometerKm  *int    `json:"return_odometer_km,omitempty"`
	ReturnFuelPercent *int    `json:"return_fuel_percent,omitempty"`
	ReturnNotes       string  `json:"return_notes,omitempty"`
	ReturnRecordedAt  *string `json:"return_recorded_at,omitempty"`
}

type bookingReviewPublic struct {
	Party     string      `json:"party"`
	Rating    int         `json:"rating"`
	Comment   string      `json:"comment"`
	Reviewer  userSummary `json:"reviewer"`
	CreatedAt string      `json:"created_at"`
}

func ptrRFC3339(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

func handoverFromBooking(b *models.Booking) *bookingHandoverPublic {
	if b.PickupHandoverAt == nil && b.ReturnHandoverAt == nil {
		return nil
	}
	return &bookingHandoverPublic{
		PickupOdometerKm:  b.PickupOdometerKM,
		PickupFuelPercent: b.PickupFuelPercent,
		PickupNotes:       b.PickupHandoverNotes,
		PickupRecordedAt:  ptrRFC3339(b.PickupHandoverAt),
		ReturnOdometerKm:  b.ReturnOdometerKM,
		ReturnFuelPercent: b.ReturnFuelPercent,
		ReturnNotes:       b.ReturnHandoverNotes,
		ReturnRecordedAt:  ptrRFC3339(b.ReturnHandoverAt),
	}
}

type bookingPublic struct {
	ID                string                 `json:"id"`
	CarID             string                 `json:"car_id"`
	CustomerID        string                 `json:"customer_id"`
	OwnerID           string                 `json:"owner_id"`
	Status            string                 `json:"status"`
	FinalBookingPrice *string                `json:"final_booking_price,omitempty"`
	CustomerNote      string                 `json:"customer_note,omitempty"`
	RentalFrom        string                 `json:"rental_from"`
	RentalTo          string                 `json:"rental_to"`
	PickupPoint       string                 `json:"pickup_point"`
	DropPoint         string                 `json:"drop_point"`
	CreatedAt         string                 `json:"created_at"`
	Car               carPublic              `json:"car"`
	Customer          userSummary            `json:"customer"`
	Owner             userSummary            `json:"owner"`
	Payment           *bookingPaymentPublic  `json:"payment,omitempty"`
	Cancellation      *bookingCancellationPublic `json:"cancellation,omitempty"`
	Handover          *bookingHandoverPublic     `json:"handover,omitempty"`
	Reviews           []bookingReviewPublic      `json:"reviews,omitempty"`
}

func ptrDec(d *decimal.Decimal) *string {
	if d == nil {
		return nil
	}
	s := d.StringFixed(2)
	return &s
}

func bookingPaymentFromBreakdown(b *models.Booking, bd service.PaymentBreakdown, sv service.PaymentSettlementView) *bookingPaymentPublic {
	var paidAt *string
	if b.PaidAt != nil && b.PaymentStatus == models.BookingPaymentPaid {
		s := b.PaidAt.UTC().Format(time.RFC3339)
		paidAt = &s
	}
	ps := b.PaymentStatus
	if ps == "" {
		ps = models.BookingPaymentUnpaid
	}
	out := &bookingPaymentPublic{
		PaymentStatus:              ps,
		PaymentMethod:              b.PaymentMethod,
		PaidAt:                     paidAt,
		AgreedBaseInr:              bd.AgreedBase.StringFixed(2),
		CustomerCommissionPercent:  bd.CustomerCommissionPct,
		OwnerCommissionPercent:     bd.OwnerCommissionPct,
		CustomerCommissionInr:      bd.CustomerCommissionAmount.StringFixed(2),
		OwnerCommissionInr:         bd.OwnerCommissionAmount.StringFixed(2),
		GstPercentOnCommission:     bd.GstPercentOnCommission,
		CustomerGstInr:             bd.CustomerGSTAmount.StringFixed(2),
		OwnerGstInr:                bd.OwnerGSTAmount.StringFixed(2),
		CustomerTotalInr:           bd.CustomerTotal.StringFixed(2),
		OwnerNetInr:                bd.OwnerNet.StringFixed(2),
		PlatformCommissionTotalInr: bd.PlatformTotal.StringFixed(2),
		PaymentPhase:               sv.Phase,
		DepositPercent:             sv.DepositPercent,
		DepositDueInr:              sv.DepositDueInr.StringFixed(2),
		DepositPaidInr:             sv.DepositPaidInr.StringFixed(2),
		DepositPaidAt:              ptrRFC3339(b.DepositPaidAt),
		TripBalanceInr:             sv.TripBalanceInr.StringFixed(2),
		PostTripChargesInr:         sv.PostTripChargesInr.StringFixed(2),
		FinalDueInr:                sv.FinalDueInr.StringFixed(2),
		OwnerProjectedPayoutInr:    sv.OwnerProjectedPayoutInr.StringFixed(2),
	}
	if len(b.PostTripCharges) > 0 {
		out.PostTripItems = make([]bookingPostTripItemPublic, 0, len(b.PostTripCharges))
		for i := range b.PostTripCharges {
			c := b.PostTripCharges[i]
			out.PostTripItems = append(out.PostTripItems, bookingPostTripItemPublic{
				Label:     c.Label,
				AmountInr: c.AmountInr.StringFixed(2),
			})
		}
	}
	return out
}

func toBookingPublic(b *models.Booking, svc *service.BookingService) bookingPublic {
	bp := bookingPublic{
		ID:                b.ID.String(),
		CarID:             b.CarID.String(),
		CustomerID:        b.CustomerID.String(),
		OwnerID:           b.OwnerID.String(),
		Status:            string(b.Status),
		FinalBookingPrice: ptrDec(b.FinalBookingPrice),
		CustomerNote:      b.CustomerNote,
		RentalFrom:        b.RentalFrom.UTC().Format(time.RFC3339),
		RentalTo:          b.RentalTo.UTC().Format(time.RFC3339),
		PickupPoint:       b.PickupPoint,
		DropPoint:         b.DropPoint,
		CreatedAt:         b.CreatedAt.UTC().Format(time.RFC3339),
		Car:               toCarPublic(&b.Car),
		Customer:          toUserSummary(&b.Customer),
		Owner:             toUserSummary(&b.Owner),
	}
	if svc != nil && b.Status == models.BookingConfirmed && b.FinalBookingPrice != nil {
		if bd, err := svc.BreakdownForBooking(b); err == nil {
			sv := svc.SettlementView(b, bd)
			bp.Payment = bookingPaymentFromBreakdown(b, bd, sv)
		}
	}
	if b.Status == models.BookingCancelled && b.CancelledAt != nil {
		role := "owner"
		if b.CancelledByUserID != nil && *b.CancelledByUserID == b.CustomerID {
			role = "customer"
		}
		bp.Cancellation = &bookingCancellationPublic{
			Reason:          b.CancellationReason,
			CancelledAt:     b.CancelledAt.UTC().Format(time.RFC3339),
			CancelledByRole: role,
		}
	}
	if h := handoverFromBooking(b); h != nil {
		bp.Handover = h
	}
	if len(b.Reviews) > 0 {
		bp.Reviews = make([]bookingReviewPublic, 0, len(b.Reviews))
		for i := range b.Reviews {
			r := b.Reviews[i]
			bp.Reviews = append(bp.Reviews, bookingReviewPublic{
				Party:     r.ReviewerParty,
				Rating:    r.Rating,
				Comment:   r.Comment,
				Reviewer:  toUserSummary(&r.Reviewer),
				CreatedAt: r.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
	}
	return bp
}

type messagePublic struct {
	ID        string      `json:"id"`
	BookingID string      `json:"booking_id"`
	SenderID  string      `json:"sender_id"`
	Body      string      `json:"body"`
	CreatedAt string      `json:"created_at"`
	Sender    userSummary `json:"sender"`
}

func toMessagePublic(m *models.Message) messagePublic {
	return messagePublic{
		ID:        m.ID.String(),
		BookingID: m.BookingID.String(),
		SenderID:  m.SenderID.String(),
		Body:      m.Body,
		CreatedAt: m.CreatedAt.UTC().Format(time.RFC3339),
		Sender:    toUserSummary(&m.Sender),
	}
}

type kycAttachmentPublic struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	BlobURL   string `json:"url"`
	CreatedAt string `json:"created_at"`
}
