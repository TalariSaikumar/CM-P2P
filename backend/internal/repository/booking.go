package repository

import (
	"context"
	"time"

	"carmanage/backend/internal/models"

	"github.com/google/uuid"
)

// activeBookingStatuses block another rental when their windows overlap.
var activeBookingStatuses = []models.BookingStatus{
	models.BookingPending,
	models.BookingNegotiating,
	models.BookingConfirmed,
}

// HasCarBookingDateOverlap reports whether another booking for the same car
// overlaps [from, to) (half-open style: overlap if existing.rental_from < to && existing.rental_to > from).
// If excludeBookingID is non-nil, that booking row is ignored (for trip updates).
func (d *DB) HasCarBookingDateOverlap(ctx context.Context, carID uuid.UUID, from, to time.Time, excludeBookingID *uuid.UUID) (bool, error) {
	q := d.WithContext(ctx).Model(&models.Booking{}).
		Where("car_id = ?", carID).
		Where("status IN ?", activeBookingStatuses).
		Where("rental_from < ? AND rental_to > ?", to.UTC(), from.UTC())
	if excludeBookingID != nil {
		q = q.Where("id != ?", *excludeBookingID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

// CreateBooking persists a booking.
func (d *DB) CreateBooking(ctx context.Context, b *models.Booking) error {
	return d.WithContext(ctx).Create(b).Error
}

// GetBookingByID loads booking with relations.
func (d *DB) GetBookingByID(ctx context.Context, id uuid.UUID) (*models.Booking, error) {
	var b models.Booking
	if err := d.WithContext(ctx).Preload("Car").Preload("Car.Images").Preload("Customer").Preload("Owner").First(&b, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

// ListBookingsForCustomer returns bookings where the user is the customer.
func (d *DB) ListBookingsForCustomer(ctx context.Context, customerID uuid.UUID) ([]models.Booking, error) {
	var rows []models.Booking
	if err := d.WithContext(ctx).Where("customer_id = ?", customerID).
		Preload("Car").Preload("Car.Images").Preload("Owner").
		Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ListBookingsForOwner returns bookings where the user is the owner.
func (d *DB) ListBookingsForOwner(ctx context.Context, ownerID uuid.UUID) ([]models.Booking, error) {
	var rows []models.Booking
	if err := d.WithContext(ctx).Where("owner_id = ?", ownerID).
		Preload("Car").Preload("Car.Images").Preload("Customer").
		Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// UpdateBooking saves booking fields.
func (d *DB) UpdateBooking(ctx context.Context, b *models.Booking) error {
	return d.WithContext(ctx).Save(b).Error
}
