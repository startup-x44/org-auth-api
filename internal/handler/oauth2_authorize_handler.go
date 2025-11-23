package handler

import (
	"net/http"
	"net/url"

	"auth-service/internal/service"

	"github.com/gin-gonic/gin"
)

// OAuth2AuthorizeWithCredentialsRequest represents the request to authorize with credentials
type OAuth2AuthorizeWithCredentialsRequest struct {
	Email               string `json:"email" binding:"required,email"`
	Password            string `json:"password" binding:"required"`
	ClientID            string `json:"client_id" binding:"required"`
	RedirectURI         string `json:"redirect_uri" binding:"required"`
	ResponseType        string `json:"response_type" binding:"required"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge" binding:"required"`
	CodeChallengeMethod string `json:"code_challenge_method" binding:"required"`
}

// AuthorizeWithCredentials godoc
// @Summary OAuth2 authorization with credentials (for consent page)
// @Description Authenticates user and generates authorization code if user belongs to the OAuth app's organization
// @Tags oauth2
// @Accept json
// @Produce json
// @Param request body OAuth2AuthorizeWithCredentialsRequest true "Authorization request with credentials"
// @Success 200 {object} gin.H{redirect_uri=string}
// @Failure 400 {object} gin.H{error=string}
// @Failure 401 {object} gin.H{error=string}
// @Failure 403 {object} gin.H{error=string}
// @Router /oauth/authorize-with-credentials [post]
func (h *OAuth2Handler) AuthorizeWithCredentials(c *gin.Context) {
	var req OAuth2AuthorizeWithCredentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": err.Error(),
		})
		return
	}

	// Validate response type
	if req.ResponseType != "code" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "unsupported_response_type",
			"error_description": "only 'code' response type is supported",
		})
		return
	}

	// Validate PKCE
	if req.CodeChallenge == "" || req.CodeChallengeMethod != "S256" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "PKCE with S256 method is required",
		})
		return
	}

	// Step 1: Validate OAuth2 client and redirect URI
	if err := h.clientAppSvc.ValidateRedirectURI(c.Request.Context(), req.ClientID, req.RedirectURI); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_client",
			"error_description": "invalid client_id or redirect_uri",
		})
		return
	}

	// Step 2: Get client app to check which organization owns it
	clientApp, err := h.clientAppSvc.GetClientAppByClientID(c.Request.Context(), req.ClientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_client",
			"error_description": "client not found",
		})
		return
	}

	// Step 3: Authenticate user with email/password
	user, err := h.userService.AuthenticateByEmail(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_credentials",
			"error_description": "invalid email or password",
		})
		return
	}

	// Step 4: Check if user belongs to the organization that owns this OAuth2 app
	isMember, err := h.userService.IsOrgMember(c.Request.Context(), user.ID, clientApp.OrganizationID)
	if err != nil || !isMember {
		// User is not a member of the organization that owns this OAuth app
		redirectError(c, req.RedirectURI, "access_denied", "user is not a member of the required organization", req.State)
		return
	}

	// Step 5: Create authorization code
	authReq := &service.AuthorizationRequest{
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURI,
		Scope:               req.Scope,
		State:               req.State,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		UserID:              user.ID,
		OrganizationID:      &clientApp.OrganizationID,
	}

	code, err := h.oauth2Service.CreateAuthorizationCode(c.Request.Context(), authReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "failed to create authorization code",
		})
		return
	}

	// Step 6: Build redirect URL with code
	redirectURL, _ := url.Parse(req.RedirectURI)
	q := redirectURL.Query()
	q.Set("code", code)
	if req.State != "" {
		q.Set("state", req.State)
	}
	redirectURL.RawQuery = q.Encode()

	// Return the redirect URL (frontend will handle the redirect)
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"redirect_uri": redirectURL.String(),
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.Firstname,
			"last_name":  user.Lastname,
		},
	})
}
