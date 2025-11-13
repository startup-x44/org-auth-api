package repository

import (
	"context"

	"gorm.io/gorm"

	"auth-service/internal/models"
)

// repository implements Repository interface
type repository struct {
	db             *gorm.DB
	userRepo       UserRepository
	tenantRepo     TenantRepository
	sessionRepo    UserSessionRepository
	refreshRepo    RefreshTokenRepository
	passwordRepo   PasswordResetRepository
}

// NewRepository creates a new repository instance
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db:             db,
		userRepo:       NewUserRepository(db),
		tenantRepo:     NewTenantRepository(db),
		sessionRepo:    NewUserSessionRepository(db),
		refreshRepo:    NewRefreshTokenRepository(db),
		passwordRepo:   NewPasswordResetRepository(db),
	}
}

// User returns the user repository
func (r *repository) User() UserRepository {
	return r.userRepo
}

// Tenant returns the tenant repository
func (r *repository) Tenant() TenantRepository {
	return r.tenantRepo
}

// UserSession returns the user session repository
func (r *repository) UserSession() UserSessionRepository {
	return r.sessionRepo
}

// RefreshToken returns the refresh token repository
func (r *repository) RefreshToken() RefreshTokenRepository {
	return r.refreshRepo
}

// PasswordReset returns the password reset repository
func (r *repository) PasswordReset() PasswordResetRepository {
	return r.passwordRepo
}

// BeginTransaction starts a new database transaction
func (r *repository) BeginTransaction(ctx context.Context) (Transaction, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &transaction{
		tx:             tx,
		userRepo:       NewUserRepository(tx),
		tenantRepo:     NewTenantRepository(tx),
		sessionRepo:    NewUserSessionRepository(tx),
		refreshRepo:    NewRefreshTokenRepository(tx),
		passwordRepo:   NewPasswordResetRepository(tx),
	}, nil
}

// transaction implements Transaction interface
type transaction struct {
	tx             *gorm.DB
	userRepo       UserRepository
	tenantRepo     TenantRepository
	sessionRepo    UserSessionRepository
	refreshRepo    RefreshTokenRepository
	passwordRepo   PasswordResetRepository
}

// Commit commits the transaction
func (t *transaction) Commit() error {
	return t.tx.Commit().Error
}

// Rollback rolls back the transaction
func (t *transaction) Rollback() error {
	return t.tx.Rollback().Error
}

// User returns the user repository for transaction
func (t *transaction) User() UserRepository {
	return t.userRepo
}

// Tenant returns the tenant repository for transaction
func (t *transaction) Tenant() TenantRepository {
	return t.tenantRepo
}

// UserSession returns the user session repository for transaction
func (t *transaction) UserSession() UserSessionRepository {
	return t.sessionRepo
}

// RefreshToken returns the refresh token repository for transaction
func (t *transaction) RefreshToken() RefreshTokenRepository {
	return t.refreshRepo
}

// PasswordReset returns the password reset repository for transaction
func (t *transaction) PasswordReset() PasswordResetRepository {
	return t.passwordRepo
}

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.UserSession{},
		&models.RefreshToken{},
		&models.PasswordReset{},
	)
}