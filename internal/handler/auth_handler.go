package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"auth-service/internal/errors"
	"auth-service/internal/models"
	"auth-service/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService  service.AuthService
	auditService service.AuditService
	errorMapper  *errors.ErrorMapper
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService, auditService service.AuditService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		auditService: auditService,
		errorMapper:  errors.NewErrorMapper(),
	}
}

// RegisterGlobal handles global user registration (no organization yet)
func (h *AuthHandler) RegisterGlobal(c *gin.Context) {
	var req service.RegisterGlobalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.SendErrorResponse(c, errors.ErrCodeValidationFailed, "Invalid request data", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	response, err := h.authService.UserService().RegisterGlobal(c.Request.Context(), &req)

	// Audit log: register attempt
	var userID *uuid.UUID
	if response != nil && response.User != nil {
		parsedID, parseErr := uuid.Parse(response.User.ID)
		if parseErr == nil {
			userID = &parsedID
		}
	}

	h.auditService.LogAuth(c.Request.Context(), models.ActionRegister, userID, err == nil, map[string]interface{}{
		"email":      req.Email,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
	}, err)

	if err != nil {
		errorCode, message := h.errorMapper.MapServiceError(err)
		errors.SendErrorResponse(c, errorCode, message, nil)
		return
	}

	errors.SendSuccessResponse(c, "User registered successfully. Please create or join an organization.", response)
}

// LoginGlobal handles global user login (returns list of organizations)
func (h *AuthHandler) LoginGlobal(c *gin.Context) {
	var req service.LoginGlobalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.SendErrorResponse(c, errors.ErrCodeValidationFailed, "Invalid request data", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	response, err := h.authService.UserService().LoginGlobal(c.Request.Context(), &req)

	// Audit log: login attempt
	var userID *uuid.UUID
	if response != nil && response.User != nil {
		parsedID, parseErr := uuid.Parse(response.User.ID)
		if parseErr == nil {
			userID = &parsedID
		}
	}

	action := models.ActionLogin
	if err != nil {
		action = models.ActionLoginFailed
	}

	h.auditService.LogAuth(c.Request.Context(), action, userID, err == nil, map[string]interface{}{
		"email":               req.Email,
		"organizations_count": len(response.Organizations),
	}, err)

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

	// Audit log: logout
	var parsedUserID *uuid.UUID
	if userID != "" {
		id, parseErr := uuid.Parse(userID)
		if parseErr == nil {
			parsedUserID = &id
		}
	}

	h.auditService.LogAuth(c.Request.Context(), models.ActionLogout, parsedUserID, err == nil, map[string]interface{}{
		"session_id": req.SessionID,
	}, err)

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

	// Audit log: password change
	var parsedUserID *uuid.UUID
	if userID != "" {
		id, parseErr := uuid.Parse(userID)
		if parseErr == nil {
			parsedUserID = &id
		}
	}

	h.auditService.LogAuth(c.Request.Context(), models.ActionPasswordChange, parsedUserID, err == nil, nil, err)

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

// VerifyEmail handles email verification with 6-digit code
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required,len=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	if err := h.authService.UserService().VerifyEmail(c.Request.Context(), req.Email, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email verified successfully. You can now log in.",
	})
}

// ResendVerificationEmail handles resending verification code with rate limiting
func (h *AuthHandler) ResendVerificationEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	if err := h.authService.UserService().ResendVerificationEmail(c.Request.Context(), req.Email); err != nil {
		// Return 429 for rate limit errors, 400 for others
		statusCode := http.StatusBadRequest
		if err.Error() == "please wait before requesting another verification code" {
			statusCode = http.StatusTooManyRequests
		}

		c.JSON(statusCode, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Verification code sent successfully. Please check your email.",
	})
}
