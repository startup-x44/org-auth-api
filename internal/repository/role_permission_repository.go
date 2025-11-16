package repository

import (
	"context"
	"fmt"

	"auth-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RolePermissionRepository defines the interface for role_permission operations
type RolePermissionRepository interface {
	GetByRole(ctx context.Context, roleID uuid.UUID) ([]*models.RolePermission, error)
	HasPermission(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) (bool, error)
	Create(ctx context.Context, rolePermission *models.RolePermission) error
	Delete(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error
	ListAll(ctx context.Context) ([]*models.RolePermission, error) // DEPRECATED: Unsafe global access
	// Organization-aware methods for security
	CreateWithValidation(ctx context.Context, rolePermission *models.RolePermission) error
	GetByRoleWithOrgValidation(ctx context.Context, roleID, orgID uuid.UUID) ([]*models.RolePermission, error)
}

type rolePermissionRepository struct {
	db *gorm.DB
}

func NewRolePermissionRepository(db *gorm.DB) RolePermissionRepository {
	return &rolePermissionRepository{db: db}
}

// Get all permission assignments for a role
func (r *rolePermissionRepository) GetByRole(ctx context.Context, roleID uuid.UUID) ([]*models.RolePermission, error) {
	var rolePerms []*models.RolePermission

	err := r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Find(&rolePerms).Error

	return rolePerms, err
}

// Check if a role has a specific permission
func (r *rolePermissionRepository) HasPermission(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.RolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Assign a permission to a role
func (r *rolePermissionRepository) Create(ctx context.Context, rolePermission *models.RolePermission) error {
	return r.db.WithContext(ctx).Create(rolePermission).Error
}

// Remove a permission from a role
func (r *rolePermissionRepository) Delete(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&models.RolePermission{}).Error
}

// List all role-permission mappings
func (r *rolePermissionRepository) ListAll(ctx context.Context) ([]*models.RolePermission, error) {
	var rolePermissions []*models.RolePermission
	err := r.db.WithContext(ctx).Find(&rolePermissions).Error
	return rolePermissions, err
}

// CreateWithValidation - SECURE: Creates role permission with organization validation
func (r *rolePermissionRepository) CreateWithValidation(ctx context.Context, rolePermission *models.RolePermission) error {
	// Validate that role and permission belong to same organization or permission is system-wide
	var role models.Role
	if err := r.db.WithContext(ctx).First(&role, "id = ?", rolePermission.RoleID).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	var permission models.Permission
	if err := r.db.WithContext(ctx).First(&permission, "id = ?", rolePermission.PermissionID).Error; err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// CRITICAL SECURITY CHECK: Custom permissions must be assigned only within same organization
	if permission.OrganizationID != nil {
		if permission.OrganizationID.String() != role.OrganizationID.String() {
			return fmt.Errorf("cannot assign custom permission from organization %s to role in organization %s",
				permission.OrganizationID.String(), role.OrganizationID.String())
		}
	} // System permissions (OrganizationID = nil) can be assigned to any role
	return r.db.WithContext(ctx).Create(rolePermission).Error
}

// GetByRoleWithOrgValidation - SECURE: Gets role permissions with organization validation
func (r *rolePermissionRepository) GetByRoleWithOrgValidation(ctx context.Context, roleID, orgID uuid.UUID) ([]*models.RolePermission, error) {
	var rolePermissions []*models.RolePermission
	err := r.db.WithContext(ctx).
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND roles.organization_id = ?", roleID, orgID).
		Where("(permissions.organization_id = ? OR permissions.organization_id IS NULL)", orgID).
		Find(&rolePermissions).Error
	return rolePermissions, err
}
