package handler

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"

	"auth-service/internal/service"

	"github.com/gin-gonic/gin"
)

// OAuth2ConsentHandler handles the traditional OAuth2 consent flow with HTML forms
type OAuth2ConsentHandler struct {
	oauth2Service service.OAuth2Service
	clientAppSvc  service.ClientAppService
	userService   service.UserService
}

// NewOAuth2ConsentHandler creates a new consent handler
func NewOAuth2ConsentHandler(
	oauth2Service service.OAuth2Service,
	clientAppSvc service.ClientAppService,
	userService service.UserService,
) *OAuth2ConsentHandler {
	return &OAuth2ConsentHandler{
		oauth2Service: oauth2Service,
		clientAppSvc:  clientAppSvc,
		userService:   userService,
	}
}

// ShowConsentForm godoc
// @Summary Show OAuth2 authorization/login form (GET /oauth/authorize)
// @Description Displays consent form for user to authorize the OAuth2 client
// @Tags oauth2
// @Produce html
// @Param client_id query string true "Client ID"
// @Param redirect_uri query string true "Redirect URI"
// @Param response_type query string true "Response type (must be 'code')"
// @Param scope query string false "Requested scopes (space-separated)"
// @Param state query string false "State parameter for CSRF protection"
// @Param code_challenge query string true "PKCE code challenge (S256)"
// @Param code_challenge_method query string true "PKCE method (must be 'S256')"
// @Success 200 {string} html "HTML consent/login form"
// @Failure 400 {object} gin.H{error=string}
// @Router /oauth/authorize [get]
func (h *OAuth2ConsentHandler) ShowConsentForm(c *gin.Context) {
	// 1. Extract and validate OAuth2 parameters
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	scope := c.Query("scope")
	state := c.Query("state")
	codeChallenge := c.Query("code_challenge")
	codeChallengeMethod := c.Query("code_challenge_method")

	// Validate required parameters
	if clientID == "" || redirectURI == "" || responseType == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":       "invalid_request",
			"description": "Missing required parameters: client_id, redirect_uri, or response_type",
		})
		return
	}

	// Validate response_type
	if responseType != "code" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":       "unsupported_response_type",
			"description": "Only 'code' response type is supported",
		})
		return
	}

	// Validate PKCE (required for security)
	if codeChallenge == "" || codeChallengeMethod != "S256" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":       "invalid_request",
			"description": "PKCE with S256 method is required",
		})
		return
	}

	// 2. Validate client_id and redirect_uri against database
	if err := h.clientAppSvc.ValidateRedirectURI(c.Request.Context(), clientID, redirectURI); err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":       "invalid_client",
			"description": "Invalid client_id or redirect_uri: " + err.Error(),
		})
		return
	}

	// 3. Get client app details to show in consent form
	clientApp, err := h.clientAppSvc.GetClientAppByClientID(c.Request.Context(), clientID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":       "invalid_client",
			"description": "Client application not found",
		})
		return
	}

	// 4. Validate requested scopes against allowed scopes
	requestedScopes := parseScopes(scope)
	if !validateScopes(requestedScopes, clientApp.AllowedScopes) {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":       "invalid_scope",
			"description": "One or more requested scopes are not allowed for this client",
		})
		return
	}

	// 5. Generate CSRF token for form protection
	csrfToken, err := generateCSRFToken()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error":       "server_error",
			"description": "Failed to generate CSRF token",
		})
		return
	}

	// Store CSRF token in session/cookie
	c.SetCookie("oauth_csrf", csrfToken, 600, "/", "", false, true) // 10 min expiry, HttpOnly

	// 6. Render HTML consent/login form
	c.HTML(http.StatusOK, "oauth_consent.html", gin.H{
		"client_name":           clientApp.Name,
		"client_id":             clientID,
		"redirect_uri":          redirectURI,
		"response_type":         responseType,
		"scope":                 scope,
		"scopes":                requestedScopes,
		"state":                 state,
		"code_challenge":        codeChallenge,
		"code_challenge_method": codeChallengeMethod,
		"csrf_token":            csrfToken,
		"organization_required": true, // User must belong to client's org
	})
}

// ProcessConsent godoc
// @Summary Process OAuth2 authorization form submission (POST /oauth/authorize)
// @Description Authenticates user and generates authorization code
// @Tags oauth2
// @Accept application/x-www-form-urlencoded
// @Produce html
// @Param email formData string true "User email"
// @Param password formData string true "User password"
// @Param client_id formData string true "Client ID"
// @Param redirect_uri formData string true "Redirect URI"
// @Param response_type formData string true "Response type"
// @Param scope formData string false "Requested scopes"
// @Param state formData string false "State parameter"
// @Param code_challenge formData string true "PKCE code challenge"
// @Param code_challenge_method formData string true "PKCE method"
// @Param csrf_token formData string true "CSRF token"
// @Success 302 {string} string "Redirects to redirect_uri with authorization code"
// @Failure 400 {object} gin.H{error=string}
// @Failure 401 {object} gin.H{error=string}
// @Failure 403 {object} gin.H{error=string}
// @Router /oauth/authorize [post]
func (h *OAuth2ConsentHandler) ProcessConsent(c *gin.Context) {
	// 1. Extract form parameters
	email := c.PostForm("email")
	password := c.PostForm("password")
	clientID := c.PostForm("client_id")
	redirectURI := c.PostForm("redirect_uri")
	responseType := c.PostForm("response_type")
	scope := c.PostForm("scope")
	state := c.PostForm("state")
	codeChallenge := c.PostForm("code_challenge")
	codeChallengeMethod := c.PostForm("code_challenge_method")
	csrfToken := c.PostForm("csrf_token")

	// 2. Validate CSRF token
	cookieCSRF, err := c.Cookie("oauth_csrf")
	if err != nil || cookieCSRF != csrfToken || csrfToken == "" {
		redirectErrorHTML(c, redirectURI, "invalid_request", "CSRF token validation failed", state)
		return
	}

	// Clear CSRF cookie after use
	c.SetCookie("oauth_csrf", "", -1, "/", "", false, true)

	// 3. Validate required parameters
	if email == "" || password == "" || clientID == "" || redirectURI == "" {
		redirectErrorHTML(c, redirectURI, "invalid_request", "Missing required parameters", state)
		return
	}

	if responseType != "code" {
		redirectErrorHTML(c, redirectURI, "unsupported_response_type", "Only 'code' response type is supported", state)
		return
	}

	// 4. Validate PKCE
	if codeChallenge == "" || codeChallengeMethod != "S256" {
		redirectErrorHTML(c, redirectURI, "invalid_request", "PKCE with S256 method is required", state)
		return
	}

	// 5. Validate client and redirect URI
	if err := h.clientAppSvc.ValidateRedirectURI(c.Request.Context(), clientID, redirectURI); err != nil {
		redirectErrorHTML(c, redirectURI, "invalid_client", "Invalid client or redirect URI", state)
		return
	}

	// 6. Get client app to check organization ownership
	clientApp, err := h.clientAppSvc.GetClientAppByClientID(c.Request.Context(), clientID)
	if err != nil {
		redirectErrorHTML(c, redirectURI, "invalid_client", "Client not found", state)
		return
	}

	// 7. Authenticate user with email and password (using Argon2/bcrypt)
	user, err := h.userService.AuthenticateByEmail(c.Request.Context(), email, password)
	if err != nil {
		// Return to form with error - don't expose whether email exists
		c.HTML(http.StatusUnauthorized, "oauth_consent.html", gin.H{
			"error":                 "invalid_credentials",
			"error_description":     "Invalid email or password",
			"client_name":           clientApp.Name,
			"client_id":             clientID,
			"redirect_uri":          redirectURI,
			"response_type":         responseType,
			"scope":                 scope,
			"state":                 state,
			"code_challenge":        codeChallenge,
			"code_challenge_method": codeChallengeMethod,
			"email":                 email, // Pre-fill email on error
		})
		return
	}

	// 8. **CRITICAL**: Verify user belongs to the organization that owns this OAuth2 client
	isMember, err := h.userService.IsOrgMember(c.Request.Context(), user.ID, clientApp.OrganizationID)
	if err != nil || !isMember {
		// User authenticated successfully but doesn't belong to the required organization
		c.HTML(http.StatusForbidden, "oauth_consent.html", gin.H{
			"error":             "access_denied",
			"error_description": "You are not a member of the organization that owns this application",
			"client_name":       clientApp.Name,
		})
		return
	}

	// 9. Create authorization code (short-lived: 10 minutes)
	authReq := &service.AuthorizationRequest{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		UserID:              user.ID,
		OrganizationID:      &clientApp.OrganizationID,
	}

	code, err := h.oauth2Service.CreateAuthorizationCode(c.Request.Context(), authReq)
	if err != nil {
		redirectErrorHTML(c, redirectURI, "server_error", "Failed to create authorization code", state)
		return
	}

	// 10. Build redirect URL with authorization code
	redirectURL, _ := url.Parse(redirectURI)
	q := redirectURL.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	redirectURL.RawQuery = q.Encode()

	// 11. Redirect user back to client application with authorization code
	c.Redirect(http.StatusFound, redirectURL.String())
}

// Helper: Generate cryptographically secure CSRF token
func generateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// Helper: Parse space-separated scopes
func parseScopes(scopeStr string) []string {
	if scopeStr == "" {
		return []string{}
	}
	scopes := []string{}
	for _, s := range splitBySpace(scopeStr) {
		if s != "" {
			scopes = append(scopes, s)
		}
	}
	return scopes
}

func splitBySpace(s string) []string {
	result := []string{}
	current := ""
	for _, char := range s {
		if char == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// Helper: Validate requested scopes against allowed scopes
func validateScopes(requested []string, allowed []string) bool {
	if len(requested) == 0 {
		return true // No scopes requested is valid
	}

	// If no allowed scopes are configured, allow any scopes (unrestricted client)
	if len(allowed) == 0 {
		return true
	}

	allowedMap := make(map[string]bool)
	for _, scope := range allowed {
		allowedMap[scope] = true
	}

	for _, scope := range requested {
		if !allowedMap[scope] {
			return false
		}
	}
	return true
}

// Helper: Redirect with OAuth2 error
func redirectErrorHTML(c *gin.Context, redirectURI, errorCode, description, state string) {
	redirectURL, err := url.Parse(redirectURI)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":       errorCode,
			"description": description,
		})
		return
	}

	q := redirectURL.Query()
	q.Set("error", errorCode)
	q.Set("error_description", description)
	if state != "" {
		q.Set("state", state)
	}
	redirectURL.RawQuery = q.Encode()

	c.Redirect(http.StatusFound, redirectURL.String())
}
