package service

import (
	"context"
	"time"

	"auth-service/internal/repository"
	"auth-service/pkg/jwt"
	"auth-service/pkg/password"
)

// AuthService defines the interface for authentication business logic
type AuthService interface {
	UserService() UserService
	TenantService() TenantService
	OrganizationService() OrganizationService
	SessionService() SessionService
	BackgroundJobService() BackgroundJobService
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	HealthCheck(ctx context.Context) (*HealthCheckResponse, error)
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID       string  `json:"user_id"`
	Email        string  `json:"email"`
	IsSuperadmin bool    `json:"is_superadmin"`
	CurrentOrgID *string `json:"current_org_id,omitempty"`
}

// HealthCheckResponse represents health check response
type HealthCheckResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Redis     string `json:"redis"`
	Timestamp string `json:"timestamp"`
}

// authService implements AuthService interface
type authService struct {
	userService         UserService
	tenantService       TenantService
	organizationService OrganizationService
	sessionSvc          SessionService
	jobSvc              BackgroundJobService
	jwtService          *jwt.Service
	repo                repository.Repository
}

// NewAuthService creates a new auth service
func NewAuthService(repo repository.Repository, jwtService *jwt.Service, passwordService *password.Service) AuthService {
	// Create session service
	sessionConfig := &SessionConfig{
		MaxSessionsPerUser:          5,
		SessionTimeout:              24 * time.Hour,
		MaxInactiveTime:             7 * 24 * time.Hour,
		EnableGeoTracking:           true,
		EnableDeviceTracking:        true,
		SuspiciousActivityThreshold: 3,
	}
	sessionSvc := NewSessionService(repo, sessionConfig)

	// Create background job service
	jobConfig := &BackgroundJobConfig{
		SessionCleanupInterval:       1 * time.Hour,
		TokenCleanupInterval:         6 * time.Hour,
		FailedAttemptCleanupInterval: 24 * time.Hour,
		MaxInactiveSessionTime:       30 * 24 * time.Hour,
		MaxFailedAttemptAge:          7 * 24 * time.Hour,
	}
	jobSvc := NewBackgroundJobService(repo, sessionSvc, jobConfig)

	userSvc := NewUserService(repo, jwtService, passwordService)
	userSvc.SetSessionService(sessionSvc)

	return &authService{
		userService:         userSvc,
		tenantService:       NewTenantService(repo),
		organizationService: NewOrganizationService(repo),
		sessionSvc:          sessionSvc,
		jobSvc:              jobSvc,
		jwtService:          jwtService,
		repo:                repo,
	}
}

// UserService returns the user service
func (s *authService) UserService() UserService {
	return s.userService
}

// TenantService returns the tenant service
func (s *authService) TenantService() TenantService {
	return s.tenantService
}

// OrganizationService returns the organization service
func (s *authService) OrganizationService() OrganizationService {
	return s.organizationService
}

// SessionService returns the session service
func (s *authService) SessionService() SessionService {
	return s.sessionSvc
}

// BackgroundJobService returns the background job service
func (s *authService) BackgroundJobService() BackgroundJobService {
	return s.jobSvc
}

// ValidateToken validates JWT token and returns claims
func (s *authService) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return &TokenClaims{
		UserID:       claims.UserID.String(),
		Email:        claims.Email,
		IsSuperadmin: claims.IsSuperadmin,
	}, nil
}

// HealthCheck performs a health check on all dependencies
func (s *authService) HealthCheck(ctx context.Context) (*HealthCheckResponse, error) {
	response := &HealthCheckResponse{
		Status: "healthy",
	}

	// Check database connection
	if err := s.checkDatabase(ctx); err != nil {
		response.Status = "unhealthy"
		response.Database = "error"
	} else {
		response.Database = "ok"
	}

	// Check Redis connection (placeholder - would need Redis client)
	response.Redis = "ok" // TODO: Implement Redis health check

	// Set timestamp
	response.Timestamp = "2024-01-01T00:00:00Z" // TODO: Use actual timestamp

	return response, nil
}

// checkDatabase performs a simple database health check
func (s *authService) checkDatabase(ctx context.Context) error {
	// Try to count tenants as a simple health check
	_, err := s.repo.Tenant().Count(ctx)
	return err
}
