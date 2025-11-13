package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"auth-service/internal/config"
	"auth-service/internal/models"
)

// Claims represents the JWT claims
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	TenantID uuid.UUID `json:"tenant_id"`
	UserType string    `json:"user_type"`
	jwt.RegisteredClaims
}

// Service handles JWT operations
type Service struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	config     *config.JWTConfig
}

// NewService creates a new JWT service
func NewService(cfg *config.JWTConfig) (*Service, error) {
	// For development, generate a key pair. In production, load from environment
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

// GenerateAccessToken generates a new access token for a user
func (s *Service) GenerateAccessToken(user *models.User) (string, error) {
	now := time.Now()
	expirationTime := now.Add(time.Duration(s.config.AccessTokenTTL) * time.Minute)

	claims := &Claims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		UserType: user.UserType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   user.ID.String(),
			Audience:  jwt.ClaimStrings{user.TenantID.String()},
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a new refresh token for a user
func (s *Service) GenerateRefreshToken(user *models.User) (string, error) {
	now := time.Now()
	expirationTime := now.AddDate(0, 0, s.config.RefreshTokenTTL)

	claims := &Claims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		UserType: user.UserType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   user.ID.String(),
			Audience:  jwt.ClaimStrings{user.TenantID.String()},
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates and parses a JWT token
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

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshAccessToken validates a refresh token and generates a new access token
func (s *Service) RefreshAccessToken(refreshTokenString string) (string, error) {
	claims, err := s.ValidateToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Create a user object from claims for token generation
	user := &models.User{
		ID:       claims.UserID,
		TenantID: claims.TenantID,
		UserType: claims.UserType,
	}

	return s.GenerateAccessToken(user)
}

// ExtractTenantID extracts tenant ID from token claims
func (s *Service) ExtractTenantID(tokenString string) (uuid.UUID, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.TenantID, nil
}

// ExtractUserID extracts user ID from token claims
func (s *Service) ExtractUserID(tokenString string) (uuid.UUID, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

// ExtractUserType extracts user type from token claims
func (s *Service) ExtractUserType(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserType, nil
}

// IsTokenExpired checks if a token is expired
func (s *Service) IsTokenExpired(tokenString string) bool {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return true
	}
	return time.Now().After(claims.ExpiresAt.Time)
}

// GetTokenExpiration returns the expiration time of a token
func (s *Service) GetTokenExpiration(tokenString string) (time.Time, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}
	return claims.ExpiresAt.Time, nil
}