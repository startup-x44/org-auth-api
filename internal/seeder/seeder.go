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
	// Seed tenants
	if err := s.seedTenants(ctx); err != nil {
		return err
	}

	// Seed users
	if err := s.seedUsers(ctx); err != nil {
		return err
	}

	return nil
}