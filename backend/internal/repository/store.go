package repository

import "gorm.io/gorm"

// DB wraps GORM for repository methods.
type DB struct {
	*gorm.DB
}

// New returns a repository wrapper.
func New(db *gorm.DB) *DB {
	return &DB{db}
}
