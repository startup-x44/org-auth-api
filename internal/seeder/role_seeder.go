package seeder

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"auth-service/internal/models"
)

// seedSystemRoles creates global system roles that are reusable across all organizations
func (s *DatabaseSeeder) seedSystemRoles(ctx context.Context) error {
	log.Println("Seeding system roles...")

	systemRoles := []models.Role{
		{
			ID:             uuid.New(),
			Name:           "owner",
			DisplayName:    "Owner",
			Description:    "Full access to all organization features",
			IsSystem:       true,
			OrganizationID: nil, // System role - no org ID
			CreatedBy:      uuid.Nil,
		},
		{
			ID:             uuid.New(),
			Name:           "admin",
			DisplayName:    "Administrator",
			Description:    "Administrative access to organization",
			IsSystem:       true,
			OrganizationID: nil, // System role - no org ID
			CreatedBy:      uuid.Nil,
		},
		{
			ID:             uuid.New(),
			Name:           "member",
			DisplayName:    "Member",
			Description:    "Standard member access",
			IsSystem:       true,
			OrganizationID: nil, // System role - no org ID
			CreatedBy:      uuid.Nil,
		},
	}

	for _, role := range systemRoles {
		// Check if role already exists
		var existing models.Role
		err := s.db.WithContext(ctx).
			Where("name = ? AND is_system = ? AND organization_id IS NULL", role.Name, true).
			First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// Create new system role
			if err := s.db.WithContext(ctx).Create(&role).Error; err != nil {
				return fmt.Errorf("failed to create system role %s: %w", role.Name, err)
			}
			log.Printf("Created system role: %s", role.DisplayName)

			// Assign permissions based on role
			if role.Name == "owner" || role.Name == "admin" {
				// Get all system permissions
				var permissions []models.Permission
				adminPermNames := models.DefaultAdminPermissions()
				if err := s.db.WithContext(ctx).
					Where("name IN ? AND is_system = ? AND organization_id IS NULL", adminPermNames, true).
					Find(&permissions).Error; err != nil {
					return fmt.Errorf("failed to get permissions for role %s: %w", role.Name, err)
				}

				// Assign all permissions to owner/admin roles
				for _, perm := range permissions {
					rolePermission := &models.RolePermission{
						RoleID:       role.ID,
						PermissionID: perm.ID,
					}
					if err := s.db.WithContext(ctx).Create(rolePermission).Error; err != nil {
						return fmt.Errorf("failed to assign permission %s to role %s: %w", perm.Name, role.Name, err)
					}
				}
				log.Printf("Assigned %d permissions to %s role", len(permissions), role.DisplayName)
			}
		} else if err != nil {
			return fmt.Errorf("failed to check existing role %s: %w", role.Name, err)
		} else {
			log.Printf("System role already exists: %s", role.DisplayName)
		}
	}

	return nil
}
