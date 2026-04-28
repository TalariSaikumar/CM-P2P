package handler

import (
	"time"

	"carmanage/backend/internal/models"

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
}

func decStr(d decimal.Decimal) string {
	return d.StringFixed(2)
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

type bookingPublic struct {
	ID                  string  `json:"id"`
	CarID               string  `json:"car_id"`
	CustomerID          string  `json:"customer_id"`
	OwnerID             string  `json:"owner_id"`
	Status              string  `json:"status"`
	FinalBookingPrice   *string `json:"final_booking_price,omitempty"`
	CustomerNote        string  `json:"customer_note,omitempty"`
	CreatedAt           string      `json:"created_at"`
	Car                 carPublic   `json:"car"`
	Customer            userSummary `json:"customer"`
	Owner               userSummary `json:"owner"`
}

func ptrDec(d *decimal.Decimal) *string {
	if d == nil {
		return nil
	}
	s := d.StringFixed(2)
	return &s
}

func toBookingPublic(b *models.Booking) bookingPublic {
	bp := bookingPublic{
		ID:                b.ID.String(),
		CarID:             b.CarID.String(),
		CustomerID:        b.CustomerID.String(),
		OwnerID:           b.OwnerID.String(),
		Status:            string(b.Status),
		FinalBookingPrice: ptrDec(b.FinalBookingPrice),
		CustomerNote:      b.CustomerNote,
		CreatedAt:         b.CreatedAt.UTC().Format(time.RFC3339),
		Car:        toCarPublic(&b.Car),
		Customer:   toUserSummary(&b.Customer),
		Owner:      toUserSummary(&b.Owner),
	}
	return bp
}

type messagePublic struct {
	ID        string     `json:"id"`
	BookingID string     `json:"booking_id"`
	SenderID  string     `json:"sender_id"`
	Body      string     `json:"body"`
	CreatedAt string     `json:"created_at"`
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
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	BlobURL  string `json:"url"`
	CreatedAt string `json:"created_at"`
}
