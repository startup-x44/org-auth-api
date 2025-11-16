package seeder

import (
	"context"
	"fmt"
	"log"

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

// Seed runs all database seeders in correct order
func (s *DatabaseSeeder) Seed(ctx context.Context) error {
	log.Println("🌱 Starting database seeding...")

	// 1. Seed system permissions
	if err := s.seedPermissions(ctx); err != nil {
		return fmt.Errorf("failed to seed permissions: %w", err)
	}
	log.Println("✅ Permissions seeded")

	// 2. Seed users (including superadmin)
	if err := s.seedUsers(ctx); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}
	log.Println("✅ Users seeded")

	// 3. Seed test organizations with roles
	if err := s.seedOrganizations(ctx); err != nil {
		return fmt.Errorf("failed to seed organizations: %w", err)
	}
	log.Println("✅ Organizations seeded")

	log.Println("🎉 Database seeding completed successfully!")
	return nil
}
