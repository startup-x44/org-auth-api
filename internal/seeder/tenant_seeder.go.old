package seeder

import (
	"context"
	"time"

	"gorm.io/gorm"

	"auth-service/internal/models"
)

// seedTenants creates default tenants
func (s *DatabaseSeeder) seedTenants(ctx context.Context) error {
	tenants := []models.Tenant{
		{
			Name:      "Default Organization",
			Domain:    "default.local",
			Status:    models.TenantStatusActive,
			Settings:  "{}",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Name:      "Demo Company",
			Domain:    "demo.company.com",
			Status:    models.TenantStatusActive,
			Settings:  `{"features": ["advanced_reporting", "multi_user"]}`,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Name:      "Test Organization",
			Domain:    "test.org",
			Status:    models.TenantStatusActive,
			Settings:  "{}",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, tenant := range tenants {
		// Check if tenant already exists
		var existingTenant models.Tenant
		result := s.db.WithContext(ctx).Where("domain = ?", tenant.Domain).First(&existingTenant)
		if result.Error == gorm.ErrRecordNotFound {
			// Create tenant if it doesn't exist
			if err := s.db.WithContext(ctx).Create(&tenant).Error; err != nil {
				return err
			}
		}
	}

	return nil
}