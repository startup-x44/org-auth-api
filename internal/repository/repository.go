package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"auth-service/internal/models"
)

// repository implements Repository interface
type repository struct {
	db                *gorm.DB
	userRepo          UserRepository
	tenantRepo        TenantRepository
	orgRepo           OrganizationRepository
	orgMembershipRepo OrganizationMembershipRepository
	orgInvitationRepo OrganizationInvitationRepository
	sessionRepo       UserSessionRepository
	refreshRepo       RefreshTokenRepository
	passwordRepo      PasswordResetRepository
	failedAttemptRepo FailedLoginAttemptRepository
}

// NewRepository creates a new repository instance
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db:                db,
		userRepo:          NewUserRepository(db),
		tenantRepo:        NewTenantRepository(db),
		orgRepo:           NewOrganizationRepository(db),
		orgMembershipRepo: NewOrganizationMembershipRepository(db),
		orgInvitationRepo: NewOrganizationInvitationRepository(db),
		sessionRepo:       NewUserSessionRepository(db),
		refreshRepo:       NewRefreshTokenRepository(db),
		passwordRepo:      NewPasswordResetRepository(db),
		failedAttemptRepo: NewFailedLoginAttemptRepository(db),
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

// Organization returns the organization repository
func (r *repository) Organization() OrganizationRepository {
	return r.orgRepo
}

// OrganizationMembership returns the organization membership repository
func (r *repository) OrganizationMembership() OrganizationMembershipRepository {
	return r.orgMembershipRepo
}

// OrganizationInvitation returns the organization invitation repository
func (r *repository) OrganizationInvitation() OrganizationInvitationRepository {
	return r.orgInvitationRepo
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

// FailedLoginAttempt returns the failed login attempt repository
func (r *repository) FailedLoginAttempt() FailedLoginAttemptRepository {
	return r.failedAttemptRepo
}

// BeginTransaction starts a new database transaction
func (r *repository) BeginTransaction(ctx context.Context) (Transaction, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &transaction{
		tx:                tx,
		userRepo:          NewUserRepository(tx),
		tenantRepo:        NewTenantRepository(tx),
		orgRepo:           NewOrganizationRepository(tx),
		orgMembershipRepo: NewOrganizationMembershipRepository(tx),
		orgInvitationRepo: NewOrganizationInvitationRepository(tx),
		sessionRepo:       NewUserSessionRepository(tx),
		refreshRepo:       NewRefreshTokenRepository(tx),
		passwordRepo:      NewPasswordResetRepository(tx),
		failedAttemptRepo: NewFailedLoginAttemptRepository(tx),
	}, nil
}

// transaction implements Transaction interface
type transaction struct {
	tx                *gorm.DB
	userRepo          UserRepository
	tenantRepo        TenantRepository
	orgRepo           OrganizationRepository
	orgMembershipRepo OrganizationMembershipRepository
	orgInvitationRepo OrganizationInvitationRepository
	sessionRepo       UserSessionRepository
	refreshRepo       RefreshTokenRepository
	passwordRepo      PasswordResetRepository
	failedAttemptRepo FailedLoginAttemptRepository
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

// Organization returns the organization repository for transaction
func (t *transaction) Organization() OrganizationRepository {
	return t.orgRepo
}

// OrganizationMembership returns the organization membership repository for transaction
func (t *transaction) OrganizationMembership() OrganizationMembershipRepository {
	return t.orgMembershipRepo
}

// OrganizationInvitation returns the organization invitation repository for transaction
func (t *transaction) OrganizationInvitation() OrganizationInvitationRepository {
	return t.orgInvitationRepo
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

// FailedLoginAttempt returns the failed login attempt repository for transaction
func (t *transaction) FailedLoginAttempt() FailedLoginAttemptRepository {
	return t.failedAttemptRepo
}

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	// Auto migrate all models
	if err := db.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.Organization{},
		&models.OrganizationMembership{},
		&models.OrganizationInvitation{},
		&models.UserSession{},
		&models.RefreshToken{},
		&models.PasswordReset{},
		&models.FailedLoginAttempt{},
	); err != nil {
		return err
	}

	// Create composite indexes
	return createIndexes(db)
}

// createIndexes creates additional indexes not covered by AutoMigrate
func createIndexes(db *gorm.DB) error {
	// Composite index for users(email, tenant_id) for faster tenant-specific email lookups
	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email_tenant ON users(email, tenant_id)").Error; err != nil {
		return fmt.Errorf("failed to create users email-tenant index: %w", err)
	}

	// Composite index for failed_login_attempts(email, ip_address, tenant_id, attempted_at) for account lockout queries
	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_failed_attempts_lookup ON failed_login_attempts(email, ip_address, tenant_id, attempted_at DESC)").Error; err != nil {
		return fmt.Errorf("failed to create failed login attempts index: %w", err)
	}

	// Composite index for user_sessions(user_id, expires_at) for session cleanup and user session queries
	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_user_expires ON user_sessions(user_id, expires_at)").Error; err != nil {
		return fmt.Errorf("failed to create user sessions index: %w", err)
	}

	// Composite index for organization_memberships(organization_id, user_id) for membership lookups
	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_memberships_org_user ON organization_memberships(organization_id, user_id)").Error; err != nil {
		return fmt.Errorf("failed to create organization memberships index: %w", err)
	}

	// Composite index for organization_memberships(user_id, organization_id) for user organization lookups
	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_memberships_user_org ON organization_memberships(user_id, organization_id)").Error; err != nil {
		return fmt.Errorf("failed to create user organization memberships index: %w", err)
	}

	// Composite index for organization_invitations(organization_id, email) for invitation lookups
	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_invitations_org_email ON organization_invitations(organization_id, email)").Error; err != nil {
		return fmt.Errorf("failed to create organization invitations index: %w", err)
	}

	// Index for organization_invitations(token_hash) for token lookups
	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_invitations_token ON organization_invitations(token_hash)").Error; err != nil {
		return fmt.Errorf("failed to create organization invitations token index: %w", err)
	}

	// Index for organization_invitations(expires_at) for cleanup
	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_invitations_expires ON organization_invitations(expires_at)").Error; err != nil {
		return fmt.Errorf("failed to create organization invitations expires index: %w", err)
	}

	return nil
}
