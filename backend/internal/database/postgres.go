package database

import (
	"fmt"

	"carmanage/backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgres opens a GORM connection. Schema is managed by backend/migrations (go run ./cmd/migrate up).
// GORM AutoMigrate is not run at startup — it conflicts with SQL migration constraint/index names on Postgres.
func NewPostgres(dsn string, logLevel logger.LogLevel) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	return db, nil
}

// AutoMigrate registers all models with GORM (optional; not used at server startup).
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.KYCAttachment{},
		&models.Car{},
		&models.CarImage{},
		&models.Booking{},
		&models.BookingPostTripCharge{},
		&models.BookingReview{},
		&models.Message{},
	)
}
