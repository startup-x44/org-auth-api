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

	// Get default tenant
	var defaultTenant models.Tenant
	if err := s.db.WithContext(ctx).Where("domain = ?", "default.local").First(&defaultTenant).Error; err != nil {
		return err
	}

	// Get demo tenant
	var demoTenant models.Tenant
	if err := s.db.WithContext(ctx).Where("domain = ?", "demo.company.com").First(&demoTenant).Error; err != nil {
		return err
	}

	// Hash passwords
	adminPassword, err := passwordService.Hash("Admin123!")
	if err != nil {
		return err
	}

	studentPassword, err := passwordService.Hash("Student123!")
	if err != nil {
		return err
	}

	rtoPassword, err := passwordService.Hash("RtoPass123!")
	if err != nil {
		return err
	}

	users := []models.User{
		// Admin users
		{
			Email:        "admin@default.local",
			PasswordHash: adminPassword,
			UserType:     models.UserTypeAdmin,
			TenantID:     defaultTenant.ID,
			Firstname:    stringPtr("System"),
			Lastname:     stringPtr("Administrator"),
			Phone:        stringPtr("+1234567890"),
			Status:       models.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Email:        "admin@demo.company.com",
			PasswordHash: adminPassword,
			UserType:     models.UserTypeAdmin,
			TenantID:     demoTenant.ID,
			Firstname:    stringPtr("Demo"),
			Lastname:     stringPtr("Admin"),
			Phone:        stringPtr("+1987654321"),
			Status:       models.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// Student users
		{
			Email:        "student@default.local",
			PasswordHash: studentPassword,
			UserType:     models.UserTypeStudent,
			TenantID:     defaultTenant.ID,
			Firstname:    stringPtr("John"),
			Lastname:     stringPtr("Doe"),
			Phone:        stringPtr("+1555123456"),
			Status:       models.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Email:        "student@demo.company.com",
			PasswordHash: studentPassword,
			UserType:     models.UserTypeStudent,
			TenantID:     demoTenant.ID,
			Firstname:    stringPtr("Jane"),
			Lastname:     stringPtr("Smith"),
			Phone:        stringPtr("+1555987654"),
			Status:       models.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// RTO users
		{
			Email:        "rto@default.local",
			PasswordHash: rtoPassword,
			UserType:     models.UserTypeRTO,
			TenantID:     defaultTenant.ID,
			Firstname:    stringPtr("Training"),
			Lastname:     stringPtr("Organization"),
			Phone:        stringPtr("+1555111111"),
			Status:       models.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Email:        "rto@demo.company.com",
			PasswordHash: rtoPassword,
			UserType:     models.UserTypeRTO,
			TenantID:     demoTenant.ID,
			Firstname:    stringPtr("Demo"),
			Lastname:     stringPtr("RTO"),
			Phone:        stringPtr("+1555222222"),
			Status:       models.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, user := range users {
		// Check if user already exists
		var existingUser models.User
		result := s.db.WithContext(ctx).Where("email = ? AND tenant_id = ?", user.Email, user.TenantID).First(&existingUser)
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
