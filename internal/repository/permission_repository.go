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
	GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	GetByName(ctx context.Context, name string) (*models.Permission, error)
	GetByNameAndOrganization(ctx context.Context, name, orgID string) (*models.Permission, error)
	GetByNames(ctx context.Context, names []string) ([]*models.Permission, error)
	GetByNamesAndOrganization(ctx context.Context, names []string, orgID string) ([]*models.Permission, error)
	ListAll(ctx context.Context) ([]*models.Permission, error)
	ListAllForOrganization(ctx context.Context, orgID string) ([]*models.Permission, error)
	ListByCategory(ctx context.Context, category string) ([]*models.Permission, error)
	ListByCategoryAndOrganization(ctx context.Context, category, orgID string) ([]*models.Permission, error)

	Create(ctx context.Context, permission *models.Permission) (*models.Permission, error)
	Update(ctx context.Context, permission *models.Permission) (*models.Permission, error)
	Delete(ctx context.Context, id uuid.UUID) error

	AssignToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	RevokeFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error)
	HasPermission(ctx context.Context, roleID uuid.UUID, permissionName string) (bool, error)

	BelongsToOrganization(perm *models.Permission, orgID uuid.UUID) bool
}

// permissionRepository implements PermissionRepository
type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

//
// ─────────────────────────────────────────────
//   Permission Ownership Helper
// ─────────────────────────────────────────────
//

// Returns true if permission belongs to org OR is system-wide
func (r *permissionRepository) BelongsToOrganization(perm *models.Permission, orgID uuid.UUID) bool {
	if perm.IsSystem {
		return true
	}
	if perm.OrganizationID != nil && *perm.OrganizationID == orgID {
		return true
	}
	return false
}

//
// ─────────────────────────────────────────────
//   READ OPERATIONS
// ─────────────────────────────────────────────
//

// GetByID retrieves a permission by ID
func (r *permissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	var perm models.Permission
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&perm).Error

	if err != nil {
		return nil, err
	}

	return &perm, nil
}

// GetByName retrieves a permission by name (system OR any org)
// NOTE: This is global; use GetByNameAndOrganization for secure org lookups
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

// SECURE VERSION: Ensures name belongs to system OR the org ONLY
func (r *permissionRepository) GetByNameAndOrganization(ctx context.Context, name, orgID string) (*models.Permission, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var perm models.Permission
	err = r.db.WithContext(ctx).
		Where("name = ? AND (organization_id IS NULL OR organization_id = ?)", name, orgUUID).
		First(&perm).Error
	if err != nil {
		return nil, err
	}

	// Extra safety: ensure correct org ownership
	if !r.BelongsToOrganization(&perm, orgUUID) {
		return nil, gorm.ErrRecordNotFound
	}

	return &perm, nil
}

// GetByNames retrieves multiple permissions (unsafe, global)
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

// SECURE VERSION: multiple permissions, ensures org ownership + system visibility
func (r *permissionRepository) GetByNamesAndOrganization(ctx context.Context, names []string, orgID string) ([]*models.Permission, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var perms []*models.Permission
	err = r.db.WithContext(ctx).
		Where("(organization_id IS NULL OR organization_id = ?) AND name IN ?", orgUUID, names).
		Find(&perms).Error

	if err != nil {
		return nil, err
	}

	// STRICT: Ensure all requested names exist in org/system
	if len(perms) != len(names) {
		return nil, fmt.Errorf("one or more permissions do not exist in this organization")
	}

	// Extra validation for sanity
	for _, perm := range perms {
		if !r.BelongsToOrganization(perm, orgUUID) {
			return nil, fmt.Errorf("invalid permission access detected")
		}
	}

	return perms, nil
}

//
// ─────────────────────────────────────────────
//   LIST OPERATIONS
// ─────────────────────────────────────────────
//

// ListAll retrieves all permissions (system + ALL orgs) — unsafe unless only called internally
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

// ListForOrganization: system-wide + org-specific only
func (r *permissionRepository) ListAllForOrganization(ctx context.Context, orgID string) ([]*models.Permission, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var perms []*models.Permission
	err = r.db.WithContext(ctx).
		Where("organization_id IS NULL OR organization_id = ?", orgUUID).
		Order("category, name").
		Find(&perms).Error

	if err != nil {
		return nil, err
	}

	return perms, nil
}

func (r *permissionRepository) ListByCategory(ctx context.Context, category string) ([]*models.Permission, error) {
	var perms []*models.Permission
	err := r.db.WithContext(ctx).
		Where("category = ?", category).
		Order("name").
		Find(&perms).Error

	return perms, err
}

func (r *permissionRepository) ListByCategoryAndOrganization(ctx context.Context, category, orgID string) ([]*models.Permission, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var perms []*models.Permission
	err = r.db.WithContext(ctx).
		Where("category = ? AND (organization_id IS NULL OR organization_id = ?)", category, orgUUID).
		Order("name").
		Find(&perms).Error

	return perms, err
}

//
// ─────────────────────────────────────────────
//   ROLE PERMISSION OPERATIONS
// ─────────────────────────────────────────────
//

// AssignToRole assigns a permission to a role with STRICT ORGANIZATION VALIDATION
// CRITICAL: This enforces the rule that custom permissions can only be assigned within their organization
func (r *permissionRepository) AssignToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	// Load the role to get its organization
	var role models.Role
	if err := r.db.WithContext(ctx).Where("id = ?", roleID).First(&role).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Load the permission to validate organization ownership
	var permission models.Permission
	if err := r.db.WithContext(ctx).Where("id = ?", permissionID).First(&permission).Error; err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// CRITICAL SECURITY CHECK: Enforce permission assignment rule
	// Rule: permission.IsSystem == true OR permission.OrganizationID == role.OrganizationID
	if !permission.IsSystem {
		if permission.OrganizationID == nil || role.OrganizationID == nil || *permission.OrganizationID != *role.OrganizationID {
			return fmt.Errorf("SECURITY VIOLATION: Permission %s (org: %v) cannot be assigned to role %s (org: %v)",
				permission.Name, permission.OrganizationID, role.Name, role.OrganizationID)
		}
	}

	return r.db.WithContext(ctx).Create(&models.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}).Error
}

func (r *permissionRepository) RevokeFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&models.RolePermission{}).Error
}

// GetRolePermissions retrieves all permissions for a role with ORGANIZATION-AWARE FILTERING
// SECURITY: Only returns system permissions + permissions from the role's organization
func (r *permissionRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
	// First get the role to determine its organization
	var role models.Role
	if err := r.db.WithContext(ctx).Where("id = ?", roleID).First(&role).Error; err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	var perms []*models.Permission
	err := r.db.WithContext(ctx).
		Joins("INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ? AND (permissions.is_system = true OR permissions.organization_id = ? OR permissions.organization_id IS NULL)",
			roleID, role.OrganizationID).
		Order("category, name").
		Find(&perms).Error

	return perms, err
}

func (r *permissionRepository) HasPermission(ctx context.Context, roleID uuid.UUID, permissionName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("role_permissions").
		Joins("INNER JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.name = ?", roleID, permissionName).
		Count(&count).Error

	return count > 0, err
}

//
// ─────────────────────────────────────────────
//   CRUD
// ─────────────────────────────────────────────
//

func (r *permissionRepository) Create(ctx context.Context, permission *models.Permission) (*models.Permission, error) {
	err := r.db.WithContext(ctx).Create(permission).Error
	return permission, err
}

func (r *permissionRepository) Update(ctx context.Context, permission *models.Permission) (*models.Permission, error) {
	err := r.db.WithContext(ctx).Save(permission).Error
	return permission, err
}

func (r *permissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.Permission{}).Error
}
