package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"auth-service/internal/config"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// JWTService defines the interface for JWT operations
type JWTService interface {
	GenerateAccessToken(ctx *TokenContext) (string, error)
	GenerateRefreshToken(ctx *TokenContext) (string, string, error)
	GenerateOAuthAccessToken(ctx *OAuthTokenContext) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	ParseAccessToken(tokenString string) (*Claims, error)
	ParseRefreshToken(tokenString string) (*Claims, error)
}

// TokenContext carries metadata for org-scoped token generation (Slack style)
type TokenContext struct {
	UserID           uuid.UUID
	OrganizationID   uuid.UUID
	SessionID        uuid.UUID
	RoleID           uuid.UUID
	Email            string
	GlobalRole       string
	OrganizationRole string
	Permissions      []string // List of permission names for this user in this org
	IsSuperadmin     bool
}

// OAuthTokenContext carries metadata for OAuth2 token generation
type OAuthTokenContext struct {
	UserID         uuid.UUID
	Email          string
	OrganizationID *uuid.UUID // Optional
	Roles          []string   // User's role names in the organization
	Permissions    []string   // Scopes/permissions
	Issuer         string     // https://auth.myservice.com/{client_id}
	Audience       string     // client_id
	Subject        string     // user_id
	IsSuperadmin   bool
}

// Claims represents the JWT claims stored in access/refresh tokens
type Claims struct {
	UserID           uuid.UUID  `json:"user_id"`
	OrganizationID   uuid.UUID  `json:"organization_id,omitempty"`
	SessionID        uuid.UUID  `json:"session_id,omitempty"`
	RoleID           uuid.UUID  `json:"role_id,omitempty"`
	Email            string     `json:"email"`
	GlobalRole       string     `json:"global_role,omitempty"`
	OrganizationRole string     `json:"organization_role,omitempty"`
	Roles            []string   `json:"roles,omitempty"`       // OAuth2 role names
	Permissions      []string   `json:"permissions,omitempty"` // Cached permission names
	Scope            string     `json:"scope,omitempty"`       // OAuth2 scopes (space-separated)
	IsSuperadmin     bool       `json:"is_superadmin"`
	TokenType        string     `json:"token_type"`
	Org              *uuid.UUID `json:"org,omitempty"` // OAuth2 org claim
	jwt.RegisteredClaims
}

// Service handles JWT operations using RSA keys
type Service struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	config     *config.JWTConfig
}

// NewService creates a new JWT service (dev uses runtime RSA key)
func NewService(cfg *config.JWTConfig) (*Service, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	return &Service{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		config:     cfg,
	}, nil
}

// GenerateAccessToken creates a short-lived org-scoped access token
func (s *Service) GenerateAccessToken(ctxInput *TokenContext) (string, error) {
	if ctxInput == nil {
		return "", errors.New("token context is required")
	}

	now := time.Now()
	exp := now.Add(time.Duration(s.config.AccessTokenTTL) * time.Minute)

	claims := &Claims{
		UserID:           ctxInput.UserID,
		OrganizationID:   ctxInput.OrganizationID,
		SessionID:        ctxInput.SessionID,
		RoleID:           ctxInput.RoleID,
		Email:            ctxInput.Email,
		GlobalRole:       ctxInput.GlobalRole,
		OrganizationRole: ctxInput.OrganizationRole,
		Permissions:      ctxInput.Permissions,
		IsSuperadmin:     ctxInput.IsSuperadmin,
		TokenType:        "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   ctxInput.UserID.String(),
			Audience:  jwt.ClaimStrings{"auth-service"},
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

// GenerateOAuthAccessToken creates an OAuth2-compliant access token with iss, aud, scope, roles, permissions
func (s *Service) GenerateOAuthAccessToken(ctxInput *OAuthTokenContext) (string, error) {
	if ctxInput == nil {
		return "", errors.New("token context is required")
	}

	now := time.Now()
	exp := now.Add(1 * time.Hour) // OAuth tokens typically 1 hour

	claims := &Claims{
		UserID:       ctxInput.UserID,
		Email:        ctxInput.Email,
		IsSuperadmin: ctxInput.IsSuperadmin,
		TokenType:    "access",
		Roles:        ctxInput.Roles,                          // RBAC role names
		Permissions:  ctxInput.Permissions,                    // RBAC permission names
		Scope:        strings.Join(ctxInput.Permissions, " "), // OAuth2 scope (space-separated permissions)
		Org:          ctxInput.OrganizationID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    ctxInput.Issuer,                     // https://auth.myservice.com/{client_id}
			Subject:   ctxInput.Subject,                    // user_id
			Audience:  jwt.ClaimStrings{ctxInput.Audience}, // client_id
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

// GenerateRefreshToken creates a long-lived refresh token bound to session + org
func (s *Service) GenerateRefreshToken(ctxInput *TokenContext) (string, string, error) {
	if ctxInput == nil {
		return "", "", errors.New("token context is required")
	}

	now := time.Now()
	exp := now.AddDate(0, 0, s.config.RefreshTokenTTL)
	refreshID := uuid.New().String()

	claims := &Claims{
		UserID:           ctxInput.UserID,
		OrganizationID:   ctxInput.OrganizationID,
		SessionID:        ctxInput.SessionID,
		RoleID:           ctxInput.RoleID,
		Email:            ctxInput.Email,
		GlobalRole:       ctxInput.GlobalRole,
		OrganizationRole: ctxInput.OrganizationRole,
		Permissions:      ctxInput.Permissions,
		IsSuperadmin:     ctxInput.IsSuperadmin,
		TokenType:        "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   ctxInput.UserID.String(),
			Audience:  jwt.ClaimStrings{"auth-service"},
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        refreshID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, refreshID, nil
}

// ValidateToken parses a JWT and extracts its claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if c, ok := token.Claims.(*Claims); ok && token.Valid {
		// Handle nil expiration gracefully
		if c.ExpiresAt != nil && time.Now().After(c.ExpiresAt.Time) {
			return nil, errors.New("token expired")
		}
		return c, nil
	}

	return nil, errors.New("invalid token")
}

// ParseAccessToken ensures token type === access
func (s *Service) ParseAccessToken(tokenString string) (*Claims, error) {
	c, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if c.TokenType != "access" {
		return nil, errors.New("token is not an access token")
	}
	return c, nil
}

// ParseRefreshToken ensures token type === refresh
func (s *Service) ParseRefreshToken(tokenString string) (*Claims, error) {
	c, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if c.TokenType != "refresh" {
		return nil, errors.New("token is not a refresh token")
	}
	return c, nil
}

// RefreshAccessToken regenerates a new access token from a refresh token
func (s *Service) RefreshAccessToken(refreshTokenString string) (string, error) {
	c, err := s.ParseRefreshToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	return s.GenerateAccessToken(&TokenContext{
		UserID:           c.UserID,
		OrganizationID:   c.OrganizationID,
		SessionID:        c.SessionID,
		RoleID:           c.RoleID,
		Email:            c.Email,
		GlobalRole:       c.GlobalRole,
		OrganizationRole: c.OrganizationRole,
		Permissions:      c.Permissions,
		IsSuperadmin:     c.IsSuperadmin,
	})
}

// Helpers
func (s *Service) ExtractUserID(tokenString string) (uuid.UUID, error) {
	c, err := s.ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}
	return c.UserID, nil
}

func (s *Service) ExtractEmail(tokenString string) (string, error) {
	c, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return c.Email, nil
}

func (s *Service) IsSuperadmin(tokenString string) (bool, error) {
	c, err := s.ValidateToken(tokenString)
	if err != nil {
		return false, err
	}
	return c.IsSuperadmin, nil
}

func (s *Service) IsTokenExpired(tokenString string) bool {
	c, err := s.ValidateToken(tokenString)
	if err != nil {
		return true
	}
	if c.ExpiresAt == nil {
		return true
	}
	return time.Now().After(c.ExpiresAt.Time)
}

func (s *Service) GetTokenExpiration(tokenString string) (time.Time, error) {
	c, err := s.ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}
	if c.ExpiresAt == nil {
		return time.Time{}, errors.New("token has no expiration claim")
	}
	return c.ExpiresAt.Time, nil
}
