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
