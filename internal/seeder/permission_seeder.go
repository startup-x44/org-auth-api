package seeder

import (
	"context"

	"auth-service/internal/models"

	"gorm.io/gorm"
)

// seedPermissions creates default system permissions
func (s *DatabaseSeeder) seedPermissions(ctx context.Context) error {
	permissions := []models.Permission{
		// Member permissions
		{
			Name:        "member:view",
			DisplayName: "View Members",
			Description: "View organization members",
			Category:    "member",
			IsSystem:    true,
		},
		{
			Name:        "member:invite",
			DisplayName: "Invite Members",
			Description: "Invite new members to organization",
			Category:    "member",
			IsSystem:    true,
		},
		{
			Name:        "member:update",
			DisplayName: "Update Members",
			Description: "Update member information and roles",
			Category:    "member",
			IsSystem:    true,
		},
		{
			Name:        "member:remove",
			DisplayName: "Remove Members",
			Description: "Remove members from organization",
			Category:    "member",
			IsSystem:    true,
		},

		// Role permissions
		{
			Name:        "role:view",
			DisplayName: "View Roles",
			Description: "View organization roles and permissions",
			Category:    "role",
			IsSystem:    true,
		},
		{
			Name:        "role:create",
			DisplayName: "Create Roles",
			Description: "Create new custom roles",
			Category:    "role",
			IsSystem:    true,
		},
		{
			Name:        "role:update",
			DisplayName: "Update Roles",
			Description: "Modify existing roles and permissions",
			Category:    "role",
			IsSystem:    true,
		},
		{
			Name:        "role:delete",
			DisplayName: "Delete Roles",
			Description: "Delete custom roles",
			Category:    "role",
			IsSystem:    true,
		},

		// Invitation permissions
		{
			Name:        "invitation:view",
			DisplayName: "View Invitations",
			Description: "View pending invitations",
			Category:    "invitation",
			IsSystem:    true,
		},
		{
			Name:        "invitation:resend",
			DisplayName: "Resend Invitations",
			Description: "Resend invitation emails",
			Category:    "invitation",
			IsSystem:    true,
		},
		{
			Name:        "invitation:cancel",
			DisplayName: "Cancel Invitations",
			Description: "Cancel pending invitations",
			Category:    "invitation",
			IsSystem:    true,
		},

		// Organization settings
		{
			Name:        "organization:update",
			DisplayName: "Update Organization",
			Description: "Update organization settings",
			Category:    "organization",
			IsSystem:    true,
		},
		{
			Name:        "organization:delete",
			DisplayName: "Delete Organization",
			Description: "Delete the organization",
			Category:    "organization",
			IsSystem:    true,
		},
	}

	for _, perm := range permissions {
		// Check if permission already exists
		var existing models.Permission
		result := s.db.WithContext(ctx).Where("name = ?", perm.Name).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			// Create permission if it doesn't exist
			if err := s.db.WithContext(ctx).Create(&perm).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
