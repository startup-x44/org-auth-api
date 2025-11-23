package handler

import (
	"fmt"
	"net/http"

	"auth-service/internal/models"
	"auth-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ClientAppHandler handles client application management endpoints
type ClientAppHandler struct {
	clientAppService service.ClientAppService
}

// NewClientAppHandler creates a new client app handler
func NewClientAppHandler(clientAppService service.ClientAppService) *ClientAppHandler {
	return &ClientAppHandler{
		clientAppService: clientAppService,
	}
}

// checkOwnerOrAdmin checks if user is superadmin or organization owner
func checkOwnerOrAdmin(c *gin.Context) bool {
	isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)
	if isSuperadmin {
		return true
	}

	orgRole, _ := c.Request.Context().Value("organization_role").(string)
	return orgRole == "owner"
}

// CreateClientApp godoc
// @Summary Create a new OAuth2 client application (superadmin only)
// @Tags client-apps
// @Accept json
// @Produce json
// @Param request body service.CreateClientAppRequest true "Client app details"
// @Success 201 {object} gin.H{data=service.ClientAppResponse,client_secret=string}
// @Failure 400 {object} gin.H{success=false,message=string}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /client-apps [post]
func (h *ClientAppHandler) CreateClientApp(c *gin.Context) {
	// Get user_id from context (set by AuthRequired middleware)
	// Handle both uuid.UUID and string types
	var userID uuid.UUID
	var exists bool

	userIDVal := c.Request.Context().Value("user_id")
	if userIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized - user_id not found in context",
		})
		return
	}

	switch v := userIDVal.(type) {
	case uuid.UUID:
		userID = v
		exists = true
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "unauthorized - invalid user_id format",
			})
			return
		}
		userID = parsed
		exists = true
	default:
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": fmt.Sprintf("unauthorized - unexpected user_id type: %T", v),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized - user_id not found in context",
		})
		return
	}

	// Check if superadmin OR organization owner
	isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)
	orgRole, _ := c.Request.Context().Value("organization_role").(string)

	if !isSuperadmin && orgRole != "owner" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "owner or superadmin access required",
		})
		return
	}

	// Get organization_id from context
	var organizationID uuid.UUID
	if orgIDVal := c.Request.Context().Value("organization_id"); orgIDVal != nil {
		switch v := orgIDVal.(type) {
		case uuid.UUID:
			organizationID = v
		case string:
			parsed, err := uuid.Parse(v)
			if err == nil {
				organizationID = parsed
			}
		}
	}

	// Create minimal user object for service layer
	user := &models.User{
		ID:           userID,
		IsSuperadmin: isSuperadmin,
	}

	var req service.CreateClientAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid request body",
			"error":   err.Error(),
		})
		return
	}

	clientApp, plainSecret, err := h.clientAppService.CreateClientApp(c.Request.Context(), organizationID, &req, user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"data":          clientApp,
		"client_secret": plainSecret, // Only returned once on creation
		"message":       "Client application created successfully. Store the client_secret securely, it will not be shown again.",
	})
}

// ListClientApps godoc
// @Summary List all OAuth2 client applications (superadmin only)
// @Tags client-apps
// @Produce json
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} gin.H{data=[]service.ClientAppResponse,total=int}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /client-apps [get]
func (h *ClientAppHandler) ListClientApps(c *gin.Context) {
	// Get user_id from context (set by AuthRequired middleware)
	// Handle both uuid.UUID and string types
	var userID uuid.UUID
	var exists bool

	userIDVal := c.Request.Context().Value("user_id")
	if userIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized - user_id not found in context",
		})
		return
	}

	switch v := userIDVal.(type) {
	case uuid.UUID:
		userID = v
		exists = true
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "unauthorized - invalid user_id format",
			})
			return
		}
		userID = parsed
		exists = true
	default:
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": fmt.Sprintf("unauthorized - unexpected user_id type: %T", v),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized - user_id not found in context",
		})
		return
	}

	// Check if superadmin OR organization owner
	isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)
	orgRole, _ := c.Request.Context().Value("organization_role").(string)

	if !isSuperadmin && orgRole != "owner" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "owner or superadmin access required",
		})
		return
	}

	// Create minimal user object for service layer
	user := &models.User{
		ID:           userID,
		IsSuperadmin: isSuperadmin,
	}

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 50
		}
	}
	if o := c.Query("offset"); o != "" {
		if _, err := fmt.Sscanf(o, "%d", &offset); err != nil {
			offset = 0
		}
	}

	clientApps, total, err := h.clientAppService.ListClientApps(c.Request.Context(), limit, offset, user)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    clientApps,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetClientApp godoc
// @Summary Get a specific OAuth2 client application (superadmin only)
// @Tags client-apps
// @Produce json
// @Param id path string true "Client App ID (UUID)"
// @Success 200 {object} gin.H{data=service.ClientAppResponse}
// @Failure 400 {object} gin.H{success=false,message=string}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Failure 404 {object} gin.H{success=false,message=string}
// @Router /client-apps/{id} [get]
func (h *ClientAppHandler) GetClientApp(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	// Check if superadmin OR organization owner
	if !checkOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "owner or superadmin access required",
		})
		return
	}

	user := userVal.(*models.User)

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid client app ID",
		})
		return
	}

	clientApp, err := h.clientAppService.GetClientApp(c.Request.Context(), id, user)
	if err != nil {
		status := http.StatusForbidden
		if err.Error() == "client app not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    clientApp,
	})
}

// UpdateClientApp godoc
// @Summary Update an OAuth2 client application (superadmin only)
// @Tags client-apps
// @Accept json
// @Produce json
// @Param id path string true "Client App ID (UUID)"
// @Param request body service.UpdateClientAppRequest true "Updated client app details"
// @Success 200 {object} gin.H{data=service.ClientAppResponse}
// @Failure 400 {object} gin.H{success=false,message=string}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Failure 404 {object} gin.H{success=false,message=string}
// @Router /client-apps/{id} [put]
func (h *ClientAppHandler) UpdateClientApp(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	// Check if superadmin OR organization owner
	if !checkOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "owner or superadmin access required",
		})
		return
	}

	user := userVal.(*models.User)

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid client app ID",
		})
		return
	}

	var req service.UpdateClientAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid request body",
			"error":   err.Error(),
		})
		return
	}

	clientApp, err := h.clientAppService.UpdateClientApp(c.Request.Context(), id, &req, user)
	if err != nil {
		status := http.StatusForbidden
		if err.Error() == "client app not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    clientApp,
		"message": "Client application updated successfully",
	})
}

// DeleteClientApp godoc
// @Summary Delete an OAuth2 client application (superadmin only)
// @Tags client-apps
// @Produce json
// @Param id path string true "Client App ID (UUID)"
// @Success 200 {object} gin.H{success=true,message=string}
// @Failure 400 {object} gin.H{success=false,message=string}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /client-apps/{id} [delete]
func (h *ClientAppHandler) DeleteClientApp(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	// Check if superadmin OR organization owner
	if !checkOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "owner or superadmin access required",
		})
		return
	}

	user := userVal.(*models.User)

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid client app ID",
		})
		return
	}

	err = h.clientAppService.DeleteClientApp(c.Request.Context(), id, user)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Client application deleted successfully",
	})
}

// RotateClientSecret godoc
// @Summary Rotate client secret for an OAuth2 application (superadmin only)
// @Tags client-apps
// @Produce json
// @Param id path string true "Client App ID (UUID)"
// @Success 200 {object} gin.H{client_secret=string}
// @Failure 400 {object} gin.H{success=false,message=string}
// @Failure 401 {object} gin.H{success=false,message=string}
// @Failure 403 {object} gin.H{success=false,message=string}
// @Router /client-apps/{id}/rotate-secret [post]
func (h *ClientAppHandler) RotateClientSecret(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	// Check if superadmin OR organization owner
	if !checkOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "owner or superadmin access required",
		})
		return
	}

	user := userVal.(*models.User)

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid client app ID",
		})
		return
	}

	newSecret, err := h.clientAppService.RotateClientSecret(c.Request.Context(), id, user)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"client_secret": newSecret,
		"message":       "Client secret rotated successfully. Store the new secret securely, it will not be shown again.",
	})
}

// RegisterClientAppRoutes registers all OAuth2 client app routes
func RegisterClientAppRoutes(r *gin.RouterGroup, handler *ClientAppHandler) {
	clientApps := r.Group("/client-apps")
	{
		clientApps.POST("", handler.CreateClientApp)
		clientApps.GET("", handler.ListClientApps)
		clientApps.GET("/:id", handler.GetClientApp)
		clientApps.PUT("/:id", handler.UpdateClientApp)
		clientApps.DELETE("/:id", handler.DeleteClientApp)
		clientApps.POST("/:id/rotate-secret", handler.RotateClientSecret)
	}
}
