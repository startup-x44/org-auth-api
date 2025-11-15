package service

import (
	"context"
	"errors"
	"fmt"

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
	GetRolesByOrganization(ctx context.Context, orgID uuid.UUID) ([]*RoleResponse, error)
	UpdateRole(ctx context.Context, roleID uuid.UUID, req *UpdateRoleRequest) (*RoleResponse, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) error

	// Permission assignment
	AssignPermissionsToRole(ctx context.Context, roleID uuid.UUID, permissionNames []string) error
	RevokePermissionsFromRole(ctx context.Context, roleID uuid.UUID, permissionNames []string) error
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]string, error)

	// System permissions
	ListAllPermissions(ctx context.Context) ([]*PermissionResponse, error)
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
	OrganizationID uuid.UUID `json:"organization_id" binding:"required"`
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
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	DisplayName    string    `json:"display_name"`
	Description    string    `json:"description"`
	IsSystem       bool      `json:"is_system"`
	Permissions    []string  `json:"permissions,omitempty"`
	MemberCount    int       `json:"member_count,omitempty"`
}

type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
}

// HasPermission checks if a user has a specific permission in an organization
func (s *roleService) HasPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) (bool, error) {
	// Get user's membership
	membership, err := s.repo.OrganizationMembership().GetByOrganizationAndUser(ctx, orgID.String(), userID.String())
	if err != nil {
		return false, fmt.Errorf("membership not found: %w", err)
	}

	if membership.Status != models.MembershipStatusActive {
		return false, errors.New("membership is not active")
	}

	// Load role
	role, err := s.repo.Role().GetByID(ctx, membership.RoleID.String())
	if err != nil {
		return false, fmt.Errorf("failed to load role: %w", err)
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

	// Load role
	role, err := s.repo.Role().GetByID(ctx, membership.RoleID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to load role: %w", err)
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

	// Prevent creating system role names
	if req.Name == models.RoleNameAdmin {
		return nil, errors.New("cannot create system role 'admin'")
	}

	// Check if role name already exists in organization
	existingRole, err := s.repo.Role().GetByOrganizationAndName(ctx, req.OrganizationID.String(), req.Name)
	if err == nil && existingRole != nil {
		return nil, errors.New("role name already exists in organization")
	}

	// Create role
	role := &models.Role{
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		IsSystem:       false,
		CreatedBy:      uuid.MustParse(userID),
	}

	if err := s.repo.Role().Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Assign permissions if provided
	if len(req.Permissions) > 0 {
		if err := s.AssignPermissionsToRole(ctx, role.ID, req.Permissions); err != nil {
			return nil, fmt.Errorf("failed to assign permissions: %w", err)
		}
	}

	s.auditLogger.LogOrganizationAction(userID, "create_role", req.OrganizationID.String(), role.ID.String(), "", true, nil, fmt.Sprintf("Created role: %s", role.Name))

	return s.GetRole(ctx, role.ID)
}

// GetRole retrieves a role by ID
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

// GetRolesByOrganization retrieves all roles for an organization
func (s *roleService) GetRolesByOrganization(ctx context.Context, orgID uuid.UUID) ([]*RoleResponse, error) {
	roles, err := s.repo.Role().GetByOrganization(ctx, orgID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
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

// UpdateRole updates a role's details
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

	// Update fields
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
		existingPerms, _ := s.GetRolePermissions(ctx, roleID)
		if len(existingPerms) > 0 {
			if err := s.RevokePermissionsFromRole(ctx, roleID, existingPerms); err != nil {
				return nil, fmt.Errorf("failed to revoke existing permissions: %w", err)
			}
		}

		// Assign new permissions
		if len(req.Permissions) > 0 {
			if err := s.AssignPermissionsToRole(ctx, roleID, req.Permissions); err != nil {
				return nil, fmt.Errorf("failed to assign new permissions: %w", err)
			}
		}
	}

	s.auditLogger.LogOrganizationAction(userID, "update_role", role.OrganizationID.String(), roleID.String(), "", true, nil, fmt.Sprintf("Updated role: %s", role.Name))

	return s.GetRole(ctx, roleID)
}

// DeleteRole deletes a custom role
func (s *roleService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	userID := s.getUserID(ctx)

	role, err := s.repo.Role().GetByID(ctx, roleID.String())
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
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
	if err := s.repo.Role().Delete(ctx, roleID.String()); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	s.auditLogger.LogOrganizationAction(userID, "delete_role", role.OrganizationID.String(), roleID.String(), "", true, nil, fmt.Sprintf("Deleted role: %s", role.Name))

	return nil
}

// AssignPermissionsToRole assigns permissions to a role
func (s *roleService) AssignPermissionsToRole(ctx context.Context, roleID uuid.UUID, permissionNames []string) error {
	userID := s.getUserID(ctx)

	role, err := s.repo.Role().GetByID(ctx, roleID.String())
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Prevent modifying system role permissions via API (should only be done in code)
	if role.IsSystem {
		return errors.New("cannot modify system role permissions")
	}

	// Get permissions by names
	perms, err := s.repo.Permission().GetByNames(ctx, permissionNames)
	if err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	if len(perms) != len(permissionNames) {
		return errors.New("some permissions not found")
	}

	// Assign each permission
	for _, perm := range perms {
		if err := s.repo.Permission().AssignToRole(ctx, roleID, perm.ID); err != nil {
			return fmt.Errorf("failed to assign permission %s: %w", perm.Name, err)
		}
	}

	s.auditLogger.LogOrganizationAction(userID, "assign_permissions", role.OrganizationID.String(), roleID.String(), "", true, nil, fmt.Sprintf("Assigned %d permissions to role %s", len(permissionNames), role.Name))

	return nil
}

// RevokePermissionsFromRole revokes permissions from a role
func (s *roleService) RevokePermissionsFromRole(ctx context.Context, roleID uuid.UUID, permissionNames []string) error {
	userID := s.getUserID(ctx)

	role, err := s.repo.Role().GetByID(ctx, roleID.String())
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Prevent modifying system role permissions
	if role.IsSystem {
		return errors.New("cannot modify system role permissions")
	}

	// Get permissions by names
	perms, err := s.repo.Permission().GetByNames(ctx, permissionNames)
	if err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	// Revoke each permission
	for _, perm := range perms {
		if err := s.repo.Permission().RevokeFromRole(ctx, roleID, perm.ID); err != nil {
			return fmt.Errorf("failed to revoke permission %s: %w", perm.Name, err)
		}
	}

	s.auditLogger.LogOrganizationAction(userID, "revoke_permissions", role.OrganizationID.String(), roleID.String(), "", true, nil, fmt.Sprintf("Revoked %d permissions from role %s", len(permissionNames), role.Name))

	return nil
}

// GetRolePermissions returns all permissions for a role
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
			ID:          perm.ID,
			Name:        perm.Name,
			DisplayName: perm.DisplayName,
			Description: perm.Description,
			Category:    perm.Category,
		}
	}

	return responses, nil
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
	if role.Permissions != nil && len(role.Permissions) > 0 {
		permissions := make([]string, len(role.Permissions))
		for i, perm := range role.Permissions {
			permissions[i] = perm.Name
		}
		response.Permissions = permissions
	}

	return response
}
