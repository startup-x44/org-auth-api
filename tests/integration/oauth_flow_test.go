package integration_test

import (
	"context"
	"testing"

	"auth-service/internal/config"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/pkg/jwt"
	"auth-service/pkg/pkce"
	"auth-service/tests/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthFlow_Integration(t *testing.T) {
	// Setup test database and services
	testDB := testutils.SetupTestDB(t)

	// Create test user and organization
	testOrg := testutils.CreateTestOrganization(t, testDB.DB, "Test Organization", "test-org")
	testUser := testutils.CreateTestUser(t, testDB.DB, "test@example.com")

	// Make user a superadmin to create client apps
	testUser.IsSuperadmin = true
	err := testDB.DB.Save(testUser).Error
	require.NoError(t, err)

	// Initialize services
	repo := repository.NewRepository(testDB.DB)
	jwtService, err := jwt.NewService(&config.JWTConfig{
		AccessTokenTTL:  60, // 60 minutes
		RefreshTokenTTL: 1,  // 1 day
		Issuer:          "auth-service-test",
		SigningMethod:   "RS256",
	})
	require.NoError(t, err)
	clientAppSvc := service.NewClientAppService(repo)
	oauth2Svc := service.NewOAuth2Service(repo, jwtService)

	ctx := context.Background()

	t.Run("Complete OAuth2.1 Authorization Code + PKCE Flow", func(t *testing.T) {
		// Step 1: Create a test client application
		clientApp, clientSecret, err := clientAppSvc.CreateClientApp(ctx, testOrg.ID, &service.CreateClientAppRequest{
			Name:           "Test OAuth Client",
			RedirectURIs:   []string{"http://localhost:3000/callback"},
			AllowedScopes:  []string{"profile", "email"},
			IsConfidential: true,
		}, testUser)
		require.NoError(t, err)
		require.NotEmpty(t, clientApp.ClientID)
		require.NotEmpty(t, clientSecret)

		// Step 2: Generate PKCE parameters
		codeVerifier, codeChallenge, err := pkce.GeneratePKCEPair()
		require.NoError(t, err)
		require.NotEmpty(t, codeVerifier)
		require.NotEmpty(t, codeChallenge)

		// Step 3: Create authorization request
		authReq := &service.AuthorizationRequest{
			ClientID:            clientApp.ClientID,
			RedirectURI:         "http://localhost:3000/callback",
			Scope:               "profile email",
			State:               "test-state-" + uuid.New().String()[:8],
			CodeChallenge:       codeChallenge,
			CodeChallengeMethod: "S256",
			UserID:              testUser.ID,
			OrganizationID:      &testOrg.ID,
		}

		// Step 4: Generate authorization code (simulating user consent)
		authCode, err := oauth2Svc.CreateAuthorizationCode(ctx, authReq)
		require.NoError(t, err)
		require.NotEmpty(t, authCode)

		// Step 5: Exchange authorization code for tokens
		tokenReq := &service.TokenRequest{
			GrantType:    "authorization_code",
			Code:         authCode,
			ClientID:     clientApp.ClientID,
			ClientSecret: clientSecret,
			RedirectURI:  authReq.RedirectURI,
			CodeVerifier: codeVerifier,
			UserAgent:    "test-agent",
			IPAddress:    "127.0.0.1",
		}

		tokenResp, err := oauth2Svc.ExchangeCodeForTokens(ctx, tokenReq)
		require.NoError(t, err)
		require.NotEmpty(t, tokenResp.AccessToken)
		require.NotEmpty(t, tokenResp.RefreshToken)
		require.Equal(t, "Bearer", tokenResp.TokenType)
		require.Greater(t, tokenResp.ExpiresIn, int64(3500)) // Should be ~3600 seconds

		// Step 6: Attempt to reuse authorization code (should fail)
		_, err = oauth2Svc.ExchangeCodeForTokens(ctx, tokenReq)
		require.Error(t, err)
		require.Contains(t, err.Error(), "authorization code has already been used")

		// Step 7: Test refresh token flow
		newTokenResp, err := oauth2Svc.RefreshAccessToken(ctx, tokenResp.RefreshToken, clientApp.ClientID, "test-agent", "127.0.0.1")
		require.NoError(t, err)
		require.NotEmpty(t, newTokenResp.AccessToken)
		require.NotEqual(t, tokenResp.AccessToken, newTokenResp.AccessToken)

		t.Logf("✅ OAuth2.1 + PKCE flow completed successfully")
		t.Logf("   - Client App: %s", clientApp.ClientID)
		t.Logf("   - Authorization Code: %s", authCode[:8]+"...")
		t.Logf("   - Access Token: %s", tokenResp.AccessToken[:20]+"...")
		t.Logf("   - Refresh Token: %s", tokenResp.RefreshToken[:20]+"...")
	})

	// Additional test cases could be added here when the service interface is more stable
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
		codeVerifier, codeChallenge, err := pkce.GeneratePKCEPair()
		require.NoError(t, err)

		// Validate code verifier
		assert.GreaterOrEqual(t, len(codeVerifier), 43)
		assert.LessOrEqual(t, len(codeVerifier), 128)
		assert.Regexp(t, `^[A-Za-z0-9\-._~]+$`, codeVerifier)

		// Validate code challenge
		assert.NotEmpty(t, codeChallenge)

		// Verify challenge can be validated
		err = pkce.VerifyCodeChallenge(codeVerifier, codeChallenge, "S256")
		assert.NoError(t, err)

		t.Logf("✅ PKCE pair generated and validated")
		t.Logf("   - Code Verifier: %s", codeVerifier[:20]+"...")
		t.Logf("   - Code Challenge: %s", codeChallenge[:20]+"...")
	})

	t.Run("Invalid PKCE Verification", func(t *testing.T) {
		codeVerifier, codeChallenge, err := pkce.GeneratePKCEPair()
		require.NoError(t, err)

		// Test with wrong verifier
		err = pkce.VerifyCodeChallenge("wrong-verifier", codeChallenge, "S256")
		assert.Error(t, err)

		// Test with wrong method
		err = pkce.VerifyCodeChallenge(codeVerifier, codeChallenge, "plain")
		assert.Error(t, err)

		t.Logf("✅ Invalid PKCE verification correctly rejected")
	})
}
