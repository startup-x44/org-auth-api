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

	// Hash passwords
	adminPassword, err := passwordService.Hash("Admin123!")
	if err != nil {
		return err
	}

	userPassword, err := passwordService.Hash("User123!")
	if err != nil {
		return err
	}

	users := []models.User{
		// Super admin user
		{
			Email:        "superadmin@platform.com",
			PasswordHash: adminPassword,
			IsSuperadmin: true,
			GlobalRole:   "admin",
			Firstname:    stringPtr("Super"),
			Lastname:     stringPtr("Administrator"),
			Phone:        stringPtr("+1234567890"),
			Status:       "active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// Regular users for testing
		{
			Email:        "john.doe@example.com",
			PasswordHash: userPassword,
			IsSuperadmin: false,
			GlobalRole:   "user",
			Firstname:    stringPtr("John"),
			Lastname:     stringPtr("Doe"),
			Phone:        stringPtr("+1555123456"),
			Status:       "active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Email:        "jane.smith@example.com",
			PasswordHash: userPassword,
			IsSuperadmin: false,
			GlobalRole:   "user",
			Firstname:    stringPtr("Jane"),
			Lastname:     stringPtr("Smith"),
			Phone:        stringPtr("+1555987654"),
			Status:       "active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Email:        "admin@example.com",
			PasswordHash: adminPassword,
			IsSuperadmin: false,
			GlobalRole:   "user",
			Firstname:    stringPtr("Admin"),
			Lastname:     stringPtr("User"),
			Phone:        stringPtr("+1555111111"),
			Status:       "active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

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

	return nil
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
