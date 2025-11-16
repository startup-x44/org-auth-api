package repository

import (
	"context"
	"fmt"

	"auth-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	GetByID(ctx context.Context, id string) (*models.Role, error) // DEPRECATED: Use GetByIDAndOrganization for security
	GetByIDAndOrganization(ctx context.Context, id, orgID string) (*models.Role, error)
	GetByOrganizationAndName(ctx context.Context, orgID, name string) (*models.Role, error)
	GetByOrganization(ctx context.Context, orgID string) ([]*models.Role, error)
	Update(ctx context.Context, role *models.Role) error
	DeleteByID(ctx context.Context, id string) error // DEPRECATED: Use DeleteByIDAndOrganization for security
	DeleteByIDAndOrganization(ctx context.Context, id, orgID string) error
	CountMembersByRole(ctx context.Context, roleID string) (int64, error) // DEPRECATED: Use CountMembersByRoleAndOrganization for security
	CountMembersByRoleAndOrganization(ctx context.Context, roleID, orgID string) (int64, error)
	GetRolesByPermission(ctx context.Context, permissionName string) ([]*models.Role, error) // DEPRECATED: Use GetRolesByPermissionAndOrganization for security
	GetRolesByPermissionAndOrganization(ctx context.Context, permissionName, orgID string) ([]*models.Role, error)

	// System role management (superadmin only)
	GetAllSystemRoles(ctx context.Context) ([]*models.Role, error)
	GetSystemRoleByName(ctx context.Context, name string) (*models.Role, error)

	// New methods for OAuth2 RBAC integration
	ListSystemRoles(ctx context.Context) ([]*models.Role, error)
	ListRolesByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]*models.Role, error)

	// Combined queries for superadmin
	GetAllRoles(ctx context.Context, includeSystem bool) ([]*models.Role, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

// Create a new role
func (r *roleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// Secure: load role scoped to organization ONLY
func (r *roleRepository) GetByIDAndOrganization(ctx context.Context, id, orgID string) (*models.Role, error) {
	roleID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var role models.Role
	err = r.db.WithContext(ctx).
		Preload("Permissions", "organization_id = ? OR (is_system = TRUE AND organization_id IS NULL)", orgUUID).
		Where("id = ? AND organization_id = ?", roleID, orgUUID).
		First(&role).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}

// Secure: get role by name + org
func (r *roleRepository) GetByOrganizationAndName(ctx context.Context, orgID, name string) (*models.Role, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var role models.Role
	err = r.db.WithContext(ctx).
		Preload("Permissions", "organization_id = ? OR (is_system = TRUE AND organization_id IS NULL)", orgUUID).
		Where("organization_id = ? AND name = ?", orgUUID, name).
		First(&role).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}

// Secure: roles only for this org
func (r *roleRepository) GetByOrganization(ctx context.Context, orgID string) ([]*models.Role, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var roles []*models.Role
	err = r.db.WithContext(ctx).
		Preload("Permissions", "organization_id = ? OR (is_system = TRUE AND organization_id IS NULL)", orgUUID).
		Where("organization_id = ?", orgUUID).
		Order("is_system DESC, name ASC").
		Find(&roles).Error

	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *roleRepository) Update(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// DEPRECATED: DeleteByID deletes a role without organization validation (UNSAFE)
// WARNING: This method may allow cross-organization role deletion
func (r *roleRepository) DeleteByID(ctx context.Context, id string) error {
	roleUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	return r.db.WithContext(ctx).
		Where("id = ?", roleUUID).
		Delete(&models.Role{}).Error
}

// Secure delete: requires correct org context
func (r *roleRepository) DeleteByIDAndOrganization(ctx context.Context, id, orgID string) error {
	roleUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization ID: %w", err)
	}

	return r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", roleUUID, orgUUID).
		Delete(&models.Role{}).Error
}

// Count members inside org only
func (r *roleRepository) CountMembersByRoleAndOrganization(ctx context.Context, roleID, orgID string) (int64, error) {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return 0, fmt.Errorf("invalid role ID: %w", err)
	}
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return 0, fmt.Errorf("invalid organization ID: %w", err)
	}

	var count int64
	err = r.db.WithContext(ctx).
		Model(&models.OrganizationMembership{}).
		Where("role_id = ? AND organization_id = ?", roleUUID, orgUUID).
		Count(&count).Error

	return count, err
}

// Secure: permission+org scope
func (r *roleRepository) GetRolesByPermissionAndOrganization(ctx context.Context, permissionName, orgID string) ([]*models.Role, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var roles []*models.Role

	err = r.db.WithContext(ctx).
		Joins("INNER JOIN role_permissions ON role_permissions.role_id = roles.id").
		Joins("INNER JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("permissions.name = ? AND roles.organization_id = ?", permissionName, orgUUID).
		Find(&roles).Error

	if err != nil {
		return nil, err
	}

	return roles, nil
}

// DEPRECATED: GetByID retrieves a role by ID without organization validation (UNSAFE)
// WARNING: This method does not validate organization context and may expose roles from other organizations
func (r *roleRepository) GetByID(ctx context.Context, id string) (*models.Role, error) {
	roleID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	var role models.Role
	err = r.db.WithContext(ctx).
		Where("id = ?", roleID).
		First(&role).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}

// DEPRECATED: CountMembersByRole counts members without organization validation (UNSAFE)
// WARNING: This method may count members across organizations
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

// DEPRECATED: GetRolesByPermission gets roles by permission without organization validation (UNSAFE)
// WARNING: This method may return roles from other organizations
func (r *roleRepository) GetRolesByPermission(ctx context.Context, permissionName string) ([]*models.Role, error) {
	var roles []*models.Role

	err := r.db.WithContext(ctx).
		Joins("INNER JOIN role_permissions ON role_permissions.role_id = roles.id").
		Joins("INNER JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("permissions.name = ?", permissionName).
		Find(&roles).Error

	if err != nil {
		return nil, err
	}

	return roles, nil
}

// GetAllSystemRoles retrieves all system roles (superadmin only)
// System roles have IsSystem=true and organization_id=NULL
func (r *roleRepository) GetAllSystemRoles(ctx context.Context) ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions", "is_system = TRUE AND organization_id IS NULL").
		Where("is_system = ? AND organization_id IS NULL", true).
		Order("name ASC").
		Find(&roles).Error

	if err != nil {
		return nil, err
	}

	return roles, nil
}

// GetSystemRoleByName retrieves a system role by name (superadmin only)
func (r *roleRepository) GetSystemRoleByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions", "is_system = TRUE AND organization_id IS NULL").
		Where("name = ? AND is_system = ? AND organization_id IS NULL", name, true).
		First(&role).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}

// GetAllRoles retrieves all roles (system + organization) - for superadmin
// If includeSystem is false, only returns custom organization roles
func (r *roleRepository) GetAllRoles(ctx context.Context, includeSystem bool) ([]*models.Role, error) {
	var roles []*models.Role
	query := r.db.WithContext(ctx).Preload("Permissions").Preload("Organization")

	if !includeSystem {
		query = query.Where("is_system = ?", false)
	}

	err := query.Order("is_system DESC, name ASC").Find(&roles).Error
	if err != nil {
		return nil, err
	}

	return roles, nil
}

//
// ─────────────────────────────────────────────
//   OAUTH2 RBAC INTEGRATION METHODS
// ─────────────────────────────────────────────
//

// ListSystemRoles returns ONLY system roles (is_system=true, organization_id IS NULL)
// Used for superadmin role loading in OAuth2 tokens
func (r *roleRepository) ListSystemRoles(ctx context.Context) ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Where("is_system = ? AND organization_id IS NULL", true).
		Order("name").
		Find(&roles).Error

	if err != nil {
		return nil, err
	}

	return roles, nil
}

// ListRolesByOrganizationID returns ONLY custom roles for a specific organization
// (is_system=false, organization_id=orgID)
// Does NOT include system roles - use this for org member OAuth2 tokens
func (r *roleRepository) ListRolesByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions", "is_system = ? AND organization_id = ?", false, orgID).
		Where("is_system = ? AND organization_id = ?", false, orgID).
		Order("name").
		Find(&roles).Error

	if err != nil {
		return nil, err
	}

	return roles, nil
}
