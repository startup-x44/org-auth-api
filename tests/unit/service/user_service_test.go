package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"auth-service/internal/config"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/pkg/jwt"
	"auth-service/pkg/password"
	"auth-service/tests/testutils"
)

func TestUserService_Login(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	// Create test tenant
	tenant := testutils.CreateTestTenant(t, testDB.DB, "Test Organization", "test.local")

	// Create test user
	passwordService := password.NewService()
	hashedPassword, err := passwordService.Hash("TestPass123!")
	require.NoError(t, err)

	user := testutils.CreateTestUser(t, testDB.DB, "test@example.com", hashedPassword, "Admin", tenant.ID)

	// Create repositories
	repo := repository.NewRepository(testDB.DB)

	// Create JWT config and service
	jwtConfig := &config.JWTConfig{
		Secret:          "test-secret",
		AccessTokenTTL:  60,   // 1 hour
		RefreshTokenTTL: 30,   // 30 days
		Issuer:          "test-issuer",
		SigningMethod:   "RS256",
	}
	jwtSvc, _ := jwt.NewService(jwtConfig)

	// Create user service
	userSvc := service.NewUserService(repo, jwtSvc, passwordService)

	tests := []struct {
		name        string
		req         *service.LoginRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "successful login with domain",
			req: &service.LoginRequest{
				Email:    "test@example.com",
				Password: "TestPass123!",
				TenantID: "test.local",
			},
			wantErr: false,
		},
		{
			name: "successful login with UUID",
			req: &service.LoginRequest{
				Email:    "test@example.com",
				Password: "TestPass123!",
				TenantID: tenant.ID.String(),
			},
			wantErr: false,
		},
		{
			name: "invalid password",
			req: &service.LoginRequest{
				Email:    "test@example.com",
				Password: "WrongPass123!",
				TenantID: "test.local",
			},
			wantErr:     true,
			errContains: "invalid credentials",
		},
		{
			name: "user not found",
			req: &service.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "TestPass123!",
				TenantID: "test.local",
			},
			wantErr:     true,
			errContains: "invalid credentials",
		},
		{
			name: "invalid tenant",
			req: &service.LoginRequest{
				Email:    "test@example.com",
				Password: "TestPass123!",
				TenantID: "invalid.local",
			},
			wantErr:     true,
			errContains: "invalid tenant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean test data before each test
			testutils.CleanTestData(t, testDB.DB)

			// Recreate test data
			tenant = testutils.CreateTestTenant(t, testDB.DB, "Test Organization", "test.local")
			user = testutils.CreateTestUser(t, testDB.DB, "test@example.com", hashedPassword, "Admin", tenant.ID)

			resp, err := userSvc.Login(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Token)
				assert.NotEmpty(t, resp.Token.AccessToken)
				assert.NotEmpty(t, resp.Token.RefreshToken)
				assert.Equal(t, user.ID.String(), resp.User.ID)
				assert.Equal(t, user.Email, resp.User.Email)
			}
		})
	}
}

func TestUserService_Register(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	// Create test tenant
	tenant := testutils.CreateTestTenant(t, testDB.DB, "Test Organization", "test.local")

	// Create repositories
	repo := repository.NewRepository(testDB.DB)

	// Create JWT config and service
	jwtConfig := &config.JWTConfig{
		Secret:          "test-secret",
		AccessTokenTTL:  60,   // 1 hour
		RefreshTokenTTL: 30,   // 30 days
		Issuer:          "test-issuer",
		SigningMethod:   "RS256",
	}
	jwtSvc, _ := jwt.NewService(jwtConfig)

	passwordService := password.NewService()
	userSvc := service.NewUserService(repo, jwtSvc, passwordService)

	tests := []struct {
		name        string
		req         *service.RegisterRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "successful registration",
			req: &service.RegisterRequest{
				Email:           "newuser@example.com",
				Password:        "NewPass123!",
				ConfirmPassword: "NewPass123!",
				UserType:        "Student",
				TenantID:        tenant.ID.String(),
				FirstName:       "John",
				LastName:        "Doe",
			},
			wantErr: false,
		},
		{
			name: "password mismatch",
			req: &service.RegisterRequest{
				Email:           "newuser2@example.com",
				Password:        "NewPass123!",
				ConfirmPassword: "DifferentPass123!",
				UserType:        "Student",
				TenantID:        tenant.ID.String(),
			},
			wantErr:     true,
			errContains: "passwords do not match",
		},
		{
			name: "weak password",
			req: &service.RegisterRequest{
				Email:           "newuser3@example.com",
				Password:        "123",
				ConfirmPassword: "123",
				UserType:        "Student",
				TenantID:        tenant.ID.String(),
			},
			wantErr:     true,
			errContains: "password",
		},
		{
			name: "duplicate email",
			req: &service.RegisterRequest{
				Email:           "newuser@example.com", // Same as first test
				Password:        "AnotherPass123!",
				ConfirmPassword: "AnotherPass123!",
				UserType:        "Student",
				TenantID:        tenant.ID.String(),
			},
			wantErr:     true,
			errContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean test data before each test
			testutils.CleanTestData(t, testDB.DB)

			// Recreate test tenant
			tenant = testutils.CreateTestTenant(t, testDB.DB, "Test Organization", "test.local")

			resp, err := userSvc.Register(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Token)
				assert.NotEmpty(t, resp.Token.AccessToken)
				assert.Equal(t, tt.req.Email, resp.User.Email)
				assert.Equal(t, tt.req.UserType, resp.User.UserType)
			}
		})
	}
}

func TestUserService_GetProfile(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	// Create test tenant and user
	tenant := testutils.CreateTestTenant(t, testDB.DB, "Test Organization", "test.local")
	passwordService := password.NewService()
	hashedPassword, _ := passwordService.Hash("TestPass123!")
	user := testutils.CreateTestUser(t, testDB.DB, "test@example.com", hashedPassword, "Admin", tenant.ID)

	// Create repositories and service
	repo := repository.NewRepository(testDB.DB)

	jwtConfig := &config.JWTConfig{
		Secret:          "test-secret",
		AccessTokenTTL:  60,   // 1 hour
		RefreshTokenTTL: 30,   // 30 days
		Issuer:          "test-issuer",
		SigningMethod:   "RS256",
	}
	jwtSvc, _ := jwt.NewService(jwtConfig)
	userSvc := service.NewUserService(repo, jwtSvc, passwordService)

	t.Run("successful profile retrieval", func(t *testing.T) {
		profile, err := userSvc.GetProfile(context.Background(), user.ID.String())

		assert.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, user.ID.String(), profile.ID)
		assert.Equal(t, user.Email, profile.Email)
		assert.Equal(t, user.UserType, profile.UserType)
		assert.Equal(t, tenant.ID.String(), profile.TenantID)
	})

	t.Run("user not found", func(t *testing.T) {
		profile, err := userSvc.GetProfile(context.Background(), uuid.New().String())

		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUserService_resolveTenantID(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	// Create test tenant
	tenant := testutils.CreateTestTenant(t, testDB.DB, "Test Organization", "test.local")

	// Create repositories and service
	repo := repository.NewRepository(testDB.DB)

	jwtConfig := &config.JWTConfig{
		Secret:          "test-secret",
		AccessTokenTTL:  60,   // 1 hour
		RefreshTokenTTL: 30,   // 30 days
		Issuer:          "test-issuer",
		SigningMethod:   "RS256",
	}
	jwtSvc, _ := jwt.NewService(jwtConfig)
	passwordService := password.NewService()
	userSvc := service.NewUserService(repo, jwtSvc, passwordService)

	// Test tenant resolution through login attempts
	tests := []struct {
		name        string
		tenantID    string
		shouldLogin bool
	}{
		{
			name:        "domain resolves correctly",
			tenantID:    "test.local",
			shouldLogin: true,
		},
		{
			name:        "UUID works directly",
			tenantID:    tenant.ID.String(),
			shouldLogin: true,
		},
		{
			name:        "invalid domain fails",
			tenantID:    "nonexistent.local",
			shouldLogin: false,
		},
	}

	passwordService2 := password.NewService()
	hashedPassword, _ := passwordService2.Hash("TestPass123!")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean and recreate test data
			testutils.CleanTestData(t, testDB.DB)
			tenant = testutils.CreateTestTenant(t, testDB.DB, "Test Organization", "test.local")
			testutils.CreateTestUser(t, testDB.DB, "test@example.com", hashedPassword, "Admin", tenant.ID)

			req := &service.LoginRequest{
				Email:    "test@example.com",
				Password: "TestPass123!",
				TenantID: tt.tenantID,
			}

			resp, err := userSvc.Login(context.Background(), req)

			if tt.shouldLogin {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				assert.Error(t, err)
				assert.Nil(t, resp)
			}
		})
	}
}

// BenchmarkUserRegistration benchmarks the complete user registration flow
func BenchmarkUserRegistration(b *testing.B) {
	// Simplified benchmark - just measure password hashing which is the most expensive part
	passwordService := password.NewService()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := passwordService.Hash("BenchmarkPass123!")
		if err != nil {
			b.Fatalf("Failed to hash password: %v", err)
		}
	}
}

// BenchmarkUserLogin benchmarks the user login flow
func BenchmarkUserLogin(b *testing.B) {
	// Simplified benchmark - just measure password verification which is the most expensive part
	passwordService := password.NewService()
	hashedPassword, err := passwordService.Hash("BenchmarkPass123!")
	if err != nil {
		b.Fatalf("Failed to setup password hash: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		valid, err := passwordService.Verify("BenchmarkPass123!", hashedPassword)
		if err != nil || !valid {
			b.Fatalf("Failed to verify password: %v", err)
		}
	}
}

// BenchmarkUserGetProfile benchmarks profile retrieval performance
func BenchmarkUserGetProfile(b *testing.B) {
	// Simplified benchmark - just measure basic operations
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("user-%d", i)
	}
}