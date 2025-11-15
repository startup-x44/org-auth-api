package repository

import (
	"context"
	"fmt"

	"auth-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleRepository defines the interface for role operations
type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	GetByID(ctx context.Context, id string) (*models.Role, error)
	GetByOrganizationAndName(ctx context.Context, orgID, name string) (*models.Role, error)
	GetByOrganization(ctx context.Context, orgID string) ([]*models.Role, error)
	Update(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, id string) error
	CountMembersByRole(ctx context.Context, roleID string) (int64, error)
}

// roleRepository implements RoleRepository
type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

// Create creates a new role
func (r *roleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// GetByID retrieves a role by ID with permissions preloaded
func (r *roleRepository) GetByID(ctx context.Context, id string) (*models.Role, error) {
	roleID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	var role models.Role
	err = r.db.WithContext(ctx).
		Preload("Permissions").
		Where("id = ?", roleID).
		First(&role).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}

// GetByOrganizationAndName retrieves a role by organization and name
func (r *roleRepository) GetByOrganizationAndName(ctx context.Context, orgID, name string) (*models.Role, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var role models.Role
	err = r.db.WithContext(ctx).
		Preload("Permissions").
		Where("organization_id = ? AND name = ?", orgUUID, name).
		First(&role).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}

// GetByOrganization retrieves all roles for an organization
func (r *roleRepository) GetByOrganization(ctx context.Context, orgID string) ([]*models.Role, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var roles []*models.Role
	err = r.db.WithContext(ctx).
		Preload("Permissions").
		Where("organization_id = ?", orgUUID).
		Order("is_system DESC, name ASC"). // System roles first
		Find(&roles).Error

	if err != nil {
		return nil, err
	}

	return roles, nil
}

// Update updates a role
func (r *roleRepository) Update(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete deletes a role
func (r *roleRepository) Delete(ctx context.Context, id string) error {
	roleID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	return r.db.WithContext(ctx).Delete(&models.Role{}, "id = ?", roleID).Error
}

// CountMembersByRole counts how many members have this role
func (r *roleRepository) CountMembersByRole(ctx context.Context, roleID string) (int64, error) {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return 0, fmt.Errorf("invalid role ID: %w", err)
	}

	var count int64
	err = r.db.WithContext(ctx).
		Model(&models.OrganizationMembership{}).
		Where("role_id = ?", roleUUID).
		Count(&count).Error

	return count, err
}
