package repository

import (
	"context"
	"time"

	"auth-service/internal/models"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit int, cursor string) ([]*models.User, error)
	Count(ctx context.Context) (int64, error)
	UpdateLastLogin(ctx context.Context, id string) error
	UpdatePassword(ctx context.Context, id, hashedPassword string) error
	Activate(ctx context.Context, id string) error
	Deactivate(ctx context.Context, id string) error
}

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	Create(ctx context.Context, tenant *models.Tenant) error
	GetByID(ctx context.Context, id string) (*models.Tenant, error)
	GetByDomain(ctx context.Context, domain string) (*models.Tenant, error)
	Update(ctx context.Context, tenant *models.Tenant) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.Tenant, error)
	Count(ctx context.Context) (int64, error)
}

// UserSessionRepository defines the interface for user session data operations
type UserSessionRepository interface {
	Create(ctx context.Context, session *models.UserSession) error
	GetByID(ctx context.Context, id string) (*models.UserSession, error)
	GetByToken(ctx context.Context, token string) (*models.UserSession, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.UserSession, error)
	GetActiveByUserID(ctx context.Context, userID string) ([]*models.UserSession, error)
	GetActiveCountByUserID(ctx context.Context, userID string) (int64, error)
	UpdateActivity(ctx context.Context, id string) error
	Revoke(ctx context.Context, id string, reason string) error
	RevokeByUserID(ctx context.Context, userID string, reason string) error
	RevokeExpired(ctx context.Context) error
	RevokeInactive(ctx context.Context, maxInactive time.Duration) error
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
	CleanupExpired(ctx context.Context, maxAge time.Duration) error
	GetSessionsByIPAddress(ctx context.Context, ipAddress string, since time.Time) ([]*models.UserSession, error)
	GetSessionsByDeviceFingerprint(ctx context.Context, fingerprint string) ([]*models.UserSession, error)
}

// RefreshTokenRepository defines the interface for refresh token data operations
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*models.RefreshToken, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.RefreshToken, error)
	Update(ctx context.Context, token *models.RefreshToken) error
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
	CleanupExpired(ctx context.Context, maxAge time.Duration) error
}

// PasswordResetRepository defines the interface for password reset data operations
type PasswordResetRepository interface {
	Create(ctx context.Context, reset *models.PasswordReset) error
	GetByToken(ctx context.Context, token string) (*models.PasswordReset, error)
	GetByEmail(ctx context.Context, email string) (*models.PasswordReset, error)
	Update(ctx context.Context, reset *models.PasswordReset) error
	Delete(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
	CleanupExpired(ctx context.Context, maxAge time.Duration) error
}

// FailedLoginAttemptRepository defines the interface for failed login attempt data operations
type FailedLoginAttemptRepository interface {
	Create(ctx context.Context, attempt *models.FailedLoginAttempt) error
	GetByEmailAndIP(ctx context.Context, email, ipAddress string, since time.Time) ([]*models.FailedLoginAttempt, error)
	CountByEmailAndIP(ctx context.Context, email, ipAddress string, since time.Time) (int64, error)
	DeleteExpired(ctx context.Context, maxAge time.Duration) error
	CleanupExpired(ctx context.Context, maxAge time.Duration) error
}

// OrganizationRepository defines the interface for organization data operations
type OrganizationRepository interface {
	Create(ctx context.Context, org *models.Organization) error
	GetByID(ctx context.Context, id string) (*models.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*models.Organization, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Organization, error)
	Update(ctx context.Context, org *models.Organization) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.Organization, error)
	Count(ctx context.Context) (int64, error)
}

// OrganizationMembershipRepository defines the interface for organization membership data operations
type OrganizationMembershipRepository interface {
	Create(ctx context.Context, membership *models.OrganizationMembership) error
	GetByID(ctx context.Context, id string) (*models.OrganizationMembership, error)
	GetByOrganizationAndUser(ctx context.Context, orgID, userID string) (*models.OrganizationMembership, error)
	GetByOrganizationAndEmail(ctx context.Context, orgID, email string) (*models.OrganizationMembership, error)
	GetByOrganization(ctx context.Context, orgID string) ([]*models.OrganizationMembership, error)
	GetByUser(ctx context.Context, userID string) ([]*models.OrganizationMembership, error)
	Update(ctx context.Context, membership *models.OrganizationMembership) error
	Delete(ctx context.Context, orgID, userID string) error
	CountByOrganization(ctx context.Context, orgID string) (int64, error)
}

// OrganizationInvitationRepository defines the interface for organization invitation data operations
type OrganizationInvitationRepository interface {
	Create(ctx context.Context, invitation *models.OrganizationInvitation) error
	GetByID(ctx context.Context, id string) (*models.OrganizationInvitation, error)
	GetByToken(ctx context.Context, tokenHash string) (*models.OrganizationInvitation, error)
	GetByOrganizationAndEmail(ctx context.Context, orgID, email string) (*models.OrganizationInvitation, error)
	GetPendingByOrganization(ctx context.Context, orgID string) ([]*models.OrganizationInvitation, error)
	Update(ctx context.Context, invitation *models.OrganizationInvitation) error
	Delete(ctx context.Context, id string) error
	CleanupExpired(ctx context.Context, maxAge time.Duration) error
}

// Repository defines the interface for all repository operations
type Repository interface {
	User() UserRepository
	Tenant() TenantRepository
	Organization() OrganizationRepository
	OrganizationMembership() OrganizationMembershipRepository
	OrganizationInvitation() OrganizationInvitationRepository
	UserSession() UserSessionRepository
	RefreshToken() RefreshTokenRepository
	PasswordReset() PasswordResetRepository
	FailedLoginAttempt() FailedLoginAttemptRepository
	BeginTransaction(ctx context.Context) (Transaction, error)
}

// Transaction defines the interface for database transactions
type Transaction interface {
	Commit() error
	Rollback() error
	User() UserRepository
	Tenant() TenantRepository
	Organization() OrganizationRepository
	OrganizationMembership() OrganizationMembershipRepository
	OrganizationInvitation() OrganizationInvitationRepository
	UserSession() UserSessionRepository
	RefreshToken() RefreshTokenRepository
	PasswordReset() PasswordResetRepository
	FailedLoginAttempt() FailedLoginAttemptRepository
}
