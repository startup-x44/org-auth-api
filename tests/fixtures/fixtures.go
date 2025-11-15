package fixtures

import (
	"time"

	"github.com/google/uuid"

	"auth-service/internal/models"
)

// TestTenant returns a test tenant fixture
func TestTenant() *models.Tenant {
	return &models.Tenant{
		Name:   "Test Organization",
		Domain: "test.local",
	}
}

// TestUser returns a test user fixture
func TestUser(tenantID uuid.UUID) *models.User {
	return &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		IsSuperadmin: false,
		GlobalRole:   "user",
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// TestStudent returns a test student user fixture
func TestStudent(tenantID uuid.UUID) *models.User {
	return &models.User{
		Email:        "student@example.com",
		PasswordHash: "hashedpassword",
		IsSuperadmin: false,
		GlobalRole:   "user",
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// TestRTO returns a test RTO user fixture
func TestRTO(tenantID uuid.UUID) *models.User {
	return &models.User{
		Email:        "rto@example.com",
		PasswordHash: "hashedpassword",
		IsSuperadmin: false,
		GlobalRole:   "user",
		Status:       models.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// TestUserSession returns a test user session fixture
func TestUserSession(userID, tenantID uuid.UUID) *models.UserSession {
	return &models.UserSession{
		UserID:    userID,
		TokenHash: "sessiontoken",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

// TestRefreshToken returns a test refresh token fixture
func TestRefreshToken(userID, tenantID uuid.UUID) *models.RefreshToken {
	return &models.RefreshToken{
		UserID:    userID,
		TokenHash: "refreshtoken",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
}

// TestPasswordReset returns a test password reset fixture
func TestPasswordReset(userID, tenantID uuid.UUID) *models.PasswordReset {
	return &models.PasswordReset{
		UserID:    userID,
		TokenHash: "resettoken",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
}

// TestTenants returns multiple test tenants
func TestTenants() []*models.Tenant {
	return []*models.Tenant{
		{
			Name:   "Default Organization",
			Domain: "default.local",
		},
		{
			Name:   "Demo Company",
			Domain: "demo.company.com",
		},
		{
			Name:   "Test Organization",
			Domain: "test.org",
		},
	}
}

// TestUsers returns multiple test users for a tenant
func TestUsers(tenantID uuid.UUID) []*models.User {
	return []*models.User{
		{
			Email:        "admin@test.local",
			PasswordHash: "hashedpassword1",
			IsSuperadmin: false,
			GlobalRole:   "user",
			Status:       "active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Email:        "student@test.local",
			PasswordHash: "hashedpassword2",
			IsSuperadmin: false,
			GlobalRole:   "user",
			Status:       "active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Email:        "rto@test.local",
			PasswordHash: "hashedpassword3",
			IsSuperadmin: false,
			GlobalRole:   "user",
			Status:       "active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}
}
