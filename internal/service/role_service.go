package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/logger"

	"github.com/google/uuid"
)

// RoleService defines the interface for role operations
type RoleService interface {
	// Permission checking
	HasPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID, orgID uuid.UUID) ([]string, error)

	// Role CRUD
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*RoleResponse, error)
	GetRole(ctx context.Context, roleID uuid.UUID) (*RoleResponse, error)
	GetRoleWithOrganization(ctx context.Context, roleID, orgID uuid.UUID) (*RoleResponse, error)
	GetRolesByOrganization(ctx context.Context, orgID uuid.UUID) ([]*RoleResponse, error)
	ListSystemRoles(ctx context.Context) ([]*RoleResponse, error)
	UpdateRole(ctx context.Context, roleID uuid.UUID, req *UpdateRoleRequest) (*RoleResponse, error)
	UpdateRoleWithOrganization(ctx context.Context, roleID, orgID uuid.UUID, req *UpdateRoleRequest) (*RoleResponse, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) error
	DeleteRoleWithOrganization(ctx context.Context, roleID, orgID uuid.UUID) error

	// Permission assignment (DEPRECATED methods removed for security - use organization-scoped versions)
	// AssignPermissionsToRole - REMOVED: Use AssignPermissionsToRoleWithOrganization
	// RevokePermissionsFromRole - REMOVED: Use RevokePermissionsFromRoleWithOrganization
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]string, error)
	// Organization-scoped permission assignment (secures custom permission access)
	AssignPermissionsToRoleWithOrganization(ctx context.Context, roleID, orgID uuid.UUID, permissionNames []string) error
	RevokePermissionsFromRoleWithOrganization(ctx context.Context, roleID, orgID uuid.UUID, permissionNames []string) error
	GetRolePermissionsWithOrganization(ctx context.Context, roleID, orgID uuid.UUID) ([]string, error)

	// System permissions
	ListAllPermissions(ctx context.Context) ([]*PermissionResponse, error)
	// Organization-scoped permissions (includes system + org-specific)
	ListAllPermissionsForOrganization(ctx context.Context, orgID uuid.UUID) ([]*PermissionResponse, error)
	// Organization custom permissions only (for management, excludes system permissions)
	ListCustomPermissionsForOrganization(ctx context.Context, orgID uuid.UUID) ([]*PermissionResponse, error)

	// Permission CRUD
	CreatePermission(ctx context.Context, name, displayName, description, category string) (*PermissionResponse, error)
	// Organization-scoped permission creation (creates custom permissions tied to org)
	CreatePermissionForOrganization(ctx context.Context, orgID uuid.UUID, name, displayName, description, category string) (*PermissionResponse, error)
	UpdatePermission(ctx context.Context, permissionID, displayName, description, category string) (*PermissionResponse, error) // DEPRECATED: Use UpdatePermissionWithOrganization for security
	// Organization-scoped permission update (validates permission belongs to organization)
	UpdatePermissionWithOrganization(ctx context.Context, permissionID string, orgID uuid.UUID, displayName, description, category string) (*PermissionResponse, error)
	DeletePermission(ctx context.Context, permissionID string) error // DEPRECATED: Use DeletePermissionWithOrganization for security
	// Organization-scoped permission deletion (validates permission belongs to organization)
	DeletePermissionWithOrganization(ctx context.Context, permissionID string, orgID uuid.UUID) error
}

// roleService implements RoleService
type roleService struct {
	repo        repository.Repository
	auditLogger *logger.AuditLogger
}

// NewRoleService creates a new role service
func NewRoleService(repo repository.Repository, auditLogger *logger.AuditLogger) RoleService {
	return &roleService{
		repo:        repo,
		auditLogger: auditLogger,
	}
}

// Request/Response types
type CreateRoleRequest struct {
	OrganizationID uuid.UUID `json:"organization_id,omitempty"` // Set by handler from URL, not required in JSON
	Name           string    `json:"name" binding:"required"`
	DisplayName    string    `json:"display_name" binding:"required"`
	Description    string    `json:"description"`
	Permissions    []string  `json:"permissions"` // Permission names to assign
}

type UpdateRoleRequest struct {
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"` // If provided, replace all permissions
}

type RoleResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID *uuid.UUID `json:"organization_id"` // nil for system roles
	Name           string     `json:"name"`
	DisplayName    string     `json:"display_name"`
	Description    string     `json:"description"`
	IsSystem       bool       `json:"is_system"`
	Permissions    []string   `json:"permissions"`
	MemberCount    int        `json:"member_count,omitempty"`
}

type PermissionResponse struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	DisplayName    string     `json:"display_name"`
	Description    string     `json:"description"`
	Category       string     `json:"category"`
	IsSystem       bool       `json:"is_system"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"` // nil for system permissions
}

// HasPermission checks if a user has a specific permission in an organization
func (s *roleService) HasPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) (bool, error) {
	// Get user's membership
	membership, err := s.repo.OrganizationMembership().GetByOrganizationAndUser(ctx, orgID.String(), userID.String())
	if err != nil {
		return false, ErrMembershipNotFound
	}

	if membership.Status != models.MembershipStatusActive {
		return false, ErrMembershipSuspended
	}

	// Load role with organization validation to prevent cross-org access
	role, err := s.repo.Role().GetByIDAndOrganization(ctx, membership.RoleID.String(), orgID.String())
	if err != nil {
		return false, fmt.Errorf("failed to load role or role not in organization: %w", err)
	}

	// If admin role, allow all permissions
	if role.Name == models.RoleNameAdmin && role.IsSystem {
		return true, nil
	}

	// Check specific permission
	hasPermission, err := s.repo.Permission().HasPermission(ctx, membership.RoleID, permission)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return hasPermission, nil
}

// GetUserPermissions returns all permissions for a user in an organization
func (s *roleService) GetUserPermissions(ctx context.Context, userID, orgID uuid.UUID) ([]string, error) {
	// Get user's membership
	membership, err := s.repo.OrganizationMembership().GetByOrganizationAndUser(ctx, orgID.String(), userID.String())
	if err != nil {
		return nil, fmt.Errorf("membership not found: %w", err)
	}

	if membership.Status != models.MembershipStatusActive {
		return nil, errors.New("membership is not active")
	}

	// Load role with organization validation to prevent cross-org access
	role, err := s.repo.Role().GetByIDAndOrganization(ctx, membership.RoleID.String(), orgID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to load role or role not in organization: %w", err)
	}

	// If admin role, return all permissions
	if role.Name == models.RoleNameAdmin && role.IsSystem {
		return models.DefaultAdminPermissions(), nil
	}

	// Get role permissions
	rolePerms, err := s.repo.Permission().GetRolePermissions(ctx, membership.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	permissions := make([]string, len(rolePerms))
	for i, perm := range rolePerms {
		permissions[i] = perm.Name
	}

	return permissions, nil
}

// CreateRole creates a new custom role
func (s *roleService) CreateRole(ctx context.Context, req *CreateRoleRequest) (*RoleResponse, error) {
	userID := s.getUserID(ctx)

	// CRITICAL: Ensure IsSystem cannot be set to true by non-superadmins
	// Custom roles must have IsSystem=false
	// (Note: CreateRoleRequest should not even have IsSystem field to prevent this)

	// Check if role name already exists in organization (only check custom roles, not system roles)
	// Allow custom roles with same names as system roles (e.g., custom "owner" role in org)
	// This mirrors the permission model: custom permissions can have same names as system permissions
	existingRole, err := s.repo.Role().GetByOrganizationAndName(ctx, req.OrganizationID.String(), req.Name)
	if err == nil && existingRole != nil && !existingRole.IsSystem {
		return nil, errors.New("role name already exists in organization")
	}

	// Create CUSTOM role (IsSystem=false, OrganizationID=required)
	role := &models.Role{
		OrganizationID: &req.OrganizationID,
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		IsSystem:       false, // ALWAYS false for user-created roles
		CreatedBy:      uuid.MustParse(userID),
	}

	if err := s.repo.Role().Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Assign permissions if provided (with org validation)
	if len(req.Permissions) > 0 {
		if err := s.AssignPermissionsToRoleWithOrganization(ctx, role.ID, req.OrganizationID, req.Permissions); err != nil {
			return nil, fmt.Errorf("failed to assign permissions: %w", err)
		}
	}

	if s.auditLogger != nil {
		s.auditLogger.LogOrganizationAction(userID, "create_role", req.OrganizationID.String(), role.ID.String(), "", true, nil, fmt.Sprintf("Created custom role: %s", role.Name))
	}

	return s.GetRole(ctx, role.ID)
}

// GetRole retrieves a role by ID (DEPRECATED: Use GetRoleWithOrganization for security)
func (s *roleService) GetRole(ctx context.Context, roleID uuid.UUID) (*RoleResponse, error) {
	role, err := s.repo.Role().GetByID(ctx, roleID.String())
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	permissions, _ := s.GetRolePermissions(ctx, roleID)
	response := s.convertToRoleResponse(role)
	response.Permissions = permissions

	// Get member count
	count, _ := s.repo.Role().CountMembersByRole(ctx, roleID.String())
	response.MemberCount = int(count)

	return response, nil
}

// GetRoleWithOrganization retrieves a role by ID with organization filtering for security
func (s *roleService) GetRoleWithOrganization(ctx context.Context, roleID, orgID uuid.UUID) (*RoleResponse, error) {
	role, err := s.repo.Role().GetByIDAndOrganization(ctx, roleID.String(), orgID.String())
	if err != nil {
		return nil, fmt.Errorf("role not found in organization: %w", err)
	}

	permissions, _ := s.GetRolePermissions(ctx, roleID)
	response := s.convertToRoleResponse(role)
	response.Permissions = permissions

	// Get member count
	count, _ := s.repo.Role().CountMembersByRole(ctx, roleID.String())
	response.MemberCount = int(count)

	return response, nil
}

// GetRolesByOrganization retrieves all roles for an organization
func (s *roleService) GetRolesByOrganization(ctx context.Context, orgID uuid.UUID) ([]*RoleResponse, error) {
	// Check if user is superadmin to determine which roles to return
	isSuperAdmin, _ := ctx.Value("is_superadmin").(bool)

	roles, err := s.repo.Role().GetByOrganization(ctx, orgID.String(), isSuperAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	responses := make([]*RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = s.convertToRoleResponse(role)

		// Get permissions for each role using organization-scoped method
		permissions, permErr := s.GetRolePermissionsWithOrganization(ctx, role.ID, orgID)
		if permErr != nil {
			logger.Warn(context.Background()).Err(permErr).
				Str("role_id", role.ID.String()).
				Str("org_id", orgID.String()).
				Msg("Failed to get role permissions for role list")
		}
		responses[i].Permissions = permissions

		logger.Debug(context.Background()).
			Str("role_id", role.ID.String()).
			Str("role_name", role.Name).
			Strs("permissions", permissions).
			Int("permission_count", len(permissions)).
			Msg("Role permissions loaded for organization role list")

		// Get member count
		count, _ := s.repo.Role().CountMembersByRole(ctx, role.ID.String())
		responses[i].MemberCount = int(count)
	}

	return responses, nil
}

// ListSystemRoles retrieves only true system roles (organization_id IS NULL and is_system = true)
func (s *roleService) ListSystemRoles(ctx context.Context) ([]*RoleResponse, error) {
	roles, err := s.repo.Role().GetAllSystemRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system roles: %w", err)
	}

	responses := make([]*RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = s.convertToRoleResponse(role)

		// Get permissions for each role
		permissions, _ := s.GetRolePermissions(ctx, role.ID)
		responses[i].Permissions = permissions

		// Get member count
		count, _ := s.repo.Role().CountMembersByRole(ctx, role.ID.String())
		responses[i].MemberCount = int(count)
	}

	return responses, nil
}

// UpdateRole updates a role's details (DEPRECATED: Use UpdateRoleWithOrganization for security)
func (s *roleService) UpdateRole(ctx context.Context, roleID uuid.UUID, req *UpdateRoleRequest) (*RoleResponse, error) {
	userID := s.getUserID(ctx)

	role, err := s.repo.Role().GetByID(ctx, roleID.String())
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// Prevent updating system roles
	if role.IsSystem {
		return nil, errors.New("cannot update system role")
	}

	// Update fields only if provided (not empty)
	if req.DisplayName != "" {
		role.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	// Save role
	if err := s.repo.Role().Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// Update permissions if provided
	if req.Permissions != nil {
		// First revoke all existing permissions
		existingPerms, err := s.GetRolePermissions(ctx, roleID)
		if err != nil {
			// Log but don't fail - maybe no permissions exist yet
			fmt.Printf("Warning: failed to get existing permissions: %v\n", err)
		}

		if len(existingPerms) > 0 {
			// Ignore errors when revoking (permission might not exist)
			if err := s.RevokePermissionsFromRole(ctx, roleID, existingPerms); err != nil {
				fmt.Printf("Warning: failed to revoke permissions: %v\n", err)
			}
		}

		// Assign new permissions
		if len(req.Permissions) > 0 {
			if role.OrganizationID != nil {
				if err := s.AssignPermissionsToRoleWithOrganization(ctx, roleID, *role.OrganizationID, req.Permissions); err != nil {
					return nil, fmt.Errorf("failed to assign new permissions: %w", err)
				}
			}
		}
	}

	if s.auditLogger != nil {
		orgIDStr := ""
		if role.OrganizationID != nil {
			orgIDStr = role.OrganizationID.String()
		}
		s.auditLogger.LogOrganizationAction(userID, "update_role", orgIDStr, roleID.String(), "", true, nil, fmt.Sprintf("Updated role: %s", role.Name))
	}

	return s.GetRole(ctx, roleID)
}

// UpdateRoleWithOrganization updates a role's details with organization filtering for security
func (s *roleService) UpdateRoleWithOrganization(ctx context.Context, roleID, orgID uuid.UUID, req *UpdateRoleRequest) (*RoleResponse, error) {
	userID := s.getUserID(ctx)

	role, err := s.repo.Role().GetByIDAndOrganization(ctx, roleID.String(), orgID.String())
	if err != nil {
		return nil, fmt.Errorf("role not found in organization: %w", err)
	}

	// Prevent updating system roles
	if role.IsSystem {
		return nil, errors.New("cannot update system role")
	}

	// Update fields only if provided (not empty)
	if req.DisplayName != "" {
		role.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	// Save role
	if err := s.repo.Role().Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// Update permissions if provided - USING SECURE ORGANIZATION-SCOPED METHODS
	if req.Permissions != nil {
		// First get existing permissions with organization validation
		existingPerms, err := s.GetRolePermissionsWithOrganization(ctx, roleID, orgID)
		if err != nil {
			// Log but don't fail - maybe no permissions exist yet
			fmt.Printf("Warning: failed to get existing permissions: %v\n", err)
		}

		if len(existingPerms) > 0 {
			// Revoke existing permissions using organization-scoped method
			if err := s.RevokePermissionsFromRoleWithOrganization(ctx, roleID, orgID, existingPerms); err != nil {
				fmt.Printf("Warning: failed to revoke permissions: %v\n", err)
			}
		}

		// Assign new permissions using organization-scoped method with validation
		if len(req.Permissions) > 0 {
			if err := s.AssignPermissionsToRoleWithOrganization(ctx, roleID, orgID, req.Permissions); err != nil {
				return nil, fmt.Errorf("failed to assign new permissions: %w", err)
			}
		}
	}

	if s.auditLogger != nil {
		s.auditLogger.LogOrganizationAction(userID, "update_role", role.OrganizationID.String(), roleID.String(), "", true, nil, fmt.Sprintf("Updated role: %s", role.Name))
	}

	return s.GetRoleWithOrganization(ctx, roleID, orgID)
}

// DeleteRole deletes a custom role (DEPRECATED: Use DeleteRoleWithOrganization for security)
func (s *roleService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	userID := s.getUserID(ctx)

	role, err := s.repo.Role().GetByID(ctx, roleID.String())
	if err != nil {
		return ErrRoleNotFound
	}

	if role.IsSystem {
		return errors.New("cannot delete system role")
	}

	// Check if role is in use by any members
	memberCount, err := s.repo.Role().CountMembersByRole(ctx, roleID.String())
	if err != nil {
		return fmt.Errorf("failed to check role usage: %w", err)
	}

	if memberCount > 0 {
		return fmt.Errorf("cannot delete role with %d active members", memberCount)
	}

	// Delete role (cascade will delete role_permissions)
	if err := s.repo.Role().DeleteByID(ctx, roleID.String()); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	if s.auditLogger != nil {
		s.auditLogger.LogOrganizationAction(userID, "delete_role", role.OrganizationID.String(), roleID.String(), "", true, nil, fmt.Sprintf("Deleted role: %s", role.Name))
	}

	return nil
}

// DeleteRoleWithOrganization deletes a custom role with organization filtering for security
func (s *roleService) DeleteRoleWithOrganization(ctx context.Context, roleID, orgID uuid.UUID) error {
	userID := s.getUserID(ctx)

	role, err := s.repo.Role().GetByIDAndOrganization(ctx, roleID.String(), orgID.String())
	if err != nil {
		return fmt.Errorf("role not found in organization: %w", err)
	}

	// Prevent deleting system roles
	if role.IsSystem {
		return errors.New("cannot delete system role")
	}

	// Check if role is in use by any members
	memberCount, err := s.repo.Role().CountMembersByRole(ctx, roleID.String())
	if err != nil {
		return fmt.Errorf("failed to check role usage: %w", err)
	}

	if memberCount > 0 {
		return fmt.Errorf("cannot delete role with %d active members", memberCount)
	}

	// Delete role (cascade will delete role_permissions)
	if err := s.repo.Role().DeleteByIDAndOrganization(ctx, roleID.String(), orgID.String()); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	if s.auditLogger != nil {
		s.auditLogger.LogOrganizationAction(userID, "delete_role", role.OrganizationID.String(), roleID.String(), "", true, nil, fmt.Sprintf("Deleted role: %s", role.Name))
	}

	return nil
}

// AssignPermissionsToRole - DISABLED FOR SECURITY: Use AssignPermissionsToRoleWithOrganization instead
// This method was removed because it does not validate organization context and allows privilege escalation
func (s *roleService) AssignPermissionsToRole(ctx context.Context, roleID uuid.UUID, permissionNames []string) error {
	return errors.New("SECURITY ERROR: AssignPermissionsToRole is disabled - use AssignPermissionsToRoleWithOrganization for organization-safe permission assignment")
}

// RevokePermissionsFromRole - DISABLED FOR SECURITY: Use RevokePermissionsFromRoleWithOrganization instead
// This method was removed because it does not validate organization context and allows privilege escalation
func (s *roleService) RevokePermissionsFromRole(ctx context.Context, roleID uuid.UUID, permissionNames []string) error {
	return errors.New("SECURITY ERROR: RevokePermissionsFromRole is disabled - use RevokePermissionsFromRoleWithOrganization for organization-safe permission management")
}

// GetRolePermissions returns all permissions for a role (DEPRECATED: Use organization-scoped validation)
// WARNING: This method does not validate organization context and may expose cross-organization data
func (s *roleService) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	role, err := s.repo.Role().GetByID(ctx, roleID.String())
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// If system admin role, return all permissions
	if role.IsSystem && role.Name == models.RoleNameAdmin {
		return models.DefaultAdminPermissions(), nil
	}

	rolePerms, err := s.repo.Permission().GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	permissions := make([]string, len(rolePerms))
	for i, perm := range rolePerms {
		permissions[i] = perm.Name
	}

	return permissions, nil
}

// AssignPermissionsToRoleWithOrganization assigns permissions to a role with proper organization security
// This method ensures that custom permissions can only be assigned within their organization context
func (s *roleService) AssignPermissionsToRoleWithOrganization(ctx context.Context, roleID, orgID uuid.UUID, permissionNames []string) error {
	userID := s.getUserID(ctx)

	// Check if role exists and belongs to the organization
	role, err := s.repo.Role().GetByIDAndOrganization(ctx, roleID.String(), orgID.String())
	if err != nil {
		return ErrRoleNotFoundInOrg
	}

	// Prevent modifying system role permissions via API
	if role.IsSystem {
		return ErrCannotModifySystemPerms
	}

	// Get permissions by names with organization context (includes system + org-specific permissions)
	perms, err := s.repo.Permission().GetByNamesAndOrganization(ctx, permissionNames, orgID.String())
	if err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	if len(perms) != len(permissionNames) {
		return ErrSomePermissionsNotFound
	}

	// STRICT VALIDATION: Custom roles can ONLY have custom permissions
	// System roles can ONLY have system permissions (but this is blocked earlier)
	for _, perm := range perms {
		// RULE 1: Custom roles (is_system=false) can be assigned custom permissions OR system permissions
		// We removed the restriction that prevented system permissions from being assigned to custom roles

		// RULE 2: Custom permissions must belong to the same organization as the role
		if !perm.IsSystem && (perm.OrganizationID == nil || *perm.OrganizationID != orgID) {
			return fmt.Errorf("permission %s belongs to another organization and cannot be assigned", perm.Name)
		}
	}

	// Assign each validated permission
	for _, perm := range perms {
		if err := s.repo.Permission().AssignToRole(ctx, roleID, perm.ID); err != nil {
			return fmt.Errorf("failed to assign permission %s: %w", perm.Name, err)
		}
	}

	if s.auditLogger != nil {
		s.auditLogger.LogOrganizationAction(userID, "assign_permissions", orgID.String(), roleID.String(), "", true, nil, fmt.Sprintf("Assigned %d permissions to role %s", len(permissionNames), role.Name))
	}

	return nil
}

// RevokePermissionsFromRoleWithOrganization revokes permissions from a role with proper organization security
func (s *roleService) RevokePermissionsFromRoleWithOrganization(ctx context.Context, roleID, orgID uuid.UUID, permissionNames []string) error {
	userID := s.getUserID(ctx)

	// Check if role exists and belongs to the organization
	role, err := s.repo.Role().GetByIDAndOrganization(ctx, roleID.String(), orgID.String())
	if err != nil {
		return ErrRoleNotFoundInOrg
	}

	// Prevent modifying system role permissions
	if role.IsSystem {
		return ErrCannotModifySystemPerms
	}

	// Get permissions by names with organization context (ignore if some don't exist)
	perms, err := s.repo.Permission().GetByNamesAndOrganization(ctx, permissionNames, orgID.String())
	if err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	// If no permissions found, just return success (nothing to revoke)
	if len(perms) == 0 {
		return nil
	}

	// Revoke each permission with proper logging
	revokedCount := 0
	for _, perm := range perms {
		if err := s.repo.Permission().RevokeFromRole(ctx, roleID, perm.ID); err != nil {
			logger.Warn(ctx).Err(err).
				Str("role_id", roleID.String()).
				Str("permission_id", perm.ID.String()).
				Str("permission_name", perm.Name).
				Msg("Failed to revoke permission from role")
		} else {
			revokedCount++
			logger.Info(ctx).
				Str("role_id", roleID.String()).
				Str("permission_id", perm.ID.String()).
				Str("permission_name", perm.Name).
				Msg("Successfully revoked permission from role")
		}
	}

	logger.Info(ctx).
		Str("role_id", roleID.String()).
		Str("role_name", role.Name).
		Int("requested_count", len(permissionNames)).
		Int("found_count", len(perms)).
		Int("revoked_count", revokedCount).
		Msg("Permission revocation completed")

	if s.auditLogger != nil {
		s.auditLogger.LogOrganizationAction(userID, "revoke_permissions", orgID.String(), roleID.String(), "", true, nil, fmt.Sprintf("Revoked %d permissions from role %s", len(permissionNames), role.Name))
	}

	return nil
}

// GetRolePermissionsWithOrganization returns all permissions for a role with organization security validation
func (s *roleService) GetRolePermissionsWithOrganization(ctx context.Context, roleID, orgID uuid.UUID) ([]string, error) {
	// Validate role belongs to organization
	role, err := s.repo.Role().GetByIDAndOrganization(ctx, roleID.String(), orgID.String())
	if err != nil {
		return nil, ErrRoleNotFoundInOrg
	}

	// If system admin role, return all permissions
	if role.IsSystem && role.Name == models.RoleNameAdmin {
		return models.DefaultAdminPermissions(), nil
	}

	// Get role permissions with organization context validation
	rolePerms, err := s.repo.Permission().GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	// Validate each permission is accessible within this organization context
	permissions := make([]string, 0, len(rolePerms))
	for _, perm := range rolePerms {
		// Include system permissions (org_id IS NULL) or permissions that belong to this organization
		if perm.IsSystem || (perm.OrganizationID != nil && *perm.OrganizationID == orgID) {
			permissions = append(permissions, perm.Name)
		}
		// Skip permissions that belong to other organizations
	}

	return permissions, nil
}

// ListAllPermissions returns all system permissions
func (s *roleService) ListAllPermissions(ctx context.Context) ([]*PermissionResponse, error) {
	// Query from database
	perms, err := s.repo.Permission().ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	responses := make([]*PermissionResponse, len(perms))
	for i, perm := range perms {
		responses[i] = &PermissionResponse{
			ID:             perm.ID,
			Name:           perm.Name,
			DisplayName:    perm.DisplayName,
			Description:    perm.Description,
			Category:       perm.Category,
			IsSystem:       perm.IsSystem,
			OrganizationID: perm.OrganizationID,
		}
	}

	return responses, nil
}

// ListAllPermissionsForOrganization returns all permissions available to an organization
// This includes system permissions AND custom permissions created for this organization
func (s *roleService) ListAllPermissionsForOrganization(ctx context.Context, orgID uuid.UUID) ([]*PermissionResponse, error) {
	// Query permissions with organization context
	perms, err := s.repo.Permission().ListAllForOrganization(ctx, orgID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions for organization: %w", err)
	}

	responses := make([]*PermissionResponse, len(perms))
	for i, perm := range perms {
		responses[i] = &PermissionResponse{
			ID:             perm.ID,
			Name:           perm.Name,
			DisplayName:    perm.DisplayName,
			Description:    perm.Description,
			Category:       perm.Category,
			IsSystem:       perm.IsSystem,
			OrganizationID: perm.OrganizationID, // Will be nil for system permissions
		}
	}

	return responses, nil
}

// ListCustomPermissionsForOrganization returns only custom permissions created by the organization
// This excludes system permissions and is used for permission management (not role assignment)
func (s *roleService) ListCustomPermissionsForOrganization(ctx context.Context, orgID uuid.UUID) ([]*PermissionResponse, error) {
	// Query only custom permissions for this organization
	perms, err := s.repo.Permission().ListCustomForOrganization(ctx, orgID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list custom permissions for organization: %w", err)
	}

	responses := make([]*PermissionResponse, len(perms))
	for i, perm := range perms {
		responses[i] = &PermissionResponse{
			ID:             perm.ID,
			Name:           perm.Name,
			DisplayName:    perm.DisplayName,
			Description:    perm.Description,
			Category:       perm.Category,
			IsSystem:       perm.IsSystem, // Will always be false for custom permissions
			OrganizationID: perm.OrganizationID,
		}
	}

	return responses, nil
}

// CreatePermissionForOrganization creates a new custom permission tied to an organization
func (s *roleService) CreatePermissionForOrganization(ctx context.Context, orgID uuid.UUID, name, displayName, description, category string) (*PermissionResponse, error) {
	userID := s.getUserID(ctx)

	// Validate permission name format: action:resource (e.g., view:profile, edit:content)
	// Must contain exactly one colon, with non-empty parts before and after
	// Only lowercase letters and underscores allowed
	if !strings.Contains(name, ":") {
		return nil, errors.New("permission name must follow the format 'action:resource' (e.g., view:profile, edit:content)")
	}

	parts := strings.Split(name, ":")
	if len(parts) != 2 {
		return nil, errors.New("permission name must contain exactly one colon separating action and resource")
	}

	action, resource := parts[0], parts[1]
	if action == "" || resource == "" {
		return nil, errors.New("both action and resource parts must be non-empty in permission name")
	}

	// Validate characters: only lowercase letters and underscores
	validNameRegex := `^[a-z_]+$`
	matched, _ := regexp.MatchString(validNameRegex, action)
	if !matched {
		return nil, errors.New("action part must contain only lowercase letters and underscores")
	}
	matched, _ = regexp.MatchString(validNameRegex, resource)
	if !matched {
		return nil, errors.New("resource part must contain only lowercase letters and underscores")
	}

	// Check if a CUSTOM permission with this name already exists in THIS organization
	// System permissions (is_system=true) don't conflict - users can create custom permissions with same name
	existing, err := s.repo.Permission().GetByNameAndOrganization(ctx, name, orgID.String())
	if err == nil && existing != nil {
		// Only block if it's a custom permission in THIS org (not a system permission)
		if !existing.IsSystem && existing.OrganizationID != nil && *existing.OrganizationID == orgID {
			return nil, ErrPermissionAlreadyExists
		}
	}

	// Check if displayName is unique within the same category for this organization
	categoryPerms, err := s.repo.Permission().ListByCategoryAndOrganization(ctx, category, orgID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to check category permissions: %w", err)
	}

	for _, perm := range categoryPerms {
		if perm.DisplayName == displayName && perm.OrganizationID != nil && *perm.OrganizationID == orgID {
			return nil, fmt.Errorf("display name '%s' already exists in category '%s' for this organization", displayName, category)
		}
	}

	// Create the custom permission
	perm := &models.Permission{
		ID:             uuid.New(),
		Name:           name,
		DisplayName:    displayName,
		Description:    description,
		Category:       category,
		IsSystem:       false,  // Custom permissions are never system permissions
		OrganizationID: &orgID, // Tie to specific organization
	}

	createdPerm, err := s.repo.Permission().Create(ctx, perm)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	if s.auditLogger != nil {
		s.auditLogger.LogOrganizationAction(userID, "create_permission", orgID.String(), createdPerm.ID.String(), "", true, nil, fmt.Sprintf("Created custom permission: %s", name))
	}

	return &PermissionResponse{
		ID:             createdPerm.ID,
		Name:           createdPerm.Name,
		DisplayName:    createdPerm.DisplayName,
		Description:    createdPerm.Description,
		Category:       createdPerm.Category,
		IsSystem:       createdPerm.IsSystem,
		OrganizationID: createdPerm.OrganizationID,
	}, nil
}

// Helper methods
func (s *roleService) getUserID(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

func (s *roleService) convertToRoleResponse(role *models.Role) *RoleResponse {
	if role == nil {
		return nil
	}

	response := &RoleResponse{
		ID:             role.ID,
		OrganizationID: role.OrganizationID,
		Name:           role.Name,
		DisplayName:    role.DisplayName,
		Description:    role.Description,
		IsSystem:       role.IsSystem,
	}

	// Add permissions if loaded
	if len(role.Permissions) > 0 {
		permissions := make([]string, len(role.Permissions))
		for i, perm := range role.Permissions {
			permissions[i] = perm.Name
		}
		response.Permissions = permissions
	}

	return response
}

// CreatePermission creates a new custom permission
func (s *roleService) CreatePermission(ctx context.Context, name, displayName, description, category string) (*PermissionResponse, error) {
	// Normalize inputs to lowercase for consistent storage and comparison
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	normalizedCategory := strings.ToLower(strings.TrimSpace(category))
	trimmedDisplayName := strings.TrimSpace(displayName)

	// Format permission name as category:name
	formattedName := fmt.Sprintf("%s:%s", normalizedCategory, normalizedName)

	// Check if permission already exists by exact name
	existingPerm, err := s.repo.Permission().GetByName(ctx, formattedName)
	if err == nil && existingPerm != nil {
		return nil, fmt.Errorf("permission already exists. Please use a different name")
	}
	// Only ignore "not found" type errors, return other errors
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") && !strings.Contains(strings.ToLower(err.Error()), "record not found") {
		return nil, fmt.Errorf("failed to check existing permission: %w", err)
	}

	// Check if there's already a permission with the same display name in the same category (case-insensitive)
	categoryPerms, err := s.repo.Permission().ListByCategory(ctx, normalizedCategory)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing permissions: %w", err)
	}

	for _, perm := range categoryPerms {
		if strings.EqualFold(perm.DisplayName, trimmedDisplayName) {
			return nil, fmt.Errorf("a permission with this name already exists in this category. Please choose a different name")
		}
	}

	// Also check for similar permission names across all categories to prevent confusion
	allPerms, err := s.repo.Permission().ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing permissions: %w", err)
	}

	for _, perm := range allPerms {
		if strings.EqualFold(perm.DisplayName, trimmedDisplayName) {
			return nil, fmt.Errorf("a permission named '%s' already exists in the '%s' category. Please choose a different name", perm.DisplayName, perm.Category)
		}
	}

	permission := &models.Permission{
		Name:        formattedName,
		DisplayName: trimmedDisplayName,
		Description: description,
		Category:    normalizedCategory,
		IsSystem:    false, // Custom permissions are not system permissions
	}

	createdPermission, err := s.repo.Permission().Create(ctx, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	// TODO: Add audit logging when LogActivity method is available
	// if s.auditLogger != nil {
	//     userID := s.getUserID(ctx)
	//     s.auditLogger.LogActivity(userID, "permission:create", fmt.Sprintf("Created permission: %s", name))
	// }

	return &PermissionResponse{
		ID:          createdPermission.ID,
		Name:        createdPermission.Name,
		DisplayName: createdPermission.DisplayName,
		Description: createdPermission.Description,
		Category:    createdPermission.Category,
		IsSystem:    createdPermission.IsSystem,
	}, nil
}

// UpdatePermission updates a custom permission (DEPRECATED: Use UpdatePermissionWithOrganization for security)
// WARNING: This method does not validate organization context and may allow cross-organization permission modification
func (s *roleService) UpdatePermission(ctx context.Context, permissionID, displayName, description, category string) (*PermissionResponse, error) {
	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return nil, errors.New("invalid permission ID")
	}

	// Get existing permission
	permission, err := s.repo.Permission().GetByID(ctx, permissionUUID)
	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	// Check if it's a system permission
	if permission.IsSystem {
		return nil, errors.New("system permissions cannot be updated")
	}

	// If category is changing, check for conflicts
	if permission.Category != category {
		// Check if there's already a permission with the same display name in the new category
		categoryPerms, err := s.repo.Permission().ListByCategory(ctx, category)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing permissions: %w", err)
		}

		for _, perm := range categoryPerms {
			if perm.DisplayName == displayName && perm.ID != permission.ID {
				return nil, fmt.Errorf("a permission with display name '%s' already exists in category '%s'", displayName, category)
			}
		}
	} else {
		// Same category, just check for display name conflicts
		categoryPerms, err := s.repo.Permission().ListByCategory(ctx, category)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing permissions: %w", err)
		}

		for _, perm := range categoryPerms {
			if perm.DisplayName == displayName && perm.ID != permission.ID {
				return nil, fmt.Errorf("a permission with display name '%s' already exists in category '%s'", displayName, category)
			}
		}
	}

	// Update fields
	permission.DisplayName = displayName
	permission.Description = description
	permission.Category = category

	updatedPermission, err := s.repo.Permission().Update(ctx, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	// TODO: Add audit logging when LogActivity method is available
	// if s.auditLogger != nil {
	//     userID := s.getUserID(ctx)
	//     s.auditLogger.LogActivity(userID, "permission:update", fmt.Sprintf("Updated permission: %s", permission.Name))
	// }

	return &PermissionResponse{
		ID:          updatedPermission.ID,
		Name:        updatedPermission.Name,
		DisplayName: updatedPermission.DisplayName,
		Description: updatedPermission.Description,
		Category:    updatedPermission.Category,
		IsSystem:    updatedPermission.IsSystem,
	}, nil
}

// UpdatePermissionWithOrganization updates a custom permission with proper organization security validation
func (s *roleService) UpdatePermissionWithOrganization(ctx context.Context, permissionID string, orgID uuid.UUID, displayName, description, category string) (*PermissionResponse, error) {
	userID := s.getUserID(ctx)

	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	// Get existing permission
	permission, err := s.repo.Permission().GetByID(ctx, permissionUUID)
	if err != nil {
		return nil, ErrPermissionNotFound
	}

	// Check if it's a system permission (cannot be updated)
	if permission.IsSystem {
		return nil, ErrCannotUpdateSystemPerm
	}

	// CRITICAL SECURITY CHECK: Ensure permission belongs to the organization
	if permission.OrganizationID == nil || *permission.OrganizationID != orgID {
		return nil, ErrPermissionNotFound // Don't reveal it exists in another org
	}

	// If category is changing, check for conflicts within THIS organization only
	if permission.Category != category {
		categoryPerms, err := s.repo.Permission().ListByCategoryAndOrganization(ctx, category, orgID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to check existing permissions: %w", err)
		}

		for _, perm := range categoryPerms {
			if perm.DisplayName == displayName && perm.ID != permission.ID {
				return nil, fmt.Errorf("a permission with display name '%s' already exists in category '%s' for this organization", displayName, category)
			}
		}
	} else {
		// Same category, check for display name conflicts within THIS organization only
		categoryPerms, err := s.repo.Permission().ListByCategoryAndOrganization(ctx, category, orgID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to check existing permissions: %w", err)
		}

		for _, perm := range categoryPerms {
			if perm.DisplayName == displayName && perm.ID != permission.ID {
				return nil, fmt.Errorf("a permission with display name '%s' already exists in category '%s' for this organization", displayName, category)
			}
		}
	}

	// Update fields
	permission.DisplayName = displayName
	permission.Description = description
	permission.Category = category

	updatedPermission, err := s.repo.Permission().Update(ctx, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	if s.auditLogger != nil {
		s.auditLogger.LogOrganizationAction(userID, "update_permission", orgID.String(), permissionUUID.String(), "", true, nil, fmt.Sprintf("Updated custom permission: %s", permission.Name))
	}

	return &PermissionResponse{
		ID:             updatedPermission.ID,
		Name:           updatedPermission.Name,
		DisplayName:    updatedPermission.DisplayName,
		Description:    updatedPermission.Description,
		Category:       updatedPermission.Category,
		IsSystem:       updatedPermission.IsSystem,
		OrganizationID: updatedPermission.OrganizationID,
	}, nil
}

// DeletePermission deletes a custom permission (DEPRECATED: Use DeletePermissionWithOrganization for security)
// WARNING: This method uses global role lookup and may allow cross-organization permission deletion
func (s *roleService) DeletePermission(ctx context.Context, permissionID string) error {
	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return errors.New("invalid permission ID")
	}

	// Get existing permission
	permission, err := s.repo.Permission().GetByID(ctx, permissionUUID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// Check if it's a system permission
	if permission.IsSystem {
		return errors.New("system permissions cannot be deleted")
	}

	// Check if permission is used by any roles
	roles, err := s.repo.Role().GetRolesByPermission(ctx, permission.Name)
	if err != nil {
		return fmt.Errorf("failed to check permission usage: %w", err)
	}

	if len(roles) > 0 {
		return errors.New("permission is currently assigned to roles and cannot be deleted")
	}

	// Delete the permission
	err = s.repo.Permission().Delete(ctx, permissionUUID)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	// TODO: Add audit logging when LogActivity method is available
	// if s.auditLogger != nil {
	//     userID := s.getUserID(ctx)
	//     s.auditLogger.LogActivity(userID, "permission:delete", fmt.Sprintf("Deleted permission: %s", permission.Name))
	// }

	return nil
}

// DeletePermissionWithOrganization deletes a custom permission with proper organization security
// This method ensures only permissions owned by the organization can be deleted
func (s *roleService) DeletePermissionWithOrganization(ctx context.Context, permissionID string, orgID uuid.UUID) error {
	userID := s.getUserID(ctx)

	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return ErrInvalidUUID
	}

	// Get existing permission and validate organization ownership
	permission, err := s.repo.Permission().GetByID(ctx, permissionUUID)
	if err != nil {
		return ErrPermissionNotFound
	}

	// Check if it's a system permission (cannot be deleted)
	if permission.IsSystem {
		return ErrCannotDeleteSystemPerm
	}

	// CRITICAL SECURITY CHECK: Ensure permission belongs to the organization
	if permission.OrganizationID == nil || *permission.OrganizationID != orgID {
		return ErrPermissionNotFound // Don't reveal it exists in another org
	}

	// Check if permission is used by any roles IN THIS ORGANIZATION ONLY
	roles, err := s.repo.Role().GetRolesByPermissionAndOrganization(ctx, permission.Name, orgID.String())
	if err != nil {
		return fmt.Errorf("failed to check permission usage: %w", err)
	}

	if len(roles) > 0 {
		return ErrPermissionInUse
	}

	// Delete the permission
	err = s.repo.Permission().Delete(ctx, permissionUUID)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	if s.auditLogger != nil {
		s.auditLogger.LogOrganizationAction(userID, "delete_permission", orgID.String(), permissionUUID.String(), "", true, nil, fmt.Sprintf("Deleted custom permission: %s", permission.Name))
	}

	return nil
}
