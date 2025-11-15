package repository

import (
	"context"
	"fmt"

	"auth-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionRepository defines the interface for permission operations
type PermissionRepository interface {
	GetByID(ctx context.Context, id string) (*models.Permission, error)
	GetByName(ctx context.Context, name string) (*models.Permission, error)
	GetByNames(ctx context.Context, names []string) ([]*models.Permission, error)
	ListAll(ctx context.Context) ([]*models.Permission, error)
	ListByCategory(ctx context.Context, category string) ([]*models.Permission, error)

	// Role-Permission assignments
	AssignToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	RevokeFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error)
	HasPermission(ctx context.Context, roleID uuid.UUID, permissionName string) (bool, error)
}

// permissionRepository implements PermissionRepository
type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

// GetByID retrieves a permission by ID
func (r *permissionRepository) GetByID(ctx context.Context, id string) (*models.Permission, error) {
	permID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid permission ID: %w", err)
	}

	var perm models.Permission
	err = r.db.WithContext(ctx).
		Where("id = ?", permID).
		First(&perm).Error

	if err != nil {
		return nil, err
	}

	return &perm, nil
}

// GetByName retrieves a permission by name
func (r *permissionRepository) GetByName(ctx context.Context, name string) (*models.Permission, error) {
	var perm models.Permission
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&perm).Error

	if err != nil {
		return nil, err
	}

	return &perm, nil
}

// GetByNames retrieves multiple permissions by names
func (r *permissionRepository) GetByNames(ctx context.Context, names []string) ([]*models.Permission, error) {
	var perms []*models.Permission
	err := r.db.WithContext(ctx).
		Where("name IN ?", names).
		Find(&perms).Error

	if err != nil {
		return nil, err
	}

	return perms, nil
}

// ListAll retrieves all permissions
func (r *permissionRepository) ListAll(ctx context.Context) ([]*models.Permission, error) {
	var perms []*models.Permission
	err := r.db.WithContext(ctx).
		Order("category, name").
		Find(&perms).Error

	if err != nil {
		return nil, err
	}

	return perms, nil
}

// ListByCategory retrieves permissions by category
func (r *permissionRepository) ListByCategory(ctx context.Context, category string) ([]*models.Permission, error) {
	var perms []*models.Permission
	err := r.db.WithContext(ctx).
		Where("category = ?", category).
		Order("name").
		Find(&perms).Error

	if err != nil {
		return nil, err
	}

	return perms, nil
}

// AssignToRole assigns a permission to a role
func (r *permissionRepository) AssignToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	rolePerm := &models.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}

	return r.db.WithContext(ctx).Create(rolePerm).Error
}

// RevokeFromRole revokes a permission from a role
func (r *permissionRepository) RevokeFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&models.RolePermission{}).Error
}

// GetRolePermissions retrieves all permissions for a role
func (r *permissionRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
	var perms []*models.Permission
	err := r.db.WithContext(ctx).
		Joins("INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Order("category, name").
		Find(&perms).Error

	if err != nil {
		return nil, err
	}

	return perms, nil
}

// HasPermission checks if a role has a specific permission
func (r *permissionRepository) HasPermission(ctx context.Context, roleID uuid.UUID, permissionName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("role_permissions").
		Joins("INNER JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.name = ?", roleID, permissionName).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
