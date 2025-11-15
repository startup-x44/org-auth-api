package seeder

import (
	"context"

	"gorm.io/gorm"
)

// Seeder defines the interface for database seeding operations
type Seeder interface {
	Seed(ctx context.Context) error
}

// DatabaseSeeder implements the main seeder
type DatabaseSeeder struct {
	db *gorm.DB
}

// NewDatabaseSeeder creates a new database seeder
func NewDatabaseSeeder(db *gorm.DB) Seeder {
	return &DatabaseSeeder{db: db}
}

// Seed runs all database seeders
func (s *DatabaseSeeder) Seed(ctx context.Context) error {
	// TODO: Seed organizations and users for development
	// For now, skip seeding to avoid tenant dependencies
	return nil
}
