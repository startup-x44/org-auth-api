package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"auth-service/internal/models"
	"auth-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OAuth2Handler handles OAuth2 authorization endpoints
type OAuth2Handler struct {
	oauth2Service service.OAuth2Service
	clientAppSvc  service.ClientAppService
	userService   service.UserService
}

// NewOAuth2Handler creates a new OAuth2 handler
func NewOAuth2Handler(
	oauth2Service service.OAuth2Service,
	clientAppSvc service.ClientAppService,
	userService service.UserService,
) *OAuth2Handler {
	return &OAuth2Handler{
		oauth2Service: oauth2Service,
		clientAppSvc:  clientAppSvc,
		userService:   userService,
	}
}

// Authorize godoc
// @Summary OAuth2 authorization endpoint with PKCE
// @Tags oauth2
// @Produce json
// @Param client_id query string true "Client ID"
// @Param redirect_uri query string true "Redirect URI"
// @Param response_type query string true "Response type (must be 'code')"
// @Param scope query string false "Requested scopes"
// @Param state query string false "State parameter"
// @Param code_challenge query string true "PKCE code challenge (S256)"
// @Param code_challenge_method query string true "PKCE method (must be 'S256')"
// @Param prompt query string false "Prompt type (login, none)"
// @Success 302 {string} string "Redirects to redirect_uri with authorization code"
// @Failure 400 {object} gin.H{error=string}
// @Failure 401 {object} gin.H{error=string}
// @Router /oauth/authorize [get]
func (h *OAuth2Handler) Authorize(c *gin.Context) {
	// Extract query parameters
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	scope := c.Query("scope")
	state := c.Query("state")
	codeChallenge := c.Query("code_challenge")
	codeChallengeMethod := c.Query("code_challenge_method")
	prompt := c.Query("prompt")

	// Validate required parameters
	if clientID == "" || redirectURI == "" || responseType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "missing required parameters",
		})
		return
	}

	if responseType != "code" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "unsupported_response_type",
			"error_description": "only 'code' response type is supported",
		})
		return
	}

	// Validate PKCE parameters (required)
	if codeChallenge == "" || codeChallengeMethod != "S256" {
		redirectError(c, redirectURI, "invalid_request", "PKCE with S256 method is required", state)
		return
	}

	// Validate client and redirect URI
	if err := h.clientAppSvc.ValidateRedirectURI(c.Request.Context(), clientID, redirectURI); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": err.Error(),
		})
		return
	}

	// Check if user is authenticated
	userVal, exists := c.Get("user")
	if !exists || prompt == "login" {
		// Redirect to login page with return URL
		loginURL := fmt.Sprintf("/auth/login?return_to=%s", url.QueryEscape(c.Request.URL.String()))
		c.Redirect(http.StatusFound, loginURL)
		return
	}

	user := userVal.(*models.User)

	// Get organization context if available
	var orgID *uuid.UUID
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		parsedOrgID, err := uuid.Parse(orgIDStr)
		if err == nil {
			orgID = &parsedOrgID
		}
	}

	// Create authorization code
	authReq := &service.AuthorizationRequest{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		UserID:              user.ID,
		OrganizationID:      orgID,
	}

	code, err := h.oauth2Service.CreateAuthorizationCode(c.Request.Context(), authReq)
	if err != nil {
		redirectError(c, redirectURI, "server_error", err.Error(), state)
		return
	}

	// Build redirect URL with authorization code
	redirectURL, _ := url.Parse(redirectURI)
	q := redirectURL.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	redirectURL.RawQuery = q.Encode()

	c.Redirect(http.StatusFound, redirectURL.String())
}

// Token godoc
// @Summary OAuth2 token exchange endpoint
// @Tags oauth2
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param grant_type formData string true "Grant type (authorization_code or refresh_token)"
// @Param code formData string false "Authorization code (required for authorization_code grant)"
// @Param redirect_uri formData string false "Redirect URI (required for authorization_code grant)"
// @Param client_id formData string true "Client ID"
// @Param client_secret formData string false "Client secret (required for confidential clients)"
// @Param code_verifier formData string false "PKCE code verifier (required for authorization_code grant)"
// @Param refresh_token formData string false "Refresh token (required for refresh_token grant)"
// @Success 200 {object} service.TokenResponse
// @Failure 400 {object} gin.H{error=string}
// @Failure 401 {object} gin.H{error=string}
// @Router /oauth/token [post]
func (h *OAuth2Handler) Token(c *gin.Context) {
	grantType := c.PostForm("grant_type")

	switch grantType {
	case "authorization_code":
		h.handleAuthorizationCodeGrant(c)
	case "refresh_token":
		h.handleRefreshTokenGrant(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "unsupported_grant_type",
			"error_description": "only 'authorization_code' and 'refresh_token' grants are supported",
		})
	}
}

func (h *OAuth2Handler) handleAuthorizationCodeGrant(c *gin.Context) {
	code := c.PostForm("code")
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")
	redirectURI := c.PostForm("redirect_uri")
	codeVerifier := c.PostForm("code_verifier")

	if code == "" || clientID == "" || redirectURI == "" || codeVerifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "missing required parameters",
		})
		return
	}

	tokenReq := &service.TokenRequest{
		Code:         code,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		CodeVerifier: codeVerifier,
		GrantType:    "authorization_code",
	}

	tokenResp, err := h.oauth2Service.ExchangeCodeForTokens(c.Request.Context(), tokenReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_grant",
			"error_description": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

func (h *OAuth2Handler) handleRefreshTokenGrant(c *gin.Context) {
	refreshToken := c.PostForm("refresh_token")
	clientID := c.PostForm("client_id")

	if refreshToken == "" || clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "missing required parameters",
		})
		return
	}

	tokenResp, err := h.oauth2Service.RefreshAccessToken(c.Request.Context(), refreshToken, clientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_grant",
			"error_description": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

// UserInfo godoc
// @Summary OAuth2 UserInfo endpoint (returns user claims)
// @Tags oauth2
// @Produce json
// @Security BearerAuth
// @Success 200 {object} service.UserInfoResponse
// @Failure 401 {object} gin.H{error=string}
// @Router /oauth/userinfo [get]
func (h *OAuth2Handler) UserInfo(c *gin.Context) {
	// Get user from JWT token (set by auth middleware)
	userVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "authentication required",
		})
		return
	}

	userID, err := uuid.Parse(userVal.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "invalid user ID in token",
		})
		return
	}

	userInfo, err := h.oauth2Service.GetUserInfo(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":             "user_not_found",
			"error_description": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, userInfo)
}

// Logout godoc
// @Summary OAuth2 logout endpoint (revokes refresh token)
// @Tags oauth2
// @Accept json
// @Produce json
// @Param request body map[string]string true "Logout request with refresh_token"
// @Success 200 {object} gin.H{success=true}
// @Failure 400 {object} gin.H{error=string}
// @Router /oauth/logout [post]
func (h *OAuth2Handler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "refresh_token is required",
		})
		return
	}

	err := h.oauth2Service.RevokeRefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "logged out successfully",
	})
}

// Helper function to redirect with error
func redirectError(c *gin.Context, redirectURI, errorCode, errorDesc, state string) {
	redirectURL, _ := url.Parse(redirectURI)
	q := redirectURL.Query()
	q.Set("error", errorCode)
	q.Set("error_description", errorDesc)
	if state != "" {
		q.Set("state", state)
	}
	redirectURL.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, redirectURL.String())
}
