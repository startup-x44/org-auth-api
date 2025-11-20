package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/hashutil"
	"auth-service/pkg/jwt"
	"auth-service/pkg/password"
	"auth-service/pkg/pkce"

	"github.com/google/uuid"
)

// OAuth2Service defines OAuth2 authorization operations
type OAuth2Service interface {
	CreateAuthorizationCode(ctx context.Context, req *AuthorizationRequest) (string, error)
	ExchangeCodeForTokens(ctx context.Context, req *TokenRequest) (*TokenResponse, error)
	GetUserInfo(ctx context.Context, userID uuid.UUID) (*UserInfoResponse, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RefreshAccessToken(ctx context.Context, refreshToken, clientID, userAgent, ipAddress string) (*TokenResponse, error)
}

type oauth2Service struct {
	repo       repository.Repository
	jwtService jwt.JWTService
	pwdService password.PasswordService
}

// NewOAuth2Service creates a new OAuth2 service
func NewOAuth2Service(repo repository.Repository, jwtService jwt.JWTService) OAuth2Service {
	return &oauth2Service{
		repo:       repo,
		jwtService: jwtService,
		pwdService: password.NewService(),
	}
}

// AuthorizationRequest represents OAuth2 authorization request
type AuthorizationRequest struct {
	ClientID            string
	RedirectURI         string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	UserID              uuid.UUID
	OrganizationID      *uuid.UUID
}

// TokenRequest represents OAuth2 token exchange request
type TokenRequest struct {
	Code         string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	CodeVerifier string
	GrantType    string
	UserAgent    string // For token binding
	IPAddress    string // For token binding
}

// TokenResponse represents OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope,omitempty"`
}

// UserInfoResponse represents OAuth2 userinfo response
type UserInfoResponse struct {
	Sub           string  `json:"sub"`
	Email         string  `json:"email"`
	EmailVerified bool    `json:"email_verified"`
	Name          string  `json:"name"`
	GivenName     string  `json:"given_name,omitempty"`
	FamilyName    string  `json:"family_name,omitempty"`
	Organization  *string `json:"organization,omitempty"`
}

func (s *oauth2Service) CreateAuthorizationCode(ctx context.Context, req *AuthorizationRequest) (string, error) {
	// Validate client
	clientApp, err := s.repo.ClientApp().GetByClientID(ctx, req.ClientID)
	if err != nil {
		return "", errors.New("invalid client")
	}

	// Validate redirect URI
	validRedirect := false
	for _, uri := range clientApp.RedirectURIs {
		if uri == req.RedirectURI {
			validRedirect = true
			break
		}
	}
	if !validRedirect {
		return "", errors.New("invalid redirect URI")
	}

	// Validate PKCE (S256 only)
	if req.CodeChallengeMethod != "S256" {
		return "", errors.New("only S256 code challenge method is supported")
	}
	if req.CodeChallenge == "" {
		return "", errors.New("code challenge is required (PKCE)")
	}

	// Validate scopes
	requestedScopes := strings.Split(req.Scope, " ")
	for _, scope := range requestedScopes {
		allowed := false
		for _, allowedScope := range clientApp.AllowedScopes {
			if scope == allowedScope {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", fmt.Errorf("scope '%s' not allowed for this client", scope)
		}
	}

	// Generate authorization code
	codeBytes := make([]byte, 32)
	if _, err := rand.Read(codeBytes); err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}
	code := base64.RawURLEncoding.EncodeToString(codeBytes)

	// Hash the code for storage (deterministic HMAC-SHA256)
	codeHash, err := hashutil.HMACHash(code)
	if err != nil {
		return "", fmt.Errorf("failed to hash authorization code: %w", err)
	}

	// Store authorization code (expires in 10 minutes)
	authCode := &models.AuthorizationCode{
		CodeHash:            codeHash,
		ClientID:            req.ClientID,
		UserID:              req.UserID,
		OrganizationID:      req.OrganizationID,
		RedirectURI:         req.RedirectURI,
		Scope:               req.Scope,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ExpiresAt:           time.Now().Add(10 * time.Minute),
		Used:                false,
	}

	if err := s.repo.AuthorizationCode().Create(ctx, authCode); err != nil {
		return "", fmt.Errorf("failed to create authorization code: %w", err)
	}

	return code, nil
}

func (s *oauth2Service) ExchangeCodeForTokens(ctx context.Context, req *TokenRequest) (*TokenResponse, error) {
	if req.GrantType != "authorization_code" {
		return nil, errors.New("unsupported grant type")
	}

	// Hash the incoming code for lookup
	codeHash, err := hashutil.HMACHash(req.Code)
	if err != nil {
		return nil, errors.New("invalid authorization code")
	}

	// Get authorization code by hash
	authCode, err := s.repo.AuthorizationCode().GetByCodeHash(ctx, codeHash)
	if err != nil {
		return nil, errors.New("invalid or expired authorization code")
	}

	// Verify not already used
	if authCode.Used {
		return nil, errors.New("authorization code already used")
	}

	// Verify client ID matches
	if authCode.ClientID != req.ClientID {
		return nil, errors.New("client ID mismatch")
	}

	// Verify redirect URI matches
	if authCode.RedirectURI != req.RedirectURI {
		return nil, errors.New("redirect URI mismatch")
	}

	// Verify PKCE code verifier
	if err := pkce.VerifyCodeChallenge(req.CodeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod); err != nil {
		return nil, fmt.Errorf("PKCE verification failed: %w", err)
	}

	// Get client app
	clientApp, err := s.repo.ClientApp().GetByClientID(ctx, req.ClientID)
	if err != nil {
		return nil, errors.New("invalid client")
	}

	// Verify client secret for confidential clients
	if clientApp.IsConfidential {
		if req.ClientSecret == "" {
			return nil, errors.New("client secret required for confidential clients")
		}
		// Validate client credentials
		valid, err := s.pwdService.Verify(req.ClientSecret, clientApp.ClientSecret)
		if err != nil || !valid {
			return nil, errors.New("invalid client credentials")
		}
	}

	// Get user
	user, err := s.repo.User().GetByID(ctx, authCode.UserID.String())
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Mark code as used (use hash for lookup)
	if err := s.repo.AuthorizationCode().MarkAsUsed(ctx, codeHash); err != nil {
		return nil, fmt.Errorf("failed to mark code as used: %w", err)
	}

	// Generate access token with proper claims
	accessToken, err := s.generateAccessToken(ctx, user, authCode.OrganizationID, clientApp, authCode.Scope)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token with binding
	refreshToken, err := s.generateRefreshToken(ctx, user.ID, authCode.OrganizationID, req.ClientID, authCode.Scope, req.UserAgent, req.IPAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		RefreshToken: refreshToken,
		Scope:        authCode.Scope,
	}, nil
}

func (s *oauth2Service) GetUserInfo(ctx context.Context, userID uuid.UUID) (*UserInfoResponse, error) {
	user, err := s.repo.User().GetByID(ctx, userID.String())
	if err != nil {
		return nil, errors.New("user not found")
	}

	firstName := ""
	lastName := ""
	if user.Firstname != nil {
		firstName = *user.Firstname
	}
	if user.Lastname != nil {
		lastName = *user.Lastname
	}

	response := &UserInfoResponse{
		Sub:           user.ID.String(),
		Email:         user.Email,
		EmailVerified: user.EmailVerifiedAt != nil,
		Name:          firstName + " " + lastName,
		GivenName:     firstName,
		FamilyName:    lastName,
	}

	return response, nil
}

func (s *oauth2Service) RevokeRefreshToken(ctx context.Context, token string) error {
	// Hash token for lookup (deterministic HMAC-SHA256)
	tokenHash, err := hashutil.HMACHash(token)
	if err != nil {
		return fmt.Errorf("failed to hash token: %w", err)
	}
	return s.repo.OAuthRefreshToken().Revoke(ctx, tokenHash)
}

func (s *oauth2Service) RefreshAccessToken(ctx context.Context, refreshToken, clientID, userAgent, ipAddress string) (*TokenResponse, error) {
	// Hash the incoming refresh token for lookup (deterministic HMAC-SHA256)
	tokenHash, err := hashutil.HMACHash(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Get refresh token from database
	oauthToken, err := s.repo.OAuthRefreshToken().GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if oauthToken.Revoked {
		return nil, errors.New("refresh token has been revoked")
	}

	// Check if token has already been used (potential replay attack)
	if oauthToken.UsedAt != nil {
		// Token has been used - revoke the entire token family for security
		_ = s.repo.OAuthRefreshToken().RevokeTokenFamily(ctx, oauthToken.FamilyID)
		return nil, errors.New("refresh token has already been used - token family revoked for security")
	}

	if oauthToken.ClientID != clientID {
		return nil, errors.New("client ID mismatch")
	}

	// Verify token binding (user agent and IP)
	userAgentHash := hashutil.SHA256Hash(userAgent)
	ipHash := hashutil.SHA256Hash(ipAddress)

	if oauthToken.UserAgentHash != userAgentHash {
		// User agent mismatch - potential token theft
		_ = s.repo.OAuthRefreshToken().RevokeTokenFamily(ctx, oauthToken.FamilyID)
		return nil, errors.New("token binding violation - user agent mismatch")
	}

	if oauthToken.IPHash != ipHash {
		// IP address mismatch - potential token theft
		// Note: In production, you might want to be more lenient with IP changes
		// (e.g., allow rotation within same /24 subnet, or log warning instead of rejecting)
		_ = s.repo.OAuthRefreshToken().RevokeTokenFamily(ctx, oauthToken.FamilyID)
		return nil, errors.New("token binding violation - IP address mismatch")
	}

	// Get user
	user, err := s.repo.User().GetByID(ctx, oauthToken.UserID.String())
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Get client app
	clientApp, err := s.repo.ClientApp().GetByClientID(ctx, clientID)
	if err != nil {
		return nil, errors.New("invalid client")
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(ctx, user, oauthToken.OrganizationID, clientApp, oauthToken.Scope)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// ===== ATOMIC TRANSACTION: Rotate refresh token =====
	// Begin database transaction for atomic token rotation
	tx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Step 1: Generate new refresh token in same family with same binding
	newRefreshToken, newTokenID, err := s.generateRefreshTokenInTransaction(ctx, tx, oauthToken.FamilyID, oauthToken.UserID, oauthToken.OrganizationID, clientID, oauthToken.Scope, userAgent, ipAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	// Step 2: Mark old refresh token as used and link to new token
	if err := s.markTokenAsUsedInTransaction(ctx, tx, tokenHash, newTokenID); err != nil {
		// If MarkAsUsed fails, it means the token was already used (race condition/replay attack)
		// Transaction will rollback, then revoke the entire token family for security
		_ = s.repo.OAuthRefreshToken().RevokeTokenFamily(ctx, oauthToken.FamilyID)
		return nil, errors.New("refresh token already used - possible replay attack, token family revoked")
	}

	// Commit transaction - both operations succeed atomically
	if err := tx.Commit(); err != nil {
		// Commit failed - revoke family for security
		_ = s.repo.OAuthRefreshToken().RevokeTokenFamily(ctx, oauthToken.FamilyID)
		return nil, fmt.Errorf("failed to commit token rotation: %w", err)
	}
	// ===== END ATOMIC TRANSACTION =====

	return &TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: newRefreshToken, // Return new refresh token
		Scope:        oauthToken.Scope,
	}, nil
}

// Helper functions

func (s *oauth2Service) generateAccessToken(ctx context.Context, user *models.User, orgID *uuid.UUID, clientApp *models.ClientApp, scope string) (string, error) {
	// Get user roles
	roles, err := s.getUserRoles(ctx, user.ID, orgID)
	if err != nil {
		return "", fmt.Errorf("failed to load user roles: %w", err)
	}

	// Build roles list
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// Get user permissions filtered by user type
	permissions, err := s.getUserPermissions(ctx, user, orgID)
	if err != nil {
		return "", fmt.Errorf("failed to load user permissions: %w", err)
	}

	// Build scopes list from permissions
	scopesList := make([]string, len(permissions))
	for i, perm := range permissions {
		scopesList[i] = perm.Name
	}

	// Build issuer with client ID
	issuer := fmt.Sprintf("https://auth.myservice.com/%s", clientApp.ClientID)

	// Generate JWT access token with OAuth2 claims
	tokenCtx := &jwt.OAuthTokenContext{
		UserID:         user.ID,
		Email:          user.Email,
		OrganizationID: orgID,
		Roles:          roleNames,
		Permissions:    scopesList,
		Issuer:         issuer,
		Audience:       clientApp.ClientID,
		Subject:        user.ID.String(),
		IsSuperadmin:   user.IsSuperadmin,
	}

	return s.jwtService.GenerateOAuthAccessToken(tokenCtx)
}

func (s *oauth2Service) generateRefreshToken(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID, clientID, scope, userAgent, ipAddress string) (string, error) {
	// Generate new family ID for this token chain
	familyID := uuid.New()
	token, _, err := s.generateRefreshTokenWithFamilyID(ctx, familyID, userID, orgID, clientID, scope, userAgent, ipAddress)
	return token, err
}

func (s *oauth2Service) generateRefreshTokenWithFamilyID(ctx context.Context, familyID, userID uuid.UUID, orgID *uuid.UUID, clientID, scope, userAgent, ipAddress string) (string, uuid.UUID, error) {
	// Generate random refresh token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", uuid.Nil, err
	}
	refreshToken := base64.RawURLEncoding.EncodeToString(tokenBytes)

	// Hash refresh token for storage (deterministic HMAC-SHA256)
	tokenHash, err := hashutil.HMACHash(refreshToken)
	if err != nil {
		return "", uuid.Nil, fmt.Errorf("failed to hash refresh token: %w", err)
	}

	// Hash user agent and IP for binding (simple SHA256)
	userAgentHash := hashutil.SHA256Hash(userAgent)
	ipHash := hashutil.SHA256Hash(ipAddress)

	// Store in database (expires in 30 days)
	oauthToken := &models.OAuthRefreshToken{
		TokenHash:      tokenHash,
		FamilyID:       familyID,
		ClientID:       clientID,
		UserID:         userID,
		OrganizationID: orgID,
		Scope:          scope,
		UserAgentHash:  userAgentHash,
		IPHash:         ipHash,
		ExpiresAt:      time.Now().Add(30 * 24 * time.Hour),
		Revoked:        false,
	}

	if err := s.repo.OAuthRefreshToken().Create(ctx, oauthToken); err != nil {
		return "", uuid.Nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return refreshToken, oauthToken.ID, nil
}

// generateRefreshTokenInTransaction creates a new refresh token within a transaction
func (s *oauth2Service) generateRefreshTokenInTransaction(ctx context.Context, tx repository.Transaction, familyID, userID uuid.UUID, orgID *uuid.UUID, clientID, scope, userAgent, ipAddress string) (string, uuid.UUID, error) {
	// Generate random refresh token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", uuid.Nil, err
	}
	refreshToken := base64.RawURLEncoding.EncodeToString(tokenBytes)

	// Hash refresh token for storage (deterministic HMAC-SHA256)
	tokenHash, err := hashutil.HMACHash(refreshToken)
	if err != nil {
		return "", uuid.Nil, fmt.Errorf("failed to hash refresh token: %w", err)
	}

	// Hash user agent and IP for binding (simple SHA256)
	userAgentHash := hashutil.SHA256Hash(userAgent)
	ipHash := hashutil.SHA256Hash(ipAddress)

	// Store in database (expires in 30 days)
	oauthToken := &models.OAuthRefreshToken{
		TokenHash:      tokenHash,
		FamilyID:       familyID,
		ClientID:       clientID,
		UserID:         userID,
		OrganizationID: orgID,
		Scope:          scope,
		UserAgentHash:  userAgentHash,
		IPHash:         ipHash,
		ExpiresAt:      time.Now().Add(30 * 24 * time.Hour),
		Revoked:        false,
	}

	if err := tx.OAuthRefreshToken().Create(ctx, oauthToken); err != nil {
		return "", uuid.Nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return refreshToken, oauthToken.ID, nil
}

// markTokenAsUsedInTransaction marks a token as used within a transaction
func (s *oauth2Service) markTokenAsUsedInTransaction(ctx context.Context, tx repository.Transaction, tokenHash string, replacedByID uuid.UUID) error {
	return tx.OAuthRefreshToken().MarkAsUsed(ctx, tokenHash, replacedByID)
}

func (s *oauth2Service) getUserPermissions(ctx context.Context, user *models.User, orgID *uuid.UUID) ([]models.Permission, error) {
	var permissions []models.Permission

	if user.IsSuperadmin {
		// Superadmin gets ONLY system permissions (is_system=true, organization_id IS NULL)
		// CRITICAL: Never return organization-specific permissions to superadmin in this context
		systemPerms, err := s.repo.Permission().ListSystemPermissions(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load system permissions: %w", err)
		}
		for _, perm := range systemPerms {
			permissions = append(permissions, *perm)
		}
		return permissions, nil
	}

	// Non-superadmin user: check organization membership
	if orgID == nil {
		// No organization context - return empty permissions
		return permissions, nil
	}

	// Get user's membership in this organization
	_, err := s.repo.OrganizationMembership().GetByOrganizationAndUser(ctx, orgID.String(), user.ID.String())
	if err != nil {
		// User not a member of this organization - return empty
		return permissions, nil
	}

	// Get user's roles in this organization (only custom roles, is_system=false)
	roles, err := s.getUserRoles(ctx, user.ID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user roles: %w", err)
	}

	// Collect unique permissions from all roles
	permissionMap := make(map[uuid.UUID]models.Permission)
	for _, role := range roles {
		rolePerms, err := s.repo.Permission().ListPermissionsForRole(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load permissions for role %s: %w", role.ID, err)
		}

		// CRITICAL SECURITY: Only include custom org permissions (is_system=false, organization_id=orgID)
		// System permissions should NOT appear in non-superadmin tokens
		for _, perm := range rolePerms {
			if !perm.IsSystem && perm.OrganizationID != nil && *perm.OrganizationID == *orgID {
				permissionMap[perm.ID] = *perm
			}
		}
	}

	// Convert map to slice
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// getUserRoles returns roles for a user based on context
// Superadmin: returns pseudo-role "superadmin"
// Org member: returns only custom roles in that organization (is_system=false)
// No membership: returns empty slice
func (s *oauth2Service) getUserRoles(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) ([]models.Role, error) {
	var roles []models.Role

	// Get user to check superadmin status
	user, err := s.repo.User().GetByID(ctx, userID.String())
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.IsSuperadmin {
		// Return pseudo-role for superadmin
		roles = append(roles, models.Role{
			ID:          uuid.Nil, // Special marker for superadmin
			Name:        "superadmin",
			DisplayName: "Superadmin",
			Description: "Global system administrator",
			IsSystem:    true,
		})
		return roles, nil
	}

	// Non-superadmin: must have org context
	if orgID == nil {
		return roles, nil // Empty slice
	}

	// Verify membership
	membership, err := s.repo.OrganizationMembership().GetByOrganizationAndUser(ctx, orgID.String(), userID.String())
	if err != nil {
		// Not a member - return empty
		return roles, nil
	}

	// Get ONLY custom roles for this organization (is_system=false, organization_id=orgID)
	orgRoles, err := s.repo.Role().ListRolesByOrganizationID(ctx, *orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to load organization roles: %w", err)
	}

	// Filter to roles assigned to this user via membership
	// Note: membership.RoleID contains the user's role in the organization
	for _, role := range orgRoles {
		if role.ID == membership.RoleID {
			roles = append(roles, *role)
		}
	}

	return roles, nil
}
