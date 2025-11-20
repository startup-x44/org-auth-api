package test

import (
	"context"
	"testing"
	"time"

	"auth-service/internal/models"
	"auth-service/internal/service"
	"auth-service/pkg/pkce"
	"auth-service/tests/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthFlow_Integration(t *testing.T) {
	// Setup test database and services
	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup()

	// Create test user and organization
	testTenant := testutils.CreateTestTenant(t, testDB.DB)
	testUser := testutils.CreateTestUser(t, testDB.DB, testTenant.ID)

	// Initialize services
	userRepo := testutils.NewTestUserRepository(testDB.DB)
	clientAppSvc := service.NewClientAppService(testutils.NewTestClientAppRepository(testDB.DB))
	oauth2Svc := service.NewOAuth2Service(
		testutils.NewTestOAuth2Repository(testDB.DB),
		clientAppSvc,
		userRepo,
	)

	ctx := context.Background()

	t.Run("Complete OAuth2.1 Authorization Code + PKCE Flow", func(t *testing.T) {
		// Step 1: Create a test client application
		clientApp, err := clientAppSvc.CreateClientApp(ctx, testUser.ID, testTenant.ID, &service.CreateClientAppRequest{
			Name:          "Test OAuth Client",
			Description:   "Integration test client",
			RedirectURIs:  []string{"http://localhost:3000/callback"},
			AllowedScopes: []string{"profile", "email"},
		})
		require.NoError(t, err)
		require.NotEmpty(t, clientApp.ClientID)
		require.NotEmpty(t, clientApp.ClientSecret)

		// Step 2: Generate PKCE parameters
		pkcePair, err := pkce.GeneratePKCEPair()
		require.NoError(t, err)
		require.NotEmpty(t, pkcePair.CodeVerifier)
		require.NotEmpty(t, pkcePair.CodeChallenge)
		require.Equal(t, "S256", pkcePair.CodeChallengeMethod)

		// Step 3: Create authorization request
		authReq := &service.AuthorizeRequest{
			ClientID:            clientApp.ClientID,
			RedirectURI:         "http://localhost:3000/callback",
			ResponseType:        "code",
			Scope:               "profile email",
			State:               "test-state-" + uuid.New().String()[:8],
			CodeChallenge:       pkcePair.CodeChallenge,
			CodeChallengeMethod: pkcePair.CodeChallengeMethod,
		}

		// Step 4: Generate authorization code (simulating user consent)
		authCode, err := oauth2Svc.CreateAuthorizationCode(ctx, testUser.ID, authReq)
		require.NoError(t, err)
		require.NotEmpty(t, authCode.Code)
		require.Equal(t, authReq.ClientID, authCode.ClientID)
		require.Equal(t, authReq.RedirectURI, authCode.RedirectURI)
		require.Equal(t, authReq.Scope, authCode.Scopes)
		require.Equal(t, authReq.CodeChallenge, authCode.CodeChallenge)

		// Step 5: Exchange authorization code for tokens
		tokenReq := &service.TokenRequest{
			GrantType:    "authorization_code",
			Code:         authCode.Code,
			ClientID:     clientApp.ClientID,
			ClientSecret: clientApp.ClientSecret,
			RedirectURI:  authReq.RedirectURI,
			CodeVerifier: pkcePair.CodeVerifier,
		}

		tokenResp, err := oauth2Svc.ExchangeCodeForTokens(ctx, tokenReq)
		require.NoError(t, err)
		require.NotEmpty(t, tokenResp.AccessToken)
		require.NotEmpty(t, tokenResp.RefreshToken)
		require.Equal(t, "Bearer", tokenResp.TokenType)
		require.Greater(t, tokenResp.ExpiresIn, int64(3500)) // Should be ~3600 seconds

		// Step 6: Verify authorization code is marked as used
		usedCode, err := oauth2Svc.GetAuthorizationCode(ctx, authCode.Code)
		require.NoError(t, err)
		require.True(t, usedCode.Used)
		require.NotNil(t, usedCode.UsedAt)

		// Step 7: Attempt to reuse authorization code (should fail)
		_, err = oauth2Svc.ExchangeCodeForTokens(ctx, tokenReq)
		require.Error(t, err)
		require.Contains(t, err.Error(), "authorization code has already been used")

		// Step 8: Test refresh token flow
		refreshReq := &service.TokenRequest{
			GrantType:    "refresh_token",
			RefreshToken: tokenResp.RefreshToken,
			ClientID:     clientApp.ClientID,
			ClientSecret: clientApp.ClientSecret,
		}

		newTokenResp, err := oauth2Svc.RefreshTokens(ctx, refreshReq)
		require.NoError(t, err)
		require.NotEmpty(t, newTokenResp.AccessToken)
		require.NotEmpty(t, newTokenResp.RefreshToken)
		require.NotEqual(t, tokenResp.AccessToken, newTokenResp.AccessToken)
		require.NotEqual(t, tokenResp.RefreshToken, newTokenResp.RefreshToken) // Refresh token rotation

		t.Logf("✅ OAuth2.1 + PKCE flow completed successfully")
		t.Logf("   - Client App: %s", clientApp.ClientID)
		t.Logf("   - Authorization Code: %s", authCode.Code[:8]+"...")
		t.Logf("   - Access Token: %s", tokenResp.AccessToken[:20]+"...")
		t.Logf("   - Refresh Token: %s", tokenResp.RefreshToken[:20]+"...")
	})

	t.Run("Invalid PKCE Code Verifier", func(t *testing.T) {
		// Create client app
		clientApp, err := clientAppSvc.CreateClientApp(ctx, testUser.ID, testTenant.ID, &service.CreateClientAppRequest{
			Name:          "Test Invalid PKCE Client",
			Description:   "Test client for invalid PKCE",
			RedirectURIs:  []string{"http://localhost:3000/callback"},
			AllowedScopes: []string{"profile"},
		})
		require.NoError(t, err)

		// Generate PKCE parameters
		pkcePair, err := pkce.GeneratePKCEPair()
		require.NoError(t, err)

		// Create authorization code
		authReq := &service.AuthorizeRequest{
			ClientID:            clientApp.ClientID,
			RedirectURI:         "http://localhost:3000/callback",
			ResponseType:        "code",
			Scope:               "profile",
			CodeChallenge:       pkcePair.CodeChallenge,
			CodeChallengeMethod: pkcePair.CodeChallengeMethod,
		}

		authCode, err := oauth2Svc.CreateAuthorizationCode(ctx, testUser.ID, authReq)
		require.NoError(t, err)

		// Attempt token exchange with wrong code verifier
		tokenReq := &service.TokenRequest{
			GrantType:    "authorization_code",
			Code:         authCode.Code,
			ClientID:     clientApp.ClientID,
			ClientSecret: clientApp.ClientSecret,
			RedirectURI:  authReq.RedirectURI,
			CodeVerifier: "wrong-code-verifier",
		}

		_, err = oauth2Svc.ExchangeCodeForTokens(ctx, tokenReq)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid code verifier")

		t.Logf("✅ Invalid PKCE code verifier correctly rejected")
	})

	t.Run("Authorization Code Expiration", func(t *testing.T) {
		// Create client app
		clientApp, err := clientAppSvc.CreateClientApp(ctx, testUser.ID, testTenant.ID, &service.CreateClientAppRequest{
			Name:          "Test Expiration Client",
			Description:   "Test client for code expiration",
			RedirectURIs:  []string{"http://localhost:3000/callback"},
			AllowedScopes: []string{"profile"},
		})
		require.NoError(t, err)

		// Generate PKCE parameters
		pkcePair, err := pkce.GeneratePKCEPair()
		require.NoError(t, err)

		// Create authorization code
		authReq := &service.AuthorizeRequest{
			ClientID:            clientApp.ClientID,
			RedirectURI:         "http://localhost:3000/callback",
			ResponseType:        "code",
			Scope:               "profile",
			CodeChallenge:       pkcePair.CodeChallenge,
			CodeChallengeMethod: pkcePair.CodeChallengeMethod,
		}

		authCode, err := oauth2Svc.CreateAuthorizationCode(ctx, testUser.ID, authReq)
		require.NoError(t, err)

		// Manually expire the code by updating the database
		// In real scenario, we'd wait 10 minutes, but for testing we modify directly
		expiredTime := time.Now().Add(-15 * time.Minute)
		err = testDB.DB.Model(&models.AuthorizationCode{}).
			Where("code = ?", authCode.Code).
			Update("expires_at", expiredTime).Error
		require.NoError(t, err)

		// Attempt token exchange with expired code
		tokenReq := &service.TokenRequest{
			GrantType:    "authorization_code",
			Code:         authCode.Code,
			ClientID:     clientApp.ClientID,
			ClientSecret: clientApp.ClientSecret,
			RedirectURI:  authReq.RedirectURI,
			CodeVerifier: pkcePair.CodeVerifier,
		}

		_, err = oauth2Svc.ExchangeCodeForTokens(ctx, tokenReq)
		require.Error(t, err)
		require.Contains(t, err.Error(), "authorization code has expired")

		t.Logf("✅ Expired authorization code correctly rejected")
	})
}

func TestOAuthFlow_HTTPIntegration(t *testing.T) {
	// This test would require setting up the full HTTP server
	// For now, we'll create a simpler version that tests the handler logic
	t.Skip("HTTP integration test - requires full server setup")

	// TODO: Implement HTTP-level integration test that:
	// 1. Starts the Gin server
	// 2. Makes actual HTTP requests to /oauth/authorize
	// 3. Follows redirects and exchanges codes
	// 4. Validates JWT tokens
}

func TestPKCE_Utilities(t *testing.T) {
	t.Run("Generate PKCE Pair", func(t *testing.T) {
		pair, err := pkce.GeneratePKCEPair()
		require.NoError(t, err)

		// Validate code verifier
		assert.GreaterOrEqual(t, len(pair.CodeVerifier), 43)
		assert.LessOrEqual(t, len(pair.CodeVerifier), 128)
		assert.Regexp(t, `^[A-Za-z0-9\-._~]+$`, pair.CodeVerifier)

		// Validate code challenge
		assert.NotEmpty(t, pair.CodeChallenge)
		assert.Equal(t, "S256", pair.CodeChallengeMethod)

		// Verify challenge can be validated
		isValid := pkce.VerifyPKCE(pair.CodeVerifier, pair.CodeChallenge, "S256")
		assert.True(t, isValid)

		t.Logf("✅ PKCE pair generated and validated")
		t.Logf("   - Code Verifier: %s", pair.CodeVerifier[:20]+"...")
		t.Logf("   - Code Challenge: %s", pair.CodeChallenge[:20]+"...")
	})

	t.Run("Invalid PKCE Verification", func(t *testing.T) {
		pair, err := pkce.GeneratePKCEPair()
		require.NoError(t, err)

		// Test with wrong verifier
		isValid := pkce.VerifyPKCE("wrong-verifier", pair.CodeChallenge, "S256")
		assert.False(t, isValid)

		// Test with wrong method
		isValid = pkce.VerifyPKCE(pair.CodeVerifier, pair.CodeChallenge, "plain")
		assert.False(t, isValid)

		t.Logf("✅ Invalid PKCE verification correctly rejected")
	})
}
