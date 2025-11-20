package service

import (
	"context"
	"time"

	"auth-service/internal/repository"
	"auth-service/pkg/email"
	"auth-service/pkg/jwt"
	"auth-service/pkg/password"

	"github.com/go-redis/redis/v8"
)

// AuthService defines the interface for authentication business logic
type AuthService interface {
	UserService() UserService
	OrganizationService() OrganizationService
	SessionService() SessionService
	BackgroundJobService() BackgroundJobService
	RoleService() RoleService
	RevocationService() RevocationService
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	HealthCheck(ctx context.Context) (*HealthCheckResponse, error)
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID           string   `json:"user_id"`
	Email            string   `json:"email"`
	OrganizationID   string   `json:"organization_id"`
	SessionID        string   `json:"session_id"`
	GlobalRole       string   `json:"global_role"`
	OrganizationRole string   `json:"organization_role"`
	Permissions      []string `json:"permissions"` // Cached permission names
	IsSuperadmin     bool     `json:"is_superadmin"`
	CurrentOrgID     *string  `json:"current_org_id,omitempty"`
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
	organizationService OrganizationService
	sessionSvc          SessionService
	jobSvc              BackgroundJobService
	roleSvc             RoleService
	revocationSvc       RevocationService
	jwtService          *jwt.Service
	emailService        email.Service
	repo                repository.Repository
}

// NewAuthService creates a new auth service
func NewAuthService(repo repository.Repository, jwtService *jwt.Service, passwordService *password.Service, emailService email.Service, redisClient *redis.Client) AuthService {
	// Session service
	sessionConfig := &SessionConfig{
		MaxSessionsPerUser:          5,
		SessionTimeout:              24 * time.Hour,
		MaxInactiveTime:             7 * 24 * time.Hour,
		EnableGeoTracking:           true,
		EnableDeviceTracking:        true,
		SuspiciousActivityThreshold: 3,
	}

	sessionSvc := NewSessionService(repo, sessionConfig)

	// Background job service
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

	// Initialize role service with a nil audit logger for now
	roleSvc := NewRoleService(repo, nil)

	// Initialize revocation service
	revocationSvc := NewRevocationService(repo, jwtService, redisClient)

	return &authService{
		userService:         userSvc,
		organizationService: NewOrganizationService(repo, emailService),
		sessionSvc:          sessionSvc,
		jobSvc:              jobSvc,
		roleSvc:             roleSvc,
		revocationSvc:       revocationSvc,
		jwtService:          jwtService,
		emailService:        emailService,
		repo:                repo,
	}
}

func (s *authService) UserService() UserService                   { return s.userService }
func (s *authService) OrganizationService() OrganizationService   { return s.organizationService }
func (s *authService) SessionService() SessionService             { return s.sessionSvc }
func (s *authService) BackgroundJobService() BackgroundJobService { return s.jobSvc }
func (s *authService) RoleService() RoleService                   { return s.roleSvc }
func (s *authService) RevocationService() RevocationService       { return s.revocationSvc }

// ValidateToken validates JWT token and returns safe claims
func (s *authService) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	claims, err := s.jwtService.ParseAccessToken(token)
	if err != nil {
		return nil, err
	}

	// Optionally retrieve user "current organization" preference later
	var currentOrgID *string
	orgID := claims.OrganizationID.String()
	currentOrgID = &orgID

	return &TokenClaims{
		UserID:           claims.UserID.String(),
		Email:            claims.Email,
		OrganizationID:   claims.OrganizationID.String(),
		SessionID:        claims.SessionID.String(),
		GlobalRole:       claims.GlobalRole,
		OrganizationRole: claims.OrganizationRole,
		Permissions:      claims.Permissions,
		IsSuperadmin:     claims.IsSuperadmin,
		CurrentOrgID:     currentOrgID,
	}, nil
}

// HealthCheck performs a health check on all dependencies
func (s *authService) HealthCheck(ctx context.Context) (*HealthCheckResponse, error) {
	resp := &HealthCheckResponse{
		Status: "healthy",
	}

	// Database health check
	if err := s.checkDatabase(ctx); err != nil {
		resp.Status = "unhealthy"
		resp.Database = "error"
	} else {
		resp.Database = "ok"
	}

	// Redis health check (placeholder)
	resp.Redis = "unknown" // update when Redis client is added

	// Timestamp
	resp.Timestamp = time.Now().UTC().Format(time.RFC3339)

	return resp, nil
}

// checkDatabase performs a simple database health check
func (s *authService) checkDatabase(ctx context.Context) error {
	_, err := s.repo.Organization().Count(ctx)
	return err
}
