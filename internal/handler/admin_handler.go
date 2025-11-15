package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"auth-service/internal/service"
)

// AdminHandler handles admin-only endpoints
type AdminHandler struct {
	authService service.AuthService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(authService service.AuthService) *AdminHandler {
	return &AdminHandler{
		authService: authService,
	}
}

// ListUsers handles listing users with cursor-based pagination
func (h *AdminHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	cursor := c.DefaultQuery("cursor", "")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	response, err := h.authService.UserService().ListUsers(c.Request.Context(), limit, cursor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ActivateUser handles user activation
func (h *AdminHandler) ActivateUser(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "User ID is required",
		})
		return
	}

	err := h.authService.UserService().ActivateUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User activated successfully",
	})
}

// DeactivateUser handles user deactivation
func (h *AdminHandler) DeactivateUser(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "User ID is required",
		})
		return
	}

	err := h.authService.UserService().DeactivateUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User deactivated successfully",
	})
}

// DeleteUser handles user deletion
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "User ID is required",
		})
		return
	}

	err := h.authService.UserService().DeleteUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User deleted successfully",
	})
}
