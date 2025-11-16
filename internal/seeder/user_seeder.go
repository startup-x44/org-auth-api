package seeder

import (
	"context"
	"time"

	"gorm.io/gorm"

	"auth-service/internal/models"
	"auth-service/pkg/password"
)

// seedUsers creates default users
func (s *DatabaseSeeder) seedUsers(ctx context.Context) error {
	// Get password service
	passwordService := password.NewService()

	// Hash password for superadmin
	adminPassword, err := passwordService.Hash("Admin123!")
	if err != nil {
		return err
	}

	users := []models.User{
		// Super admin user - NO organization required
		{
			Email:           "superadmin@platform.com",
			PasswordHash:    adminPassword,
			IsSuperadmin:    true,
			GlobalRole:      "superadmin",
			Firstname:       stringPtr("Super"),
			Lastname:        stringPtr("Admin"),
			Phone:           stringPtr("+1234567890"),
			Status:          "active",
			EmailVerifiedAt: timePtr(time.Now()), // Superadmin email is pre-verified
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	// Create superadmin user
	for _, user := range users {
		// Check if user already exists
		var existingUser models.User
		result := s.db.WithContext(ctx).Where("email = ?", user.Email).First(&existingUser)
		if result.Error == gorm.ErrRecordNotFound {
			// Create user if it doesn't exist
			if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
				return err
			}
		}
	}

	// Note: Regular users should be created through the registration flow
	// They need organization context which is set up during registration
	// For testing regular users with organizations, use the organization seeder

	return nil
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// timePtr returns a pointer to a time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}
