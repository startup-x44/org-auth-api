package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"auth-service/internal/models"
	"auth-service/internal/service"
)

// RBACHandler handles RBAC management endpoints
type RBACHandler struct {
	roleService service.RoleService
}

// NewRBACHandler creates a new RBAC handler
func NewRBACHandler(roleService service.RoleService) *RBACHandler {
	return &RBACHandler{
		roleService: roleService,
	}
}

// ListAllPermissions godoc
// @Summary List all system permissions
// @Tags rbac
// @Produce json
// @Success 200 {object} gin.H{data=[]service.PermissionResponse}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /admin/rbac/permissions [get]
func (h *RBACHandler) ListAllPermissions(c *gin.Context) {
	permissions, err := h.roleService.ListAllPermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    permissions,
	})
}

// ListSystemRoles godoc
// @Summary List system roles
// @Tags rbac
// @Produce json
// @Success 200 {object} gin.H{data=[]service.RoleResponse}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /admin/rbac/roles [get]
func (h *RBACHandler) ListSystemRoles(c *gin.Context) {
	// Check if user is superadmin - only superadmin should manage system roles
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not found in context",
		})
		return
	}

	userObj := user.(*models.User)
	if !userObj.IsSuperadmin {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Only superadmin can manage system roles",
		})
		return
	}

	// List only TRUE system roles (organization_id IS NULL and is_system = true)
	roles, err := h.roleService.ListSystemRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    roles,
	})
}

// CreateSystemRole godoc
// @Summary Create a system role
// @Tags rbac
// @Accept json
// @Produce json
// @Param role body CreateRoleRequest true "Role data"
// @Success 201 {object} gin.H{data=service.RoleResponse}
// @Failure 400 {object} gin.H{success=false,message=string}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /admin/rbac/roles [post]
func (h *RBACHandler) CreateSystemRole(c *gin.Context) {
	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
		})
		return
	}

	// Ensure this is a system role (use nil UUID)
	req.OrganizationID = uuid.Nil

	role, err := h.roleService.CreateRole(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    role,
	})
}

// GetSystemRole godoc
// @Summary Get system role by ID
// @Tags rbac
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} gin.H{data=service.RoleResponse}
// @Failure 400 {object} gin.H{success=false,message=string}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Failure 404 {object} gin.H{success=false,message=string}
// @Router /admin/rbac/roles/{id} [get]
func (h *RBACHandler) GetSystemRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid role ID",
		})
		return
	}

	role, err := h.roleService.GetRole(c.Request.Context(), roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Role not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    role,
	})
}

// UpdateSystemRole godoc
// @Summary Update system role
// @Tags rbac
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param role body UpdateRoleRequest true "Role update data"
// @Success 200 {object} gin.H{data=service.RoleResponse}
// @Failure 400 {object} gin.H{success=false,message=string}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Failure 404 {object} gin.H{success=false,message=string}
// @Router /admin/rbac/roles/{id} [put]
func (h *RBACHandler) UpdateSystemRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid role ID",
		})
		return
	}

	var req service.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
		})
		return
	}

	role, err := h.roleService.UpdateRole(c.Request.Context(), roleID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    role,
	})
}

// DeleteSystemRole godoc
// @Summary Delete system role
// @Tags rbac
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} gin.H{success=bool,message=string}
// @Failure 400 {object} gin.H{success=false,message=string}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Failure 404 {object} gin.H{success=false,message=string}
// @Router /admin/rbac/roles/{id} [delete]
func (h *RBACHandler) DeleteSystemRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid role ID",
		})
		return
	}

	err = h.roleService.DeleteRole(c.Request.Context(), roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Role deleted successfully",
	})
}

// AssignPermissionsToRole godoc
// @Summary Assign permissions to a role
// @Tags rbac
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param permissions body AssignPermissionsRequest true "Permission names to assign"
// @Success 200 {object} gin.H{success=bool,message=string}
// @Failure 400 {object} gin.H{success=false,message=string}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /admin/rbac/roles/{id}/permissions [post]
func (h *RBACHandler) AssignPermissionsToRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid role ID",
		})
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
		})
		return
	}

	// For system roles, use nil organization ID
	err = h.roleService.AssignPermissionsToRoleWithOrganization(c.Request.Context(), roleID, uuid.Nil, req.PermissionNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Permissions assigned successfully",
	})
}

// RevokePermissionsFromRole godoc
// @Summary Revoke permissions from a role
// @Tags rbac
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param permissions body AssignPermissionsRequest true "Permission names to revoke"
// @Success 200 {object} gin.H{success=bool,message=string}
// @Failure 400 {object} gin.H{success=false,message=string}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /admin/rbac/roles/{id}/permissions [delete]
func (h *RBACHandler) RevokePermissionsFromRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid role ID",
		})
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
		})
		return
	}

	// For system roles, use nil organization ID
	err = h.roleService.RevokePermissionsFromRoleWithOrganization(c.Request.Context(), roleID, uuid.Nil, req.PermissionNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Permissions revoked successfully",
	})
}

// GetRolePermissions godoc
// @Summary Get permissions assigned to a role
// @Tags rbac
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} gin.H{data=[]string}
// @Failure 400 {object} gin.H{success=false,message=string}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /admin/rbac/roles/{id}/permissions [get]
func (h *RBACHandler) GetRolePermissions(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid role ID",
		})
		return
	}

	permissions, err := h.roleService.GetRolePermissions(c.Request.Context(), roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    permissions,
	})
}

// GetRBACStats godoc
// @Summary Get RBAC system statistics
// @Tags rbac
// @Produce json
// @Success 200 {object} gin.H{data=RBACStats}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /admin/rbac/stats [get]
func (h *RBACHandler) GetRBACStats(c *gin.Context) {
	// Get system permissions
	permissions, err := h.roleService.ListAllPermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get permissions",
		})
		return
	}

	// Get system roles
	roles, err := h.roleService.GetRolesByOrganization(c.Request.Context(), uuid.Nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get system roles",
		})
		return
	}

	// Calculate stats
	stats := RBACStats{
		TotalPermissions: len(permissions),
		SystemRoles:      len(roles),
		SystemPermissions: func() int {
			count := 0
			for _, p := range permissions {
				if p.IsSystem {
					count++
				}
			}
			return count
		}(),
		CustomPermissions: func() int {
			count := 0
			for _, p := range permissions {
				if !p.IsSystem {
					count++
				}
			}
			return count
		}(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// Request/Response types
type AssignPermissionsRequest struct {
	PermissionNames []string `json:"permission_names" binding:"required"`
}

type RBACStats struct {
	TotalPermissions  int `json:"total_permissions"`
	SystemPermissions int `json:"system_permissions"`
	CustomPermissions int `json:"custom_permissions"`
	SystemRoles       int `json:"system_roles"`
}
