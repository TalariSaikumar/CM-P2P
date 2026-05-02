package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"carmanage/backend/internal/httpx"
	"carmanage/backend/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CarService manages listings and media.
type CarService struct {
	Deps
}

type CreateCarInput struct {
	CarName            string
	CarModel           string
	CarNumber          string
	RegistrationNumber string
	EngineNumber       string
	PricePerHour       string
	PricePerDay        string
	PricePerKm         string
	Location           string

	ModelYear       int
	Color           string
	FuelType        string
	Transmission    string
	MileageKm       int
	NumSeats        int
	Airbags         bool
	AirbagCount     int
	AirbagDetails   []models.AirbagDetail
	CameraType      string
	AirConditioning bool
	CruiseControl   bool
	OpenRoof        bool
	Navigation      bool
	Speakers        bool
}

func normalizeFuelType(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "petrol":
		return "petrol", nil
	case "diesel":
		return "diesel", nil
	case "cng":
		return "cng", nil
	case "ev":
		return "ev", nil
	default:
		return "", httpx.WrapValidation("Fuel type must be petrol, diesel, cng, or ev.")
	}
}

func normalizeTransmission(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "auto", "automatic":
		return "auto", nil
	case "manual":
		return "manual", nil
	default:
		return "", httpx.WrapValidation("Transmission must be auto or manual.")
	}
}

func marshalAirbagDetails(airbags bool, airbagCount int, details []models.AirbagDetail) ([]byte, error) {
	if !airbags {
		return []byte("[]"), nil
	}
	if airbagCount < 1 {
		return nil, httpx.WrapValidation("When airbags are enabled, enter the total number of airbags.")
	}
	if len(details) == 0 {
		return nil, httpx.WrapValidation("When airbags are enabled, add at least one airbag type and count.")
	}
	sum := 0
	for _, d := range details {
		t := strings.TrimSpace(d.Type)
		if t == "" {
			return nil, httpx.WrapValidation("Each airbag entry needs a type (e.g. front, side curtain).")
		}
		if d.Count < 1 {
			return nil, httpx.WrapValidation("Each airbag entry needs a count of at least 1.")
		}
		sum += d.Count
	}
	if sum != airbagCount {
		return nil, httpx.WrapValidation("Airbag counts by type must add up to the total number of airbags.")
	}
	b, err := json.Marshal(details)
	if err != nil {
		return nil, httpx.WrapValidation("Could not save airbag details.")
	}
	return b, nil
}

func decodeAirbagDetails(raw []byte) []models.AirbagDetail {
	if len(raw) == 0 {
		return nil
	}
	var out []models.AirbagDetail
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func dayRangeUTC(ref time.Time) (from, to time.Time) {
	u := ref.UTC()
	from = time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
	to = from.Add(24 * time.Hour)
	return from, to
}

// CarBookedForCurrentUTCDate reports whether the car has an active booking whose rental
// window overlaps the current calendar day in UTC (same overlap rule as new bookings).
func (s *CarService) CarBookedForCurrentUTCDate(ctx context.Context, carID uuid.UUID) (bool, error) {
	from, to := dayRangeUTC(time.Now())
	return s.Repo.HasCarBookingDateOverlap(ctx, carID, from, to, nil)
}

// ListCarIDsBookedTodayUTC returns car IDs with an active booking overlapping today (UTC).
func (s *CarService) ListCarIDsBookedTodayUTC(ctx context.Context, carIDs []uuid.UUID) ([]uuid.UUID, error) {
	from, to := dayRangeUTC(time.Now())
	return s.Repo.ListCarIDsWithActiveBookingsOverlapping(ctx, carIDs, from, to)
}

func (s *CarService) Create(ctx context.Context, ownerID uuid.UUID, in CreateCarInput) (*models.Car, error) {
	ph, err := decimal.NewFromString(strings.TrimSpace(in.PricePerHour))
	if err != nil {
		return nil, httpx.WrapValidation("Price per hour must be a valid decimal number.")
	}
	pd, err := decimal.NewFromString(strings.TrimSpace(in.PricePerDay))
	if err != nil {
		return nil, httpx.WrapValidation("Price per day must be a valid decimal number.")
	}
	pk, err := decimal.NewFromString(strings.TrimSpace(in.PricePerKm))
	if err != nil {
		return nil, httpx.WrapValidation("Price per kilometer must be a valid decimal number.")
	}
	if ph.IsNegative() || pd.IsNegative() || pk.IsNegative() {
		return nil, httpx.WrapValidation("Prices cannot be negative.")
	}

	fuel, err := normalizeFuelType(in.FuelType)
	if err != nil {
		return nil, err
	}
	trans, err := normalizeTransmission(in.Transmission)
	if err != nil {
		return nil, err
	}
	y := in.ModelYear
	if y < 1980 || y > time.Now().Year()+2 {
		return nil, httpx.WrapValidation("Model year looks invalid.")
	}
	color := strings.TrimSpace(in.Color)
	if color == "" {
		return nil, httpx.WrapValidation("Color is required.")
	}
	if in.NumSeats < 1 || in.NumSeats > 20 {
		return nil, httpx.WrapValidation("Number of seats must be between 1 and 20.")
	}
	if in.MileageKm < 0 {
		return nil, httpx.WrapValidation("Mileage cannot be negative.")
	}
	abJSON, err := marshalAirbagDetails(in.Airbags, in.AirbagCount, in.AirbagDetails)
	if err != nil {
		return nil, err
	}
	abCount := in.AirbagCount
	if !in.Airbags {
		abCount = 0
	}

	c := &models.Car{
		OwnerID:            ownerID,
		CarName:            strings.TrimSpace(in.CarName),
		CarModel:           strings.TrimSpace(in.CarModel),
		CarNumber:          strings.TrimSpace(in.CarNumber),
		RegistrationNumber: strings.TrimSpace(in.RegistrationNumber),
		EngineNumber:       strings.TrimSpace(in.EngineNumber),
		PricePerHour:       ph,
		PricePerDay:        pd,
		PricePerKm:         pk,
		Location:           strings.TrimSpace(in.Location),
		IsActive:           true,

		ModelYear:       y,
		Color:           color,
		FuelType:        fuel,
		Transmission:    trans,
		MileageKm:       in.MileageKm,
		NumSeats:        in.NumSeats,
		Airbags:         in.Airbags,
		AirbagCount:     abCount,
		AirbagDetails:   abJSON,
		CameraType:      strings.TrimSpace(in.CameraType),
		AirConditioning: in.AirConditioning,
		CruiseControl:   in.CruiseControl,
		OpenRoof:        in.OpenRoof,
		Navigation:      in.Navigation,
		Speakers:        in.Speakers,
	}
	if c.CarName == "" || c.CarModel == "" || c.CarNumber == "" || c.RegistrationNumber == "" || c.EngineNumber == "" || c.Location == "" {
		return nil, httpx.WrapValidation("All vehicle identity fields and location are required.")
	}

	if err := s.Repo.CreateCar(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

type UpdateCarInput struct {
	CarName            *string
	CarModel           *string
	CarNumber          *string
	RegistrationNumber *string
	EngineNumber       *string
	PricePerHour       *string
	PricePerDay        *string
	PricePerKm         *string
	Location           *string
	IsActive           *bool

	ModelYear       *int
	Color           *string
	FuelType        *string
	Transmission    *string
	MileageKm       *int
	NumSeats        *int
	Airbags         *bool
	AirbagCount     *int
	AirbagDetails   *[]models.AirbagDetail
	CameraType      *string
	AirConditioning *bool
	CruiseControl   *bool
	OpenRoof        *bool
	Navigation      *bool
	Speakers        *bool
}

func (s *CarService) Update(ctx context.Context, ownerID, carID uuid.UUID, in UpdateCarInput) (*models.Car, error) {
	c, err := s.Repo.GetCarByID(ctx, carID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	if c.OwnerID != ownerID {
		return nil, httpx.ErrForbidden
	}
	bookedToday, err := s.CarBookedForCurrentUTCDate(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	if bookedToday {
		return nil, httpx.ErrCarBookedToday
	}
	if in.CarName != nil {
		c.CarName = strings.TrimSpace(*in.CarName)
	}
	if in.CarModel != nil {
		c.CarModel = strings.TrimSpace(*in.CarModel)
	}
	if in.CarNumber != nil {
		c.CarNumber = strings.TrimSpace(*in.CarNumber)
	}
	if in.RegistrationNumber != nil {
		c.RegistrationNumber = strings.TrimSpace(*in.RegistrationNumber)
	}
	if in.EngineNumber != nil {
		c.EngineNumber = strings.TrimSpace(*in.EngineNumber)
	}
	if in.Location != nil {
		c.Location = strings.TrimSpace(*in.Location)
	}
	if in.IsActive != nil {
		c.IsActive = *in.IsActive
	}
	if in.PricePerHour != nil {
		ph, err := decimal.NewFromString(strings.TrimSpace(*in.PricePerHour))
		if err != nil {
			return nil, httpx.WrapValidation("Price per hour must be a valid decimal number.")
		}
		c.PricePerHour = ph
	}
	if in.PricePerDay != nil {
		pd, err := decimal.NewFromString(strings.TrimSpace(*in.PricePerDay))
		if err != nil {
			return nil, httpx.WrapValidation("Price per day must be a valid decimal number.")
		}
		c.PricePerDay = pd
	}
	if in.PricePerKm != nil {
		pk, err := decimal.NewFromString(strings.TrimSpace(*in.PricePerKm))
		if err != nil {
			return nil, httpx.WrapValidation("Price per kilometer must be a valid decimal number.")
		}
		c.PricePerKm = pk
	}
	if in.ModelYear != nil {
		y := *in.ModelYear
		if y < 1980 || y > time.Now().Year()+2 {
			return nil, httpx.WrapValidation("Model year looks invalid.")
		}
		c.ModelYear = y
	}
	if in.Color != nil {
		col := strings.TrimSpace(*in.Color)
		if col == "" {
			return nil, httpx.WrapValidation("Color cannot be empty.")
		}
		c.Color = col
	}
	if in.FuelType != nil {
		f, err := normalizeFuelType(*in.FuelType)
		if err != nil {
			return nil, err
		}
		c.FuelType = f
	}
	if in.Transmission != nil {
		t, err := normalizeTransmission(*in.Transmission)
		if err != nil {
			return nil, err
		}
		c.Transmission = t
	}
	if in.MileageKm != nil {
		if *in.MileageKm < 0 {
			return nil, httpx.WrapValidation("Mileage cannot be negative.")
		}
		c.MileageKm = *in.MileageKm
	}
	if in.NumSeats != nil {
		if *in.NumSeats < 1 || *in.NumSeats > 20 {
			return nil, httpx.WrapValidation("Number of seats must be between 1 and 20.")
		}
		c.NumSeats = *in.NumSeats
	}
	if in.CameraType != nil {
		c.CameraType = strings.TrimSpace(*in.CameraType)
	}
	if in.AirConditioning != nil {
		c.AirConditioning = *in.AirConditioning
	}
	if in.CruiseControl != nil {
		c.CruiseControl = *in.CruiseControl
	}
	if in.OpenRoof != nil {
		c.OpenRoof = *in.OpenRoof
	}
	if in.Navigation != nil {
		c.Navigation = *in.Navigation
	}
	if in.Speakers != nil {
		c.Speakers = *in.Speakers
	}

	airbags := c.Airbags
	abCount := c.AirbagCount
	abDetailsSlice := decodeAirbagDetails(c.AirbagDetails)
	if in.Airbags != nil {
		airbags = *in.Airbags
		c.Airbags = airbags
	}
	if in.AirbagCount != nil {
		abCount = *in.AirbagCount
		c.AirbagCount = abCount
	}
	if in.AirbagDetails != nil {
		abDetailsSlice = *in.AirbagDetails
	}
	if in.Airbags != nil || in.AirbagCount != nil || in.AirbagDetails != nil {
		abJSON, err := marshalAirbagDetails(airbags, abCount, abDetailsSlice)
		if err != nil {
			return nil, err
		}
		if !airbags {
			c.AirbagCount = 0
		}
		c.AirbagDetails = abJSON
	}

	if err := s.Repo.UpdateCar(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CarService) Delete(ctx context.Context, ownerID, carID uuid.UUID) error {
	c, err := s.Repo.GetCarByID(ctx, carID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return httpx.ErrNotFound
		}
		return err
	}
	if c.OwnerID != ownerID {
		return httpx.ErrForbidden
	}
	return s.Repo.DeleteCar(ctx, carID)
}

func (s *CarService) AddImage(ctx context.Context, ownerID, carID uuid.UUID, filename, contentType string, body io.Reader) (*models.CarImage, error) {
	if s.Blob == nil {
		return nil, httpx.ErrStorage
	}
	c, err := s.Repo.GetCarByID(ctx, carID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	if c.OwnerID != ownerID {
		return nil, httpx.ErrForbidden
	}
	if body == nil {
		return nil, httpx.WrapValidation("Image file is required.")
	}

	safe := strings.ReplaceAll(strings.ToLower(filename), "..", "")
	if safe == "" {
		safe = "image"
	}
	blobPath := fmt.Sprintf("cars/%s/%s-%d", carID.String(), safe, time.Now().UnixNano())
	url, err := s.Blob.Upload(ctx, blobPath, body, contentType)
	if err != nil {
		return nil, fmt.Errorf("upload car image: %w", err)
	}

	order, err := s.Repo.MaxSortOrderForCar(ctx, carID)
	if err != nil {
		return nil, err
	}
	img := &models.CarImage{
		CarID:     carID,
		BlobPath:  blobPath,
		BlobURL:   url,
		SortOrder: order + 1,
	}
	if err := s.Repo.CreateCarImage(ctx, img); err != nil {
		return nil, err
	}
	return img, nil
}
