package repository

import (
	"context"
	"time"

	"carmanage/backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

// ListCarIDsWithActiveBookingsOverlapping returns distinct car_ids that have at least one
// PENDING, NEGOTIATING, or CONFIRMED booking overlapping [from, to) (same rule as HasCarBookingDateOverlap).
func (d *DB) ListCarIDsWithActiveBookingsOverlapping(ctx context.Context, carIDs []uuid.UUID, from, to time.Time) ([]uuid.UUID, error) {
	if len(carIDs) == 0 {
		return nil, nil
	}
	var ids []uuid.UUID
	err := d.WithContext(ctx).Model(&models.Booking{}).
		Where("car_id IN ?", carIDs).
		Where("status IN ?", activeBookingStatuses).
		Where("rental_from < ? AND rental_to > ?", to.UTC(), from.UTC()).
		Distinct("car_id").
		Pluck("car_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// CreateBooking persists a booking.
func (d *DB) CreateBooking(ctx context.Context, b *models.Booking) error {
	return d.WithContext(ctx).Create(b).Error
}

// GetBookingByID loads booking with relations.
func (d *DB) GetBookingByID(ctx context.Context, id uuid.UUID) (*models.Booking, error) {
	var b models.Booking
	if err := d.WithContext(ctx).
		Preload("Car").Preload("Car.Images").
		Preload("Customer").Preload("Owner").
		Preload("PostTripCharges", func(db *gorm.DB) *gorm.DB {
			return db.Order("booking_post_trip_charges.created_at asc")
		}).
		Preload("Reviews", func(db *gorm.DB) *gorm.DB {
			return db.Order("booking_reviews.created_at asc")
		}).
		Preload("Reviews.Reviewer").
		Preload("HandoverPhotos", func(db *gorm.DB) *gorm.DB {
			return db.Order("booking_handover_photos.created_at asc")
		}).
		First(&b, "id = ?", id).Error; err != nil {
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

// ListBookingsForCustomerPaged returns one page and total count for the customer.
func (d *DB) ListBookingsForCustomerPaged(ctx context.Context, customerID uuid.UUID, offset, limit int) ([]models.Booking, int64, error) {
	var total int64
	if err := d.WithContext(ctx).Model(&models.Booking{}).Where("customer_id = ?", customerID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.Booking
	err := d.WithContext(ctx).Where("customer_id = ?", customerID).
		Preload("Car").Preload("Car.Images").Preload("Owner").
		Order("created_at desc").Offset(offset).Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// ListBookingsForOwnerPaged returns one page and total count for the owner.
func (d *DB) ListBookingsForOwnerPaged(ctx context.Context, ownerID uuid.UUID, offset, limit int) ([]models.Booking, int64, error) {
	var total int64
	if err := d.WithContext(ctx).Model(&models.Booking{}).Where("owner_id = ?", ownerID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.Booking
	err := d.WithContext(ctx).Where("owner_id = ?", ownerID).
		Preload("Car").Preload("Car.Images").Preload("Customer").
		Order("created_at desc").Offset(offset).Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// UpdateBooking saves booking fields.
func (d *DB) UpdateBooking(ctx context.Context, b *models.Booking) error {
	return d.WithContext(ctx).Save(b).Error
}
