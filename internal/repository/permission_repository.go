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
	ListCustomForOrganization(ctx context.Context, orgID string) ([]*models.Permission, error)
	ListByCategory(ctx context.Context, category string) ([]*models.Permission, error)
	ListByCategoryAndOrganization(ctx context.Context, category, orgID string) ([]*models.Permission, error)

	// New methods for OAuth2 RBAC integration
	ListSystemPermissions(ctx context.Context) ([]*models.Permission, error)
	ListPermissionsByOrganization(ctx context.Context, orgID uuid.UUID) ([]*models.Permission, error)
	ListPermissionsForRole(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error)

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

// ListCustomForOrganization returns only custom permissions created by the organization (no system permissions)
func (r *permissionRepository) ListCustomForOrganization(ctx context.Context, orgID string) ([]*models.Permission, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var perms []*models.Permission
	err = r.db.WithContext(ctx).
		Where("organization_id = ? AND is_system = ?", orgUUID, false).
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
// CRITICAL: This enforces OAuth2 RBAC integration rules:
// 1. System permissions (is_system=true) can ONLY be assigned to system roles (role.is_system=true)
// 2. Custom permissions (is_system=false) can ONLY be assigned to custom roles in same org
// 3. Custom permissions CANNOT be assigned to system roles
// 4. Permissions from different orgs CANNOT be cross-assigned
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

	// CRITICAL SECURITY CHECK 1: System permissions can ONLY be assigned to system roles
	if permission.IsSystem && !role.IsSystem {
		return fmt.Errorf("SECURITY VIOLATION: Cannot assign system permission '%s' to custom role '%s'. System permissions can only be assigned to system roles",
			permission.Name, role.Name)
	}

	// CRITICAL SECURITY CHECK 2: Custom permissions CANNOT be assigned to system roles
	if !permission.IsSystem && role.IsSystem {
		return fmt.Errorf("SECURITY VIOLATION: Cannot assign custom permission '%s' to system role '%s'. Custom permissions can only be assigned to custom roles",
			permission.Name, role.Name)
	}

	// CRITICAL SECURITY CHECK 3: Custom permissions must belong to same organization as role
	if !permission.IsSystem {
		if permission.OrganizationID == nil {
			return fmt.Errorf("SECURITY VIOLATION: Custom permission '%s' has no organization but is not marked as system", permission.Name)
		}
		if role.OrganizationID == nil {
			return fmt.Errorf("SECURITY VIOLATION: Custom role '%s' has no organization but is not marked as system", role.Name)
		}
		if *permission.OrganizationID != *role.OrganizationID {
			return fmt.Errorf("SECURITY VIOLATION: Cannot assign permission '%s' (org: %s) to role '%s' (org: %s) from different organization",
				permission.Name, permission.OrganizationID.String(), role.Name, role.OrganizationID.String())
		}
	}

	// All security checks passed - create the assignment
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

//
// ─────────────────────────────────────────────
//   OAUTH2 RBAC INTEGRATION METHODS
// ─────────────────────────────────────────────
//

// ListSystemPermissions returns ONLY system permissions (is_system=true, organization_id IS NULL)
// Used for superadmin OAuth2 tokens
func (r *permissionRepository) ListSystemPermissions(ctx context.Context) ([]*models.Permission, error) {
	var perms []*models.Permission
	err := r.db.WithContext(ctx).
		Where("is_system = ? AND organization_id IS NULL", true).
		Order("category, name").
		Find(&perms).Error

	if err != nil {
		return nil, err
	}

	return perms, nil
}

// ListPermissionsByOrganization returns ONLY custom permissions for a specific organization
// (is_system=false, organization_id=orgID)
// Does NOT include system permissions - use this for org member OAuth2 tokens
func (r *permissionRepository) ListPermissionsByOrganization(ctx context.Context, orgID uuid.UUID) ([]*models.Permission, error) {
	var perms []*models.Permission
	err := r.db.WithContext(ctx).
		Where("is_system = ? AND organization_id = ?", false, orgID).
		Order("category, name").
		Find(&perms).Error

	if err != nil {
		return nil, err
	}

	return perms, nil
}

// ListPermissionsForRole returns all permissions assigned to a specific role
// Used to build permission list for OAuth2 tokens based on user's roles
func (r *permissionRepository) ListPermissionsForRole(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
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
