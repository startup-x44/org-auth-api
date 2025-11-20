package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"auth-service/internal/models"
	"auth-service/internal/service"
	"auth-service/pkg/logger"
)

// ------------------------
// Helper Response Methods
// ------------------------

func (h *RoleHandler) errorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"success": false,
		"message": message,
	})
}

func (h *RoleHandler) successResponse(c *gin.Context, code int, message string, data interface{}) {
	resp := gin.H{"success": true}
	if message != "" {
		resp["message"] = message
	}
	if data != nil {
		resp["data"] = data
	}
	c.JSON(code, resp)
}

// ------------------------
// Handler Struct
// ------------------------

type RoleHandler struct {
	authService  service.AuthService
	auditService service.AuditService
}

func NewRoleHandler(authService service.AuthService, auditService service.AuditService) *RoleHandler {
	return &RoleHandler{
		authService:  authService,
		auditService: auditService,
	}
}

// ------------------------
// Helper Methods
// ------------------------

func (h *RoleHandler) getUserID(c *gin.Context) (*uuid.UUID, error) {
	userIDStr, ok := c.Request.Context().Value("user_id").(string)
	if !ok || userIDStr == "" {
		return nil, nil
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}
	return &userID, nil
}

func (h *RoleHandler) checkPermission(c *gin.Context, orgID, permission string) bool {
	userID, ok := c.Request.Context().Value("user_id").(string)
	if !ok {
		return false
	}

	isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)
	if isSuperadmin {
		return true
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false
	}
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return false
	}

	hasPermission, err := h.authService.RoleService().HasPermission(c.Request.Context(), userUUID, orgUUID, permission)
	if err != nil {
		logger.Error(c.Request.Context()).
			Err(err).
			Str("user_id", userID).
			Str("org_id", orgID).
			Str("permission", permission).
			Msg("Failed to check permission")
		return false
	}
	return hasPermission
}

func (h *RoleHandler) parseUUIDs(c *gin.Context) (uuid.UUID, uuid.UUID, error) {
	roleID := c.Param("roleId")
	orgID := c.Param("orgId")

	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid role ID format")
		return uuid.Nil, uuid.Nil, err
	}

	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid organization ID format")
		return uuid.Nil, uuid.Nil, err
	}

	return roleUUID, orgUUID, nil
}

func (h *RoleHandler) logAuthorizationFailure(c *gin.Context, resource string, resourceID *uuid.UUID, orgID *uuid.UUID, requiredPermission string) {
	userID, _ := h.getUserID(c)
	details := map[string]interface{}{
		"required_permission": requiredPermission,
	}
	if orgID != nil {
		details["org_id"] = orgID.String()
	}
	if resourceID != nil {
		details["resource_id"] = resourceID.String()
	}

	h.auditService.LogPermission(c.Request.Context(), models.ActionAuthorizationFailed, *userID, resourceID, orgID, false, details, nil)
}

// ------------------------
// Role: CREATE
// ------------------------

func (h *RoleHandler) CreateRole(c *gin.Context) {
	orgID := c.Param("orgId")
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	if !h.checkPermission(c, orgID, "role:create") {
		h.logAuthorizationFailure(c, models.ResourceRole, nil, &orgUUID, "role:create")
		h.errorResponse(c, http.StatusForbidden, "You don't have permission to create roles")
		return
	}

	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Please provide valid role information")
		return
	}

	req.OrganizationID = orgUUID

	result, err := h.authService.RoleService().CreateRole(c.Request.Context(), &req)

	userID, _ := h.getUserID(c)
	var roleID *uuid.UUID
	if result != nil {
		roleID = &result.ID
	}

	h.auditService.LogRole(c.Request.Context(), models.ActionRoleCreate, *userID, roleID, &orgUUID, err == nil, map[string]interface{}{
		"role_name":        req.Name,
		"role_description": req.Description,
	}, err)

	if err != nil {
		logger.Error(c.Request.Context()).Err(err).Msg("Failed to create role")
		h.errorResponse(c, http.StatusBadRequest, "Failed to create role")
		return
	}

	h.successResponse(c, http.StatusCreated, "Role created successfully", result)
}

// ------------------------
// Role: GET
// ------------------------

func (h *RoleHandler) GetRole(c *gin.Context) {
	orgID := c.Param("orgId")

	if !h.checkPermission(c, orgID, "role:view") {
		orgUUID, _ := uuid.Parse(orgID)
		h.logAuthorizationFailure(c, models.ResourceRole, nil, &orgUUID, "role:view")
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	roleUUID, orgUUID, err := h.parseUUIDs(c)
	if err != nil {
		return
	}

	result, err := h.authService.RoleService().GetRoleWithOrganization(c.Request.Context(), roleUUID, orgUUID)

	userID, _ := h.getUserID(c)
	h.auditService.LogRole(c.Request.Context(), models.ActionRoleView, *userID, &roleUUID, &orgUUID, err == nil, map[string]interface{}{
		"role_id": roleUUID.String(),
	}, err)

	if err != nil {
		logger.Error(c.Request.Context()).Err(err).Str("role_id", roleUUID.String()).Msg("Failed to get role")
		h.errorResponse(c, http.StatusNotFound, "Role not found")
		return
	}

	h.successResponse(c, http.StatusOK, "", result)
}

// ------------------------
// Role: UPDATE
// ------------------------

func (h *RoleHandler) UpdateRole(c *gin.Context) {
	orgID := c.Param("orgId")

	if !h.checkPermission(c, orgID, "role:update") {
		orgUUID, _ := uuid.Parse(orgID)
		roleUUID, _ := uuid.Parse(c.Param("roleId"))
		h.logAuthorizationFailure(c, models.ResourceRole, &roleUUID, &orgUUID, "role:update")
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	roleUUID, orgUUID, err := h.parseUUIDs(c)
	if err != nil {
		return
	}

	var req service.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Please provide valid role information")
		return
	}

	result, err := h.authService.RoleService().UpdateRoleWithOrganization(c.Request.Context(), roleUUID, orgUUID, &req)

	userID, _ := h.getUserID(c)
	h.auditService.LogRole(c.Request.Context(), models.ActionRoleUpdate, *userID, &roleUUID, &orgUUID, err == nil, map[string]interface{}{
		"role_id":      roleUUID.String(),
		"display_name": req.DisplayName,
		"description":  req.Description,
	}, err)

	if err != nil {
		logger.Error(c.Request.Context()).Err(err).Str("role_id", roleUUID.String()).Msg("Failed to update role")
		h.errorResponse(c, http.StatusBadRequest, "Failed to update role")
		return
	}

	h.successResponse(c, http.StatusOK, "Role updated successfully", result)
}

// ------------------------
// Role: DELETE
// ------------------------

func (h *RoleHandler) DeleteRole(c *gin.Context) {
	orgID := c.Param("orgId")

	if !h.checkPermission(c, orgID, "role:delete") {
		orgUUID, _ := uuid.Parse(orgID)
		roleUUID, _ := uuid.Parse(c.Param("roleId"))
		h.logAuthorizationFailure(c, models.ResourceRole, &roleUUID, &orgUUID, "role:delete")
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	roleUUID, orgUUID, err := h.parseUUIDs(c)
	if err != nil {
		return
	}

	err = h.authService.RoleService().DeleteRoleWithOrganization(c.Request.Context(), roleUUID, orgUUID)

	userID, _ := h.getUserID(c)
	h.auditService.LogRole(c.Request.Context(), models.ActionRoleDelete, *userID, &roleUUID, &orgUUID, err == nil, map[string]interface{}{
		"role_id": roleUUID.String(),
	}, err)

	if err != nil {
		logger.Error(c.Request.Context()).Err(err).Str("role_id", roleUUID.String()).Msg("Failed to delete role")
		h.errorResponse(c, http.StatusUnprocessableEntity, "Failed to delete role")
		return
	}

	h.successResponse(c, http.StatusOK, "Role deleted", nil)
}

// ------------------------
// Permissions: LIST (global system + org custom)
// ------------------------

func (h *RoleHandler) ListPermissions(c *gin.Context) {
	orgID := c.Param("orgId")
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid organization ID format")
		return
	}

	if !h.checkPermission(c, orgID, "role:view") {
		h.logAuthorizationFailure(c, models.ResourcePermission, nil, &orgUUID, "role:view")
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)

	var result []*service.PermissionResponse

	if isSuperadmin {
		result, err = h.authService.RoleService().ListAllPermissionsForOrganization(c.Request.Context(), orgUUID)
	} else {
		result, err = h.authService.RoleService().ListCustomPermissionsForOrganization(c.Request.Context(), orgUUID)
	}

	userID, _ := h.getUserID(c)
	h.auditService.LogPermission(c.Request.Context(), models.ActionPermissionView, *userID, nil, &orgUUID, err == nil, map[string]interface{}{
		"is_superadmin": isSuperadmin,
		"count":         len(result),
	}, err)

	if err != nil {
		logger.Error(c.Request.Context()).Err(err).Msg("Failed to load permissions")
		h.errorResponse(c, http.StatusInternalServerError, "Failed to load permissions")
		return
	}

	h.successResponse(c, http.StatusOK, "", result)
}

// ------------------------
// Permissions: ASSIGN
// ------------------------

func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	orgID := c.Param("orgId")
	roleUUID, orgUUID, err := h.parseUUIDs(c)
	if err != nil {
		return
	}

	if !h.checkPermission(c, orgID, "role:update") {
		h.logAuthorizationFailure(c, models.ResourcePermission, &roleUUID, &orgUUID, "role:update")
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	var req struct {
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	err = h.authService.RoleService().AssignPermissionsToRoleWithOrganization(c.Request.Context(), roleUUID, orgUUID, req.Permissions)

	userID, _ := h.getUserID(c)
	h.auditService.LogPermission(c.Request.Context(), models.ActionPermissionGrant, *userID, &roleUUID, &orgUUID, err == nil, map[string]interface{}{
		"role_id":     roleUUID.String(),
		"permissions": req.Permissions,
	}, err)

	if err != nil {
		logger.Error(c.Request.Context()).Err(err).Str("role_id", roleUUID.String()).Msg("Failed to assign permissions")
		h.errorResponse(c, http.StatusUnprocessableEntity, "Failed to assign permissions")
		return
	}

	h.successResponse(c, http.StatusOK, "Permissions assigned", nil)
}

// ------------------------
// Permissions: REVOKE
// ------------------------

func (h *RoleHandler) RevokePermissions(c *gin.Context) {
	orgID := c.Param("orgId")
	roleUUID, orgUUID, err := h.parseUUIDs(c)
	if err != nil {
		return
	}

	if !h.checkPermission(c, orgID, "role:update") {
		h.logAuthorizationFailure(c, models.ResourcePermission, &roleUUID, &orgUUID, "role:update")
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	var req struct {
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	err = h.authService.RoleService().RevokePermissionsFromRoleWithOrganization(c.Request.Context(), roleUUID, orgUUID, req.Permissions)

	userID, _ := h.getUserID(c)
	h.auditService.LogPermission(c.Request.Context(), models.ActionPermissionRevoke, *userID, &roleUUID, &orgUUID, err == nil, map[string]interface{}{
		"role_id":     roleUUID.String(),
		"permissions": req.Permissions,
	}, err)

	if err != nil {
		logger.Error(c.Request.Context()).Err(err).Str("role_id", roleUUID.String()).Msg("Failed to revoke permissions")
		h.errorResponse(c, http.StatusUnprocessableEntity, "Failed to revoke permissions")
		return
	}

	h.successResponse(c, http.StatusOK, "Permissions revoked", nil)
}

// ------------------------
// Custom Permission: CREATE
// ------------------------

func (h *RoleHandler) CreatePermission(c *gin.Context) {
	orgID := c.Param("orgId")
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	if !h.checkPermission(c, orgID, "role:create") {
		h.logAuthorizationFailure(c, models.ResourcePermission, nil, &orgUUID, "role:create")
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	var req struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
		Category    string `json:"category"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	result, err := h.authService.RoleService().CreatePermissionForOrganization(
		c.Request.Context(),
		orgUUID,
		req.Name,
		req.DisplayName,
		req.Description,
		req.Category,
	)

	userID, _ := h.getUserID(c)
	var permID *uuid.UUID
	if result != nil && result.ID != uuid.Nil {
		permID = &result.ID
	}

	h.auditService.LogPermission(c.Request.Context(), models.ActionPermissionCreate, *userID, permID, &orgUUID, err == nil, map[string]interface{}{
		"name":         req.Name,
		"display_name": req.DisplayName,
		"description":  req.Description,
		"category":     req.Category,
	}, err)

	if err != nil {
		logger.Error(c.Request.Context()).Err(err).Msg("Failed to create permission")
		h.errorResponse(c, http.StatusUnprocessableEntity, "Failed to create permission")
		return
	}

	h.successResponse(c, http.StatusCreated, "Permission created", result)
}

// ------------------------
// Custom Permission: UPDATE
// ------------------------

func (h *RoleHandler) UpdatePermission(c *gin.Context) {
	orgID := c.Param("orgId")
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	permissionID := c.Param("permission_id")
	permUUID, err := uuid.Parse(permissionID)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	if !h.checkPermission(c, orgID, "role:update") {
		h.logAuthorizationFailure(c, models.ResourcePermission, &permUUID, &orgUUID, "role:update")
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
		Category    string `json:"category"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	result, err := h.authService.RoleService().UpdatePermissionWithOrganization(
		c.Request.Context(),
		permissionID,
		orgUUID,
		req.DisplayName,
		req.Description,
		req.Category,
	)

	userID, _ := h.getUserID(c)
	h.auditService.LogPermission(c.Request.Context(), models.ActionPermissionUpdate, *userID, &permUUID, &orgUUID, err == nil, map[string]interface{}{
		"permission_id": permissionID,
		"display_name":  req.DisplayName,
		"description":   req.Description,
		"category":      req.Category,
	}, err)

	if err != nil {
		logger.Error(c.Request.Context()).Err(err).Str("permission_id", permissionID).Msg("Failed to update permission")
		h.errorResponse(c, http.StatusUnprocessableEntity, "Failed to update permission")
		return
	}

	h.successResponse(c, http.StatusOK, "Permission updated successfully", result)
}

// ------------------------
// Custom Permission: DELETE
// ------------------------

func (h *RoleHandler) DeletePermission(c *gin.Context) {
	orgID := c.Param("orgId")
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	permID := c.Param("permission_id")
	permUUID, err := uuid.Parse(permID)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	if !h.checkPermission(c, orgID, "role:delete") {
		h.logAuthorizationFailure(c, models.ResourcePermission, &permUUID, &orgUUID, "role:delete")
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	err = h.authService.RoleService().DeletePermissionWithOrganization(
		c.Request.Context(),
		permID,
		orgUUID,
	)

	userID, _ := h.getUserID(c)
	h.auditService.LogPermission(c.Request.Context(), models.ActionPermissionDelete, *userID, &permUUID, &orgUUID, err == nil, map[string]interface{}{
		"permission_id": permID,
	}, err)

	if err != nil {
		logger.Error(c.Request.Context()).Err(err).Str("permission_id", permID).Msg("Failed to delete permission")
		h.errorResponse(c, http.StatusUnprocessableEntity, "Failed to delete permission")
		return
	}

	h.successResponse(c, http.StatusOK, "Permission deleted", nil)
}
