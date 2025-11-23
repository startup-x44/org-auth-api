package handler

import (
	"net/http"

	"auth-service/internal/models"
	"auth-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIKeyHandler handles HTTP requests for API key management
type APIKeyHandler struct {
	apiKeyService service.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(apiKeyService service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
	}
}

// checkOwnerOrAdminAPIKey checks if user is superadmin or organization owner
func checkOwnerOrAdminAPIKey(c *gin.Context) bool {
	isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)
	if isSuperadmin {
		return true
	}

	orgRole, _ := c.Request.Context().Value("organization_role").(string)
	return orgRole == "owner"
}

// CreateAPIKey godoc
// @Summary Create a new API key
// @Description Create a new API key for the authenticated user
// @Tags developer
// @Accept json
// @Produce json
// @Param request body models.APIKeyCreateRequest true "API key creation request"
// @Success 201 {object} models.APIKeyCreateResponse
// @Failure 400 {object} gin.H{error=string}
// @Failure 401 {object} gin.H{error=string}
// @Failure 500 {object} gin.H{error=string}
// @Router /dev/api-keys [post]
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	// Check if superadmin OR organization owner
	if !checkOwnerOrAdminAPIKey(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "owner or superadmin access required",
		})
		return
	}

	// Get user ID and organization ID from context (set by auth middleware)
	userIDValue := c.Request.Context().Value("user_id")
	organizationIDValue := c.Request.Context().Value("organization_id")
	isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)

	var userID, tenantID uuid.UUID
	var err error

	// Handle both string and UUID types from context for user ID
	switch v := userIDValue.(type) {
	case string:
		userID, err = uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID in context"})
			return
		}
	case uuid.UUID:
		userID = v
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found in context"})
		return
	}

	// Handle organization ID - superadmin users can operate without specific organization context
	switch v := organizationIDValue.(type) {
	case string:
		tenantID, err = uuid.Parse(v)
		if err != nil {
			if isSuperadmin {
				// Use nil UUID for superadmin (global scope)
				tenantID = uuid.Nil
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid organization ID in context"})
				return
			}
		}
	case uuid.UUID:
		tenantID = v
	default:
		if isSuperadmin {
			// Use nil UUID for superadmin (global scope)
			tenantID = uuid.Nil
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "organization ID not found in context"})
			return
		}
	}

	// Parse request body
	var req models.APIKeyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create API key
	apiKey, err := h.apiKeyService.CreateAPIKey(c.Request.Context(), userID, tenantID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    apiKey,
	})
}

// ListAPIKeys godoc
// @Summary List all API keys for the user
// @Description Get all API keys belonging to the authenticated user
// @Tags developer
// @Produce json
// @Success 200 {array} models.APIKeyResponse
// @Failure 401 {object} gin.H{error=string}
// @Failure 500 {object} gin.H{error=string}
// @Router /dev/api-keys [get]
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	// Check if superadmin OR organization owner
	if !checkOwnerOrAdminAPIKey(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "owner or superadmin access required",
		})
		return
	}

	// Get user ID and organization ID from context
	userIDValue := c.Request.Context().Value("user_id")
	organizationIDValue := c.Request.Context().Value("organization_id")
	isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)

	var userID, tenantID uuid.UUID
	var err error

	// Handle both string and UUID types from context for user ID
	switch v := userIDValue.(type) {
	case string:
		userID, err = uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID in context"})
			return
		}
	case uuid.UUID:
		userID = v
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found in context"})
		return
	}

	// Handle organization ID - superadmin users can operate without specific organization context
	switch v := organizationIDValue.(type) {
	case string:
		tenantID, err = uuid.Parse(v)
		if err != nil {
			if isSuperadmin {
				// Use nil UUID for superadmin (global scope)
				tenantID = uuid.Nil
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid organization ID in context"})
				return
			}
		}
	case uuid.UUID:
		tenantID = v
	default:
		if isSuperadmin {
			// Use nil UUID for superadmin (global scope)
			tenantID = uuid.Nil
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "organization ID not found in context"})
			return
		}
	}

	// Get API keys
	apiKeys, err := h.apiKeyService.ListAPIKeys(c.Request.Context(), userID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    apiKeys,
	})
}

// GetAPIKey godoc
// @Summary Get an API key by ID
// @Description Get details of a specific API key
// @Tags developer
// @Produce json
// @Param id path string true "API Key ID"
// @Success 200 {object} models.APIKeyResponse
// @Failure 400 {object} gin.H{error=string}
// @Failure 401 {object} gin.H{error=string}
// @Failure 404 {object} gin.H{error=string}
// @Router /dev/api-keys/{id} [get]
func (h *APIKeyHandler) GetAPIKey(c *gin.Context) {
	// Check if superadmin OR organization owner
	if !checkOwnerOrAdminAPIKey(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "owner or superadmin access required",
		})
		return
	}

	keyID := c.Param("id")
	if keyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key ID is required"})
		return
	}

	// Get user ID and organization ID from context
	userIDValue := c.Request.Context().Value("user_id")
	organizationIDValue := c.Request.Context().Value("organization_id")
	isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)

	var userID, tenantID uuid.UUID
	var err error

	// Handle both string and UUID types from context for user ID
	switch v := userIDValue.(type) {
	case string:
		userID, err = uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID in context"})
			return
		}
	case uuid.UUID:
		userID = v
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found in context"})
		return
	}

	// Handle organization ID - superadmin users can operate without specific organization context
	switch v := organizationIDValue.(type) {
	case string:
		tenantID, err = uuid.Parse(v)
		if err != nil {
			if isSuperadmin {
				// Use nil UUID for superadmin (global scope)
				tenantID = uuid.Nil
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid organization ID in context"})
				return
			}
		}
	case uuid.UUID:
		tenantID = v
	default:
		if isSuperadmin {
			// Use nil UUID for superadmin (global scope)
			tenantID = uuid.Nil
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "organization ID not found in context"})
			return
		}
	}

	// Get API key
	apiKey, err := h.apiKeyService.GetAPIKey(c.Request.Context(), keyID, userID, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    apiKey,
	})
}

// RevokeAPIKey godoc
// @Summary Revoke an API key
// @Description Revoke (disable) an API key
// @Tags developer
// @Param id path string true "API Key ID"
// @Success 200 {object} gin.H{success=bool}
// @Failure 400 {object} gin.H{error=string}
// @Failure 401 {object} gin.H{error=string}
// @Failure 404 {object} gin.H{error=string}
// @Router /dev/api-keys/{id} [delete]
func (h *APIKeyHandler) RevokeAPIKey(c *gin.Context) {
	// Check if superadmin OR organization owner
	if !checkOwnerOrAdminAPIKey(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "owner or superadmin access required",
		})
		return
	}

	keyID := c.Param("id")
	if keyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key ID is required"})
		return
	}

	// Get user ID and organization ID from context
	userIDValue := c.Request.Context().Value("user_id")
	organizationIDValue := c.Request.Context().Value("organization_id")
	isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)

	var userID, tenantID uuid.UUID
	var err error

	// Handle both string and UUID types from context for user ID
	switch v := userIDValue.(type) {
	case string:
		userID, err = uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID in context"})
			return
		}
	case uuid.UUID:
		userID = v
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found in context"})
		return
	}

	// Handle organization ID - superadmin users can operate without specific organization context
	switch v := organizationIDValue.(type) {
	case string:
		tenantID, err = uuid.Parse(v)
		if err != nil {
			if isSuperadmin {
				// Use nil UUID for superadmin (global scope)
				tenantID = uuid.Nil
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid organization ID in context"})
				return
			}
		}
	case uuid.UUID:
		tenantID = v
	default:
		if isSuperadmin {
			// Use nil UUID for superadmin (global scope)
			tenantID = uuid.Nil
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "organization ID not found in context"})
			return
		}
	}

	// Revoke API key
	err = h.apiKeyService.RevokeAPIKey(c.Request.Context(), keyID, userID, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "API key revoked successfully",
	})
}

// RegisterAPIKeyRoutes registers all API key routes
func RegisterAPIKeyRoutes(r *gin.RouterGroup, handler *APIKeyHandler) {
	apiKeys := r.Group("/api-keys")
	{
		apiKeys.POST("", handler.CreateAPIKey)
		apiKeys.GET("", handler.ListAPIKeys)
		apiKeys.GET("/:id", handler.GetAPIKey)
		apiKeys.DELETE("/:id", handler.RevokeAPIKey)
	}
}
