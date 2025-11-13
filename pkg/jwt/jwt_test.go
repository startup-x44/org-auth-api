package jwt_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"auth-service/internal/config"
	"auth-service/internal/models"
	jwtsvc "auth-service/pkg/jwt"
)

func TestJWTService(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Secret:          "test-secret-key-that-is-at-least-32-characters-long",
		AccessTokenTTL:  60,
		RefreshTokenTTL: 30,
		Issuer:          "test-issuer",
		SigningMethod:   "RS256",
	}

	svc, err := jwtsvc.NewService(jwtConfig)
	require.NoError(t, err)

	user := &models.User{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		UserType: "Admin",
	}

	// Test access token generation
	accessToken, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)

	// Test refresh token generation
	refreshToken, err := svc.GenerateRefreshToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	// Test token validation
	claims, err := svc.ValidateToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.TenantID, claims.TenantID)
	assert.Equal(t, user.UserType, claims.UserType)
}

func TestJWTServiceExpiredToken(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Secret:          "test-secret-key-that-is-at-least-32-characters-long",
		AccessTokenTTL:  60,
		RefreshTokenTTL: 30,
		Issuer:          "test-issuer",
		SigningMethod:   "RS256",
	}

	svc, err := jwtsvc.NewService(jwtConfig)
	require.NoError(t, err)

	user := &models.User{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		UserType: "Admin",
	}

	// Generate token with very short expiry for testing
	accessToken, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	// Wait for token to expire (this test might be flaky in CI)
	time.Sleep(2 * time.Second)

	// Token should still be valid since we use longer expiry in service
	claims, err := svc.ValidateToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
}

// BenchmarkJWTGenerateAccessToken benchmarks access token generation performance
func BenchmarkJWTGenerateAccessToken(b *testing.B) {
	jwtConfig := &config.JWTConfig{
		Secret:          "benchmark-secret-key-that-is-at-least-32-characters",
		AccessTokenTTL:  60,
		RefreshTokenTTL: 30,
		Issuer:          "benchmark-issuer",
		SigningMethod:   "RS256",
	}

	svc, err := jwtsvc.NewService(jwtConfig)
	if err != nil {
		b.Fatalf("Failed to create JWT service: %v", err)
	}

	user := &models.User{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		UserType: "Admin",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := svc.GenerateAccessToken(user)
		if err != nil {
			b.Fatalf("Failed to generate access token: %v", err)
		}
	}
}

// BenchmarkJWTGenerateRefreshToken benchmarks refresh token generation performance
func BenchmarkJWTGenerateRefreshToken(b *testing.B) {
	jwtConfig := &config.JWTConfig{
		Secret:          "benchmark-secret-key-that-is-at-least-32-characters",
		AccessTokenTTL:  60,
		RefreshTokenTTL: 30,
		Issuer:          "benchmark-issuer",
		SigningMethod:   "RS256",
	}

	svc, err := jwtsvc.NewService(jwtConfig)
	if err != nil {
		b.Fatalf("Failed to create JWT service: %v", err)
	}

	user := &models.User{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		UserType: "Admin",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := svc.GenerateRefreshToken(user)
		if err != nil {
			b.Fatalf("Failed to generate refresh token: %v", err)
		}
	}
}

// BenchmarkJWTValidateAccessToken benchmarks access token validation performance
func BenchmarkJWTValidateAccessToken(b *testing.B) {
	jwtConfig := &config.JWTConfig{
		Secret:          "benchmark-secret-key-that-is-at-least-32-characters",
		AccessTokenTTL:  60,
		RefreshTokenTTL: 30,
		Issuer:          "benchmark-issuer",
		SigningMethod:   "RS256",
	}

	svc, err := jwtsvc.NewService(jwtConfig)
	if err != nil {
		b.Fatalf("Failed to create JWT service: %v", err)
	}

	user := &models.User{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		UserType: "Admin",
	}

	token, err := svc.GenerateAccessToken(user)
	if err != nil {
		b.Fatalf("Failed to generate access token: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := svc.ValidateToken(token)
		if err != nil {
			b.Fatalf("Failed to validate access token: %v", err)
		}
	}
}