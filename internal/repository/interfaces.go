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
	GetByEmail(ctx context.Context, email, tenantID string) (*models.User, error)
	GetByEmailAndType(ctx context.Context, email, userType, tenantID string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, tenantID string, limit, offset int) ([]*models.User, error)
	Count(ctx context.Context, tenantID string) (int64, error)
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
	GetByUserID(ctx context.Context, userID string) ([]*models.UserSession, error)
	GetActiveByUserID(ctx context.Context, userID string) ([]*models.UserSession, error)
	Update(ctx context.Context, session *models.UserSession) error
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
	CleanupExpired(ctx context.Context, maxAge time.Duration) error
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
	GetByEmail(ctx context.Context, email, tenantID string) (*models.PasswordReset, error)
	Update(ctx context.Context, reset *models.PasswordReset) error
	Delete(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
	CleanupExpired(ctx context.Context, maxAge time.Duration) error
}

// Repository defines the interface for all repository operations
type Repository interface {
	User() UserRepository
	Tenant() TenantRepository
	UserSession() UserSessionRepository
	RefreshToken() RefreshTokenRepository
	PasswordReset() PasswordResetRepository
	BeginTransaction(ctx context.Context) (Transaction, error)
}

// Transaction defines the interface for database transactions
type Transaction interface {
	Commit() error
	Rollback() error
	User() UserRepository
	Tenant() TenantRepository
	UserSession() UserSessionRepository
	RefreshToken() RefreshTokenRepository
	PasswordReset() PasswordResetRepository
}