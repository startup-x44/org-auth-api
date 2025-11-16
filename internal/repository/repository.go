package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"auth-service/internal/models"
)

// repository implements Repository interface
type repository struct {
	db                *gorm.DB
	userRepo          UserRepository
	orgRepo           OrganizationRepository
	orgMembershipRepo OrganizationMembershipRepository
	orgInvitationRepo OrganizationInvitationRepository
	sessionRepo       UserSessionRepository
	refreshRepo       RefreshTokenRepository
	passwordRepo      PasswordResetRepository
	failedAttemptRepo FailedLoginAttemptRepository
	rolePermRepo      RolePermissionRepository
	roleRepo          RoleRepository
	permRepo          PermissionRepository
}

// NewRepository creates a new repository instance
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db:                db,
		userRepo:          NewUserRepository(db),
		orgRepo:           NewOrganizationRepository(db),
		orgMembershipRepo: NewOrganizationMembershipRepository(db),
		orgInvitationRepo: NewOrganizationInvitationRepository(db),
		sessionRepo:       NewUserSessionRepository(db),
		refreshRepo:       NewRefreshTokenRepository(db),
		passwordRepo:      NewPasswordResetRepository(db),
		failedAttemptRepo: NewFailedLoginAttemptRepository(db),
		rolePermRepo:      NewRolePermissionRepository(db),
		roleRepo:          NewRoleRepository(db),
		permRepo:          NewPermissionRepository(db),
	}
}

// User returns the user repository
func (r *repository) User() UserRepository {
	return r.userRepo
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

// RolePermission returns the role permission repository
func (r *repository) RolePermission() RolePermissionRepository {
	return r.rolePermRepo
}

// Role returns the role repository
func (r *repository) Role() RoleRepository {
	return r.roleRepo
}

// Permission returns the permission repository
func (r *repository) Permission() PermissionRepository {
	return r.permRepo
}

// CreateDefaultAdminRole creates the default OWNER role for an organization with all permissions
// This is a CUSTOM role (IsSystem=false) specific to the organization
func (r *repository) CreateDefaultAdminRole(ctx context.Context, orgID, createdBy string) (*models.Role, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	createdByUUID, err := uuid.Parse(createdBy)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Create OWNER role - this is a CUSTOM role, NOT a system role
	ownerRole := &models.Role{
		OrganizationID: &orgUUID,
		Name:           "owner", // Use "owner" instead of "admin" to differentiate from system roles
		DisplayName:    "Owner",
		Description:    "Full access to all organization features (organization owner)",
		IsSystem:       false, // CUSTOM role, not system
		CreatedBy:      createdByUUID,
	}

	if err := r.db.WithContext(ctx).Create(ownerRole).Error; err != nil {
		return nil, fmt.Errorf("failed to create owner role: %w", err)
	}

	// Get all SYSTEM permissions (these are global and available to all orgs)
	var permissions []models.Permission
	adminPermNames := models.DefaultAdminPermissions()
	if err := r.db.WithContext(ctx).
		Where("name IN ? AND is_system = ? AND organization_id IS NULL", adminPermNames, true).
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	// Assign all system permissions to owner role
	for _, perm := range permissions {
		rolePermission := &models.RolePermission{
			RoleID:       ownerRole.ID,
			PermissionID: perm.ID,
		}
		if err := r.db.WithContext(ctx).Create(rolePermission).Error; err != nil {
			return nil, fmt.Errorf("failed to assign permission %s to owner: %w", perm.Name, err)
		}
	}

	return ownerRole, nil
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
		orgRepo:           NewOrganizationRepository(tx),
		orgMembershipRepo: NewOrganizationMembershipRepository(tx),
		orgInvitationRepo: NewOrganizationInvitationRepository(tx),
		sessionRepo:       NewUserSessionRepository(tx),
		refreshRepo:       NewRefreshTokenRepository(tx),
		passwordRepo:      NewPasswordResetRepository(tx),
		failedAttemptRepo: NewFailedLoginAttemptRepository(tx),
		rolePermRepo:      NewRolePermissionRepository(tx),
		roleRepo:          NewRoleRepository(tx),
		permRepo:          NewPermissionRepository(tx),
	}, nil
}

// transaction implements Transaction interface
type transaction struct {
	tx                *gorm.DB
	userRepo          UserRepository
	orgRepo           OrganizationRepository
	orgMembershipRepo OrganizationMembershipRepository
	orgInvitationRepo OrganizationInvitationRepository
	sessionRepo       UserSessionRepository
	refreshRepo       RefreshTokenRepository
	passwordRepo      PasswordResetRepository
	failedAttemptRepo FailedLoginAttemptRepository
	rolePermRepo      RolePermissionRepository
	roleRepo          RoleRepository
	permRepo          PermissionRepository
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

// RolePermission returns the role permission repository for transaction
func (t *transaction) RolePermission() RolePermissionRepository {
	return t.rolePermRepo
}

// Role returns the role repository for transaction
func (t *transaction) Role() RoleRepository {
	return t.roleRepo
}

// Permission returns the permission repository for transaction
func (t *transaction) Permission() PermissionRepository {
	return t.permRepo
}

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	// Auto migrate all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationMembership{},
		&models.OrganizationInvitation{},
		&models.UserSession{},
		&models.RefreshToken{},
		&models.PasswordReset{},
		&models.FailedLoginAttempt{},
		&models.Permission{},     // Global system permissions
		&models.Role{},           // Organization-specific roles
		&models.RolePermission{}, // Role-Permission many-to-many
	); err != nil {
		return err
	}

	// Seed global system permissions
	if err := seedGlobalPermissions(db); err != nil {
		return fmt.Errorf("failed to seed permissions: %w", err)
	}

	// Create composite indexes
	return createIndexes(db)
}

// createIndexes creates additional indexes not covered by AutoMigrate
func createIndexes(db *gorm.DB) error {
	// Composite index for failed_login_attempts(email, ip_address, attempted_at) for account lockout queries
	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_failed_attempts_lookup ON failed_login_attempts(email, ip_address, attempted_at DESC)").Error; err != nil {
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

// seedGlobalPermissions seeds the global system permissions if they don't exist
func seedGlobalPermissions(db *gorm.DB) error {
	permissions := models.DefaultPermissions()

	for _, perm := range permissions {
		var existing models.Permission
		err := db.Where("name = ?", perm.Name).First(&existing).Error

		// If not found, create it
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&perm).Error; err != nil {
				return fmt.Errorf("failed to create permission %s: %w", perm.Name, err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to check permission: %w", err)
		}
	}

	return nil
}
