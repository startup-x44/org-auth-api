package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"auth-service/internal/service"
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
	authService service.AuthService
}

func NewRoleHandler(authService service.AuthService) *RoleHandler {
	return &RoleHandler{authService: authService}
}

// ------------------------
// Helper Methods
// ------------------------

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
	return err == nil && hasPermission
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

// ------------------------
// Role: CREATE
// ------------------------

func (h *RoleHandler) CreateRole(c *gin.Context) {
	orgID := c.Param("orgId")
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid organization ID format")
		return
	}

	if !h.checkPermission(c, orgID, "role:create") {
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	req.OrganizationID = orgUUID

	result, err := h.authService.RoleService().CreateRole(c.Request.Context(), &req)
	if err != nil {
		h.errorResponse(c, http.StatusUnprocessableEntity, err.Error())
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
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	roleUUID, orgUUID, err := h.parseUUIDs(c)
	if err != nil {
		return
	}

	result, err := h.authService.RoleService().GetRoleWithOrganization(c.Request.Context(), roleUUID, orgUUID)
	if err != nil {
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
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	roleUUID, orgUUID, err := h.parseUUIDs(c)
	if err != nil {
		return
	}

	var req service.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.RoleService().UpdateRoleWithOrganization(c.Request.Context(), roleUUID, orgUUID, &req)
	if err != nil {
		h.errorResponse(c, http.StatusUnprocessableEntity, err.Error())
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
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	roleUUID, orgUUID, err := h.parseUUIDs(c)
	if err != nil {
		return
	}

	err = h.authService.RoleService().DeleteRoleWithOrganization(c.Request.Context(), roleUUID, orgUUID)
	if err != nil {
		h.errorResponse(c, http.StatusUnprocessableEntity, err.Error())
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
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// FIXED: this now returns global system + org's custom only
	result, err := h.authService.RoleService().ListAllPermissionsForOrganization(c.Request.Context(), orgUUID)
	if err != nil {
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
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	var req struct {
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// FIXED: use org-aware method
	err = h.authService.RoleService().AssignPermissionsToRoleWithOrganization(c.Request.Context(), roleUUID, orgUUID, req.Permissions)
	if err != nil {
		h.errorResponse(c, http.StatusUnprocessableEntity, err.Error())
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
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	var req struct {
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.authService.RoleService().RevokePermissionsFromRoleWithOrganization(c.Request.Context(), roleUUID, orgUUID, req.Permissions)
	if err != nil {
		h.errorResponse(c, http.StatusUnprocessableEntity, err.Error())
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
		h.errorResponse(c, http.StatusForbidden,
			"Insufficient permissions")
		return
	}

	var req struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
		Category    string `json:"category"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
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
	if err != nil {
		h.errorResponse(c, http.StatusUnprocessableEntity, err.Error())
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

	if !h.checkPermission(c, orgID, "role:update") {
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
		Category    string `json:"category"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
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
	if err != nil {
		h.errorResponse(c, http.StatusUnprocessableEntity, err.Error())
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

	if !h.checkPermission(c, orgID, "role:delete") {
		h.errorResponse(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	err = h.authService.RoleService().DeletePermissionWithOrganization(
		c.Request.Context(),
		permID,
		orgUUID,
	)
	if err != nil {
		h.errorResponse(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.successResponse(c, http.StatusOK, "Permission deleted", nil)
}
