package seeder

import (
	"context"
	"fmt"

	"auth-service/internal/models"

	"github.com/google/uuid"
)

// seedOrganizations creates test organizations with their custom owner roles
// NOTE: Roles are created automatically via CreateDefaultAdminRole when org is created
func (s *DatabaseSeeder) seedOrganizations(ctx context.Context) error {
	// Find test users
	var johnDoe, janeSmith, adminUser models.User
	if err := s.db.WithContext(ctx).Where("email = ?", "john.doe@example.com").First(&johnDoe).Error; err != nil {
		return fmt.Errorf("john.doe user not found: %w", err)
	}
	if err := s.db.WithContext(ctx).Where("email = ?", "jane.smith@example.com").First(&janeSmith).Error; err != nil {
		return fmt.Errorf("jane.smith user not found: %w", err)
	}
	if err := s.db.WithContext(ctx).Where("email = ?", "admin@example.com").First(&adminUser).Error; err != nil {
		return fmt.Errorf("admin user not found: %w", err)
	}

	// Create test organizations
	organizations := []struct {
		name  string
		slug  string
		owner models.User
	}{
		{
			name:  "Acme Corporation",
			slug:  "acme",
			owner: johnDoe,
		},
		{
			name:  "Tech Innovators",
			slug:  "tech-innovators",
			owner: janeSmith,
		},
		{
			name:  "Admin Organization",
			slug:  "admin-org",
			owner: adminUser,
		},
	}

	for _, orgData := range organizations {
		// Check if organization already exists
		var existingOrg models.Organization
		result := s.db.WithContext(ctx).Where("slug = ?", orgData.slug).First(&existingOrg)
		if result.Error == nil {
			// Organization already exists, skip
			continue
		}

		// Create organization
		org := models.Organization{
			ID:          uuid.New(),
			Name:        orgData.name,
			Slug:        orgData.slug,
			Description: stringPtr(fmt.Sprintf("Test organization for %s", orgData.owner.Email)),
			Settings:    "{}",
			CreatedBy:   orgData.owner.ID,
		}

		if err := s.db.WithContext(ctx).Create(&org).Error; err != nil {
			return fmt.Errorf("failed to create organization %s: %w", org.Name, err)
		}

		// Create CUSTOM owner role for this organization (NOT a system role)
		// This is now done via repository.CreateDefaultAdminRole
		ownerRole := models.Role{
			ID:             uuid.New(),
			OrganizationID: &org.ID,
			Name:           "owner",
			DisplayName:    "Owner",
			Description:    "Full access to organization (organization owner)",
			IsSystem:       false, // CUSTOM role, NOT system
			CreatedBy:      orgData.owner.ID,
		}

		if err := s.db.WithContext(ctx).Create(&ownerRole).Error; err != nil {
			return fmt.Errorf("failed to create owner role for %s: %w", org.Name, err)
		}

		// Assign all SYSTEM permissions to owner role
		var systemPermissions []models.Permission
		if err := s.db.WithContext(ctx).Where("is_system = ? AND organization_id IS NULL", true).Find(&systemPermissions).Error; err != nil {
			return fmt.Errorf("failed to load system permissions: %w", err)
		}

		for _, perm := range systemPermissions {
			rolePermission := models.RolePermission{
				RoleID:       ownerRole.ID,
				PermissionID: perm.ID,
			}
			if err := s.db.WithContext(ctx).Create(&rolePermission).Error; err != nil {
				return fmt.Errorf("failed to assign permission to owner role: %w", err)
			}
		}

		// Add owner as member with owner role
		membership := models.OrganizationMembership{
			ID:             uuid.New(),
			OrganizationID: org.ID,
			UserID:         orgData.owner.ID,
			RoleID:         ownerRole.ID,
			Status:         "active",
		}

		if err := s.db.WithContext(ctx).Create(&membership).Error; err != nil {
			return fmt.Errorf("failed to create membership for %s: %w", org.Name, err)
		}
	}

	return nil
}
