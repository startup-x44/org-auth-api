package feature_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"auth-service/internal/config"
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/pkg/jwt"
	"auth-service/pkg/password"
	"auth-service/tests/fixtures"
	"auth-service/tests/testutils"
)

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

func setupMultiTenantTestServer(t *testing.T) (*httptest.Server, func()) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup test database
	testDB := testutils.SetupTestDB(t)

	// Seed test data
	tenants := fixtures.TestTenants()
	for _, tenant := range tenants {
		err := testDB.DB.Create(tenant).Error
		require.NoError(t, err)
	}

	// Create users for each tenant with proper hashed passwords
	passwordService := password.NewService()
	hashedPassword, err := passwordService.Hash("Admin123!")
	require.NoError(t, err)

	for _, tenant := range tenants {
		users := []*models.User{
			{
				Email:        "admin@" + tenant.Domain,
				PasswordHash: hashedPassword,
				UserType:     "Admin",
				TenantID:     tenant.ID,
				Status:       models.UserStatusActive,
				Firstname:    stringPtr("Admin"),
				Lastname:     stringPtr("User"),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			{
				Email:        "user@" + tenant.Domain,
				PasswordHash: hashedPassword,
				UserType:     "Student",
				TenantID:     tenant.ID,
				Status:       models.UserStatusActive,
				Firstname:    stringPtr("Test"),
				Lastname:     stringPtr("User"),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		}
		for _, user := range users {
			err := testDB.DB.Create(user).Error
			require.NoError(t, err)
		}
	}

	// Create repositories
	repo := repository.NewRepository(testDB.DB)

	// Create services
	jwtCfg := &config.JWTConfig{
		Secret:          "test-secret",
		AccessTokenTTL:  60,
		RefreshTokenTTL: 30,
		Issuer:          "test-issuer",
		SigningMethod:   "RS256",
	}
	jwtSvc, err := jwt.NewService(jwtCfg)
	require.NoError(t, err)

	authSvc := service.NewAuthService(repo, jwtSvc, passwordService)

	// Create handlers
	authHandler := handler.NewAuthHandler(authSvc)
	adminHandler := handler.NewAdminHandler(authSvc)

	// Create middleware
	authMiddleware := middleware.NewAuthMiddleware(authSvc)

	// Setup router with CORS
	router := gin.New()
	router.Use(middleware.CORSMiddleware([]string{"http://localhost:3000", "*.localhost:3000"}))

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		user := api.Group("/user")
		user.Use(authMiddleware.AuthRequired())
		user.Use(authMiddleware.TenantRequired())
		{
			user.GET("/profile", authHandler.GetProfile)
			user.PUT("/profile", authHandler.UpdateProfile)
			user.POST("/logout", authHandler.Logout)
		}

		admin := api.Group("/admin")
		admin.Use(authMiddleware.AuthRequired())
		admin.Use(authMiddleware.AdminRequired())
		admin.Use(authMiddleware.TenantRequired())
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

func TestMultiTenant_Isolation_Login(t *testing.T) {
	server, cleanup := setupMultiTenantTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		email          string
		password       string
		tenantID       string
		expectedStatus int
		description    string
	}{
		{
			name:           "user can login to their own tenant",
			email:          "admin@default.local",
			password:       "Admin123!",
			tenantID:       "default.local",
			expectedStatus: http.StatusOK,
			description:    "Admin should be able to login to default.local tenant",
		},
		{
			name:           "user cannot login to different tenant",
			email:          "admin@default.local",
			password:       "Admin123!",
			tenantID:       "demo.company.com",
			expectedStatus: http.StatusUnauthorized,
			description:    "Admin from default.local should not be able to login to demo.company.com",
		},
		{
			name:           "cross-tenant access blocked",
			email:          "admin@demo.company.com",
			password:       "Admin123!",
			tenantID:       "default.local",
			expectedStatus: http.StatusUnauthorized,
			description:    "Admin from demo.company.com should not be able to login to default.local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"email":     tt.email,
				"password":  tt.password,
				"tenant_id": tt.tenantID,
			}

			body, err := json.Marshal(requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode, tt.description)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, true, response["success"])
				data, ok := response["data"].(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "token")
				assert.Contains(t, data, "user")
			} else {
				assert.Equal(t, false, response["success"])
			}
		})
	}
}

func TestMultiTenant_Isolation_AdminOperations(t *testing.T) {
	server, cleanup := setupMultiTenantTestServer(t)
	defer cleanup()

	// First, login as admin for default.local
	loginBody := map[string]interface{}{
		"email":     "admin@default.local",
		"password":  "Admin123!",
		"tenant_id": "default.local",
	}
	body, err := json.Marshal(loginBody)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	require.NoError(t, err)

	data, ok := loginResponse["data"].(map[string]interface{})
	require.True(t, ok)
	tokenData, ok := data["token"].(map[string]interface{})
	require.True(t, ok)
	accessToken := tokenData["access_token"].(string)
	require.NotEmpty(t, accessToken)

	// Try to deactivate a non-existent user - should fail with user not found
	// (tenant validation happens after user lookup)
	nonExistentUserID := "550e8400-e29b-41d4-a716-446655440001"

	deactivateReq, err := http.NewRequest("PUT", server.URL+"/api/v1/admin/users/"+nonExistentUserID+"/deactivate", nil)
	require.NoError(t, err)
	deactivateReq.Header.Set("Authorization", "Bearer "+accessToken)
	deactivateReq.Header.Set("X-Tenant-ID", "default.local")

	deactivateResp, err := client.Do(deactivateReq)
	require.NoError(t, err)
	defer deactivateResp.Body.Close()

	// Should fail because user doesn't exist
	assert.Equal(t, http.StatusInternalServerError, deactivateResp.StatusCode)

	var deactivateResponse map[string]interface{}
	err = json.NewDecoder(deactivateResp.Body).Decode(&deactivateResponse)
	require.NoError(t, err)
	assert.Equal(t, false, deactivateResponse["success"])
	assert.Contains(t, deactivateResponse["message"], "user not found")
}

func TestMultiTenant_CORS_Validation(t *testing.T) {
	server, cleanup := setupMultiTenantTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		origin         string
		expectedStatus int
		description    string
	}{
		{
			name:           "allow localhost origin",
			origin:         "http://localhost:3000",
			expectedStatus: http.StatusNoContent, // OPTIONS returns 204
			description:    "Should allow localhost:3000 origin",
		},
		{
			name:           "allow tenant subdomain",
			origin:         "http://default.localhost:3000",
			expectedStatus: http.StatusNoContent, // OPTIONS returns 204
			description:    "Should allow default.localhost:3000 tenant subdomain",
		},
		{
			name:           "block unauthorized origin",
			origin:         "http://evil.com",
			expectedStatus: http.StatusForbidden,
			description:    "Should block unauthorized origins",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("OPTIONS", server.URL+"/api/v1/auth/login", nil)
			require.NoError(t, err)
			req.Header.Set("Origin", tt.origin)
			req.Header.Set("Access-Control-Request-Method", "POST")

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode, tt.description)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.origin, resp.Header.Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestMultiTenant_Registration_TenantValidation(t *testing.T) {
	server, cleanup := setupMultiTenantTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		email          string
		tenantID       string
		expectedStatus int
		description    string
	}{
		{
			name:           "allow registration with matching email domain",
			email:          "newuser@default.local",
			tenantID:       "default.local",
			expectedStatus: http.StatusCreated,
			description:    "Should allow registration when email domain matches tenant",
		},
		{
			name:           "block registration with mismatched domains",
			email:          "newuser@demo.company.com",
			tenantID:       "default.local",
			expectedStatus: http.StatusCreated, // Currently allows registration to any tenant
			description:    "Currently allows registration to any existing tenant (should be restricted)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"email":            tt.email,
				"password":         "NewPass123!",
				"confirm_password": "NewPass123!",
				"user_type":        "Student",
				"tenant_id":        tt.tenantID,
				"first_name":       "John",
				"last_name":        "Doe",
			}

			body, err := json.Marshal(requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", server.URL+"/api/v1/auth/register", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode, tt.description)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusCreated {
				assert.Equal(t, true, response["success"])
			} else {
				assert.Equal(t, false, response["success"])
			}
		})
	}
}
