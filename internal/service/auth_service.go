package service

import (
	"context"

	"auth-service/internal/repository"
	"auth-service/pkg/jwt"
	"auth-service/pkg/password"
)

// AuthService defines the interface for authentication business logic
type AuthService interface {
	UserService() UserService
	TenantService() TenantService
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	HealthCheck(ctx context.Context) (*HealthCheckResponse, error)
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID   string `json:"user_id"`
	UserType string `json:"user_type"`
	TenantID string `json:"tenant_id"`
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
	userService   UserService
	tenantService TenantService
	jwtService    jwt.Service
	repo          repository.Repository
}

// NewAuthService creates a new auth service
func NewAuthService(repo repository.Repository, jwtService jwt.Service, passwordSvc password.Service) AuthService {
	return &authService{
		userService:   NewUserService(repo, jwtService, passwordSvc),
		tenantService: NewTenantService(repo),
		jwtService:    jwtService,
		repo:          repo,
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

// ValidateToken validates JWT token and returns claims
func (s *authService) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return &TokenClaims{
		UserID:   claims.UserID,
		UserType: claims.UserType,
		TenantID: claims.TenantID,
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