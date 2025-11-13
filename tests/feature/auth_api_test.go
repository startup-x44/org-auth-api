package feature_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"auth-service/internal/handler"
	"auth-service/internal/repository"
	"auth-service/pkg/jwt"
	"auth-service/pkg/password"
	"auth-service/tests/testutils"
)

func setupTestServer(t *testing.T) (*httptest.Server, func()) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup test database
	testDB := testutils.SetupTestDB(t)

	// Create repositories
	userRepo := repository.NewUserRepository(testDB.DB)
	tenantRepo := repository.NewTenantRepository(testDB.DB)
	sessionRepo := repository.NewUserSessionRepository(testDB.DB)
	refreshTokenRepo := repository.NewRefreshTokenRepository(testDB.DB)
	passwordResetRepo := repository.NewPasswordResetRepository(testDB.DB)
	txRepo := repository.NewTransactionRepository(testDB.DB)

	repo := &repository.RepositoryImpl{
		UserRepo:         userRepo,
		TenantRepo:       tenantRepo,
		UserSessionRepo:  sessionRepo,
		RefreshTokenRepo: refreshTokenRepo,
		PasswordResetRepo: passwordResetRepo,
		TxRepo:           txRepo,
	}

	// Create services
	jwtSvc := jwt.NewService("test-secret", time.Hour, 24*time.Hour)
	passwordService := password.NewService()

	authSvc := service.NewAuthService(repo, jwtSvc, passwordService)

	// Create handlers
	authHandler := handler.NewAuthHandler(authSvc)
	adminHandler := handler.NewAdminHandler(authSvc)

	// Setup router
	router := gin.New()
	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		user := api.Group("/user")
		user.Use(handler.AuthMiddleware(jwtSvc))
		{
			user.GET("/profile", authHandler.GetProfile)
			user.PUT("/profile", authHandler.UpdateProfile)
			user.POST("/change-password", authHandler.ChangePassword)
			user.POST("/logout", authHandler.Logout)
		}

		admin := api.Group("/admin")
		admin.Use(handler.AuthMiddleware(jwtSvc))
		admin.Use(handler.AdminMiddleware())
		{
			admin.GET("/users", adminHandler.ListUsers)
			admin.PUT("/users/:userId/activate", adminHandler.ActivateUser)
			admin.PUT("/users/:userId/deactivate", adminHandler.DeactivateUser)
			admin.DELETE("/users/:userId", adminHandler.DeleteUser)
		}
	}

	server := httptest.NewServer(router)

	cleanup := func() {
		server.Close()
		testDB.TeardownTestDB(t)
	}

	return server, cleanup
}

func TestAuthAPI_Login(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// First, we need to seed some test data
	// Since we can't directly access the test DB from here, we'll need to modify the setup
	// For now, let's test the endpoint structure

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectToken    bool
	}{
		{
			name: "successful login",
			requestBody: map[string]interface{}{
				"email":    "admin@default.local",
				"password": "Admin123!",
				"tenant_id": "default.local",
			},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name: "invalid credentials",
			requestBody: map[string]interface{}{
				"email":    "admin@default.local",
				"password": "wrongpassword",
				"tenant_id": "default.local",
			},
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"email": "admin@default.local",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req, err := http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Make request
			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Check status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Parse response
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			if tt.expectToken {
				assert.Equal(t, true, response["success"])
				data, ok := response["data"].(map[string]interface{})
				require.True(t, ok)
				token, ok := data["token"].(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, token["access_token"])
				assert.NotEmpty(t, token["refresh_token"])
			} else {
				assert.Equal(t, false, response["success"])
				assert.Contains(t, response, "message")
			}
		})
	}
}

func TestAuthAPI_Register(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectToken    bool
	}{
		{
			name: "successful registration",
			requestBody: map[string]interface{}{
				"email":            "newuser@example.com",
				"password":         "NewPass123!",
				"confirm_password": "NewPass123!",
				"user_type":        "Student",
				"tenant_id":        "default.local",
				"first_name":       "John",
				"last_name":        "Doe",
			},
			expectedStatus: http.StatusCreated,
			expectToken:    true,
		},
		{
			name: "password mismatch",
			requestBody: map[string]interface{}{
				"email":            "newuser2@example.com",
				"password":         "NewPass123!",
				"confirm_password": "DifferentPass123!",
				"user_type":        "Student",
				"tenant_id":        "default.local",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
		{
			name: "weak password",
			requestBody: map[string]interface{}{
				"email":            "newuser3@example.com",
				"password":         "123",
				"confirm_password": "123",
				"user_type":        "Student",
				"tenant_id":        "default.local",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req, err := http.NewRequest("POST", server.URL+"/api/v1/auth/register", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Make request
			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Check status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Parse response
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			if tt.expectToken {
				assert.Equal(t, true, response["success"])
				data, ok := response["data"].(map[string]interface{})
				require.True(t, ok)
				token, ok := data["token"].(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, token["access_token"])
			} else {
				assert.Equal(t, false, response["success"])
			}
		})
	}
}

func TestAuthAPI_GetProfile(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// This test would require setting up authentication tokens
	// For now, we'll test the endpoint structure

	t.Run("unauthorized access", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/api/v1/user/profile", nil)
		require.NoError(t, err)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return 401 Unauthorized due to missing token
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAuthAPI_RefreshToken(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("missing refresh token", func(t *testing.T) {
		requestBody := map[string]interface{}{}
		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", server.URL+"/api/v1/auth/refresh", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, false, response["success"])
	})
}

func TestAuthAPI_ForgotPassword(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "forgot password request",
			requestBody: map[string]interface{}{
				"email":    "admin@default.local",
				"tenant_id": "default.local",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"tenant_id": "default.local",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", server.URL+"/api/v1/auth/forgot-password", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}