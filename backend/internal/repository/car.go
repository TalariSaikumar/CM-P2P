package repository

import (
	"context"

	"carmanage/backend/internal/models"

	"github.com/google/uuid"
)

// CreateCar persists a car listing.
func (d *DB) CreateCar(ctx context.Context, c *models.Car) error {
	return d.WithContext(ctx).Create(c).Error
}

// GetCarByID loads a car with owner and images.
func (d *DB) GetCarByID(ctx context.Context, id uuid.UUID) (*models.Car, error) {
	var c models.Car
	if err := d.WithContext(ctx).Preload("Owner").Preload("Images").First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// ListCarsForOwner returns cars owned by a user.
func (d *DB) ListCarsForOwner(ctx context.Context, ownerID uuid.UUID) ([]models.Car, error) {
	var cars []models.Car
	if err := d.WithContext(ctx).Where("owner_id = ?", ownerID).Preload("Images").Order("created_at desc").Find(&cars).Error; err != nil {
		return nil, err
	}
	return cars, nil
}

// SearchCars filters active cars by optional location and model substrings (case-insensitive).
func (d *DB) SearchCars(ctx context.Context, location, model string) ([]models.Car, error) {
	q := d.WithContext(ctx).Model(&models.Car{}).Where("is_active = ?", true).Preload("Images").Order("created_at desc")
	if location != "" {
		q = q.Where("location ILIKE ?", "%"+location+"%")
	}
	if model != "" {
		q = q.Where("car_model ILIKE ?", "%"+model+"%")
	}
	var cars []models.Car
	if err := q.Find(&cars).Error; err != nil {
		return nil, err
	}
	return cars, nil
}

// UpdateCar updates editable fields.
func (d *DB) UpdateCar(ctx context.Context, c *models.Car) error {
	return d.WithContext(ctx).Save(c).Error
}

// DeleteCar soft-deletes a car.
func (d *DB) DeleteCar(ctx context.Context, id uuid.UUID) error {
	return d.WithContext(ctx).Delete(&models.Car{}, "id = ?", id).Error
}

// CreateCarImage persists an image row.
func (d *DB) CreateCarImage(ctx context.Context, img *models.CarImage) error {
	return d.WithContext(ctx).Create(img).Error
}

// MaxSortOrderForCar returns max sort_order for images on a car.
func (d *DB) MaxSortOrderForCar(ctx context.Context, carID uuid.UUID) (int, error) {
	var m int
	err := d.WithContext(ctx).Model(&models.CarImage{}).
		Select("COALESCE(MAX(sort_order), -1)").
		Where("car_id = ?", carID).
		Scan(&m).Error
	if err != nil {
		return -1, err
	}
	return m, nil
}
