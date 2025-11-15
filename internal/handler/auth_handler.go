package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"auth-service/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterGlobal handles global user registration (no organization yet)
func (h *AuthHandler) RegisterGlobal(c *gin.Context) {
	var req service.RegisterGlobalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	response, err := h.authService.UserService().RegisterGlobal(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "User registered successfully. Please create or join an organization.",
	})
}

// LoginGlobal handles global user login (returns list of organizations)
func (h *AuthHandler) LoginGlobal(c *gin.Context) {
	var req service.LoginGlobalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	response, err := h.authService.UserService().LoginGlobal(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Login successful. Please select an organization.",
	})
}

// SelectOrganization issues org-scoped JWT after user selects an organization
func (h *AuthHandler) SelectOrganization(c *gin.Context) {
	var req service.SelectOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	response, err := h.authService.UserService().SelectOrganization(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Organization selected successfully",
	})
}

// CreateOrganization creates a new Slack-style workspace
func (h *AuthHandler) CreateOrganization(c *gin.Context) {
	var req service.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateOrganization: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	log.Printf("CreateOrganization: Received request: %+v", req)

	// Get user_id from request body (user is not authenticated with org yet)
	userID := req.UserID
	if userID == "" {
		log.Printf("CreateOrganization: UserID is empty in request")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "user ID is required",
		})
		return
	}

	log.Printf("CreateOrganization: Calling service with userID: %s", userID)

	response, err := h.authService.UserService().CreateOrganization(c.Request.Context(), userID, &req)
	if err != nil {
		log.Printf("CreateOrganization: Service error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	log.Printf("CreateOrganization: Success, organization ID: %s", response.Organization.OrganizationID)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "Organization created successfully",
	})
}

// GetMyOrganizations returns list of user's organization memberships
func (h *AuthHandler) GetMyOrganizations(c *gin.Context) {
	userID, _ := c.Request.Context().Value("user_id").(string)

	organizations, err := h.authService.UserService().GetMyOrganizations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    organizations,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req service.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	response, err := h.authService.UserService().RefreshToken(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Token refreshed successfully",
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Request.Context().Value("user_id").(string)

	var req service.LogoutRequest
	req.UserID = userID

	// Optional: get refresh token from request
	if err := c.ShouldBindJSON(&req); err != nil {
		// Ignore binding errors for logout
	}

	err := h.authService.UserService().Logout(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout successful",
	})
}

// GetProfile handles getting user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Request.Context().Value("user_id").(string)

	profile, err := h.authService.UserService().GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}

// UpdateProfile handles updating user profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Request.Context().Value("user_id").(string)

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	profile, err := h.authService.UserService().UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
		"message": "Profile updated successfully",
	})
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, _ := c.Request.Context().Value("user_id").(string)

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	err := h.authService.UserService().ChangePassword(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password changed successfully",
	})
}

// ForgotPassword handles forgot password request
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req service.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	err := h.authService.UserService().ForgotPassword(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password reset email sent",
	})
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	err := h.authService.UserService().ResetPassword(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password reset successfully",
	})
}

// HealthCheck handles health check endpoint
func (h *AuthHandler) HealthCheck(c *gin.Context) {
	response, err := h.authService.HealthCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	statusCode := http.StatusOK
	if response.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"success": response.Status == "healthy",
		"data":    response,
	})
}
