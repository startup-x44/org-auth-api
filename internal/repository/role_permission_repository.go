package repository

import (
	"context"

	"auth-service/internal/models"

	"gorm.io/gorm"
)

// RolePermissionRepository defines the interface for role permission repository
type RolePermissionRepository interface {
	GetByRole(ctx context.Context, role string) ([]*models.RolePermission, error)
	HasPermission(ctx context.Context, role, permission string) (bool, error)
	Create(ctx context.Context, rolePermission *models.RolePermission) error
	Delete(ctx context.Context, role, permission string) error
	ListAll(ctx context.Context) ([]*models.RolePermission, error)
}

// rolePermissionRepository implements RolePermissionRepository
type rolePermissionRepository struct {
	db *gorm.DB
}

// NewRolePermissionRepository creates a new role permission repository
func NewRolePermissionRepository(db *gorm.DB) RolePermissionRepository {
	return &rolePermissionRepository{db: db}
}

// GetByRole gets all permissions for a role
func (r *rolePermissionRepository) GetByRole(ctx context.Context, role string) ([]*models.RolePermission, error) {
	var permissions []*models.RolePermission
	err := r.db.WithContext(ctx).
		Where("role = ?", role).
		Find(&permissions).Error
	return permissions, err
}

// HasPermission checks if a role has a specific permission
func (r *rolePermissionRepository) HasPermission(ctx context.Context, role, permission string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.RolePermission{}).
		Where("role = ? AND permission = ?", role, permission).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Create creates a new role permission
func (r *rolePermissionRepository) Create(ctx context.Context, rolePermission *models.RolePermission) error {
	return r.db.WithContext(ctx).Create(rolePermission).Error
}

// Delete deletes a role permission
func (r *rolePermissionRepository) Delete(ctx context.Context, role, permission string) error {
	return r.db.WithContext(ctx).
		Where("role = ? AND permission = ?", role, permission).
		Delete(&models.RolePermission{}).Error
}

// ListAll lists all role permissions
func (r *rolePermissionRepository) ListAll(ctx context.Context) ([]*models.RolePermission, error) {
	var permissions []*models.RolePermission
	err := r.db.WithContext(ctx).
		Order("role, permission").
		Find(&permissions).Error
	return permissions, err
}
