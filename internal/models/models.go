package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system - GLOBAL USER (not organization-bound)
type User struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email           string     `json:"email" gorm:"uniqueIndex;not null"` // Global unique email
	EmailVerifiedAt *time.Time `json:"email_verified_at" gorm:"default:null"`
	PasswordHash    string     `json:"-" gorm:"not null"` // Never expose in JSON
	Firstname       *string    `json:"firstname" gorm:"size:100"`
	Lastname        *string    `json:"lastname" gorm:"size:100"`
	Address         *string    `json:"address" gorm:"type:text"`
	Phone           *string    `json:"phone" gorm:"size:20"`
	IsSuperadmin    bool       `json:"is_superadmin" gorm:"default:false"`                    // Global platform admin
	GlobalRole      string     `json:"global_role" gorm:"default:'user'"`                     // user, admin
	Status          string     `json:"status" gorm:"index:idx_users_status;default:'active'"` // active, suspended, deactivated
	LastLoginAt     *time.Time `json:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relations
	Organizations []*Organization `json:"organizations,omitempty" gorm:"many2many:organization_memberships;"`
}

func (User) TableName() string {
	return "users"
}

// BeforeCreate will set a UUID rather than numeric ID.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// BeforeSave normalizes email to lowercase
func (u *User) BeforeSave(tx *gorm.DB) error {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	return nil
}

// UserSession represents an active user session - ORG-SCOPED
type UserSession struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID            uuid.UUID  `json:"user_id" gorm:"type:uuid;index:idx_sessions_user;not null"`
	OrganizationID    uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null;index:idx_sessions_org"` // ORG-SCOPED SESSION
	TokenHash         string     `json:"-" gorm:"uniqueIndex:idx_sessions_token;not null"`                 // Never expose in JSON
	IPAddress         string     `json:"-" gorm:"type:inet"`                                               // Never expose in JSON
	UserAgent         string     `json:"-" gorm:"type:text"`                                               // Never expose in JSON
	DeviceFingerprint string     `json:"-" gorm:"type:text"`                                               // Device fingerprint for tracking
	Location          string     `json:"-" gorm:"type:text"`                                               // Geographic location (optional)
	IsActive          bool       `json:"-" gorm:"default:true"`                                            // Whether session is active
	LastActivity      time.Time  `json:"-" gorm:"index:idx_sessions_activity"`                             // Last activity timestamp
	ExpiresAt         time.Time  `json:"-" gorm:"index:idx_sessions_expires;not null"`                     // Never expose in JSON
	RevokedAt         *time.Time `json:"-"`                                                                // Never expose in JSON
	RevokedReason     string     `json:"-" gorm:"type:text"`                                               // Reason for revocation
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	// Relations
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (us *UserSession) BeforeCreate(tx *gorm.DB) error {
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}
	return nil
}

// RefreshToken represents a refresh token - ORG-SCOPED
type RefreshToken struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	OrganizationID uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null;index:idx_refresh_org"` // ORG-SCOPED TOKEN
	SessionID      uuid.UUID  `json:"session_id" gorm:"type:uuid;not null;index:idx_refresh_session"`
	TokenHash      string     `json:"-" gorm:"not null"` // Argon2 hash of refresh token
	ExpiresAt      time.Time  `json:"-" gorm:"not null"` // Never expose in JSON
	RevokedAt      *time.Time `json:"-"`
	RevokedReason  string     `json:"-" gorm:"type:text"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relations
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

// PasswordReset represents a password reset request - GLOBAL (user-scoped, not org-scoped)
type PasswordReset struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	TokenHash string     `json:"-" gorm:"unique;not null"` // Never expose in JSON
	ExpiresAt time.Time  `json:"-" gorm:"not null"`        // Never expose in JSON
	UsedAt    *time.Time `json:"-"`                        // Never expose in JSON
	CreatedAt time.Time  `json:"created_at"`

	// Relations
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (pr *PasswordReset) BeforeCreate(tx *gorm.DB) error {
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	return nil
}

// FailedLoginAttempt represents a failed login attempt for account lockout - GLOBAL (email-scoped)
type FailedLoginAttempt struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      *uuid.UUID `json:"user_id" gorm:"type:uuid;index:idx_failed_attempts_user"` // Nullable for attempts on non-existent users
	Email       string     `json:"email" gorm:"not null;index:idx_failed_attempts_email"`
	IPAddress   string     `json:"-" gorm:"type:inet;not null"` // Never expose in JSON
	UserAgent   string     `json:"-" gorm:"type:text"`          // Never expose in JSON
	AttemptedAt time.Time  `json:"-" gorm:"not null"`           // Never expose in JSON
	CreatedAt   time.Time  `json:"created_at"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (fla *FailedLoginAttempt) BeforeCreate(tx *gorm.DB) error {
	if fla.ID == uuid.Nil {
		fla.ID = uuid.New()
	}
	return nil
}

// User status constants
const (
	UserStatusActive      = "active"
	UserStatusSuspended   = "suspended"
	UserStatusDeactivated = "deactivated"
)

// Organization represents a Slack-style organization/workspace
type Organization struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"not null;size:100"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;not null;size:50"` // URL-friendly identifier
	Description *string   `json:"description" gorm:"type:text"`
	Settings    string    `json:"settings" gorm:"type:jsonb;default:'{}'"` // JSONB for flexible org settings
	Status      string    `json:"status" gorm:"default:'active'"`          // active, suspended, archived
	CreatedBy   uuid.UUID `json:"created_by" gorm:"type:uuid;not null"`    // User who created the org
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	Creator *User `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

// OrganizationMembership represents user membership in organizations
type OrganizationMembership struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrganizationID uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null;uniqueIndex:idx_unique_membership"`
	UserID         uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_unique_membership"`
	RoleID         uuid.UUID  `json:"role_id" gorm:"type:uuid;not null"` // Foreign key to roles table
	Status         string     `json:"status" gorm:"default:'active'"`    // active, invited, pending, suspended
	InvitedBy      *uuid.UUID `json:"invited_by" gorm:"type:uuid"`       // User who sent the invitation
	InvitedAt      *time.Time `json:"invited_at"`
	JoinedAt       *time.Time `json:"joined_at"`
	LastActivityAt *time.Time `json:"last_activity_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relations
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Role         *Role         `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	Inviter      *User         `json:"inviter,omitempty" gorm:"foreignKey:InvitedBy"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (om *OrganizationMembership) BeforeCreate(tx *gorm.DB) error {
	if om.ID == uuid.Nil {
		om.ID = uuid.New()
	}
	return nil
}

// OrganizationInvitation represents pending organization invitations
type OrganizationInvitation struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrganizationID uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null;uniqueIndex:idx_unique_invitation"`
	Email          string     `json:"email" gorm:"not null;uniqueIndex:idx_unique_invitation"`
	TokenHash      string     `json:"-" gorm:"unique;not null"`          // Never expose in JSON
	RoleID         uuid.UUID  `json:"role_id" gorm:"type:uuid;not null"` // Foreign key to roles table
	Status         string     `json:"status" gorm:"default:'pending'"`   // pending, accepted, expired, cancelled
	InvitedBy      uuid.UUID  `json:"invited_by" gorm:"type:uuid;not null"`
	ExpiresAt      time.Time  `json:"-" gorm:"not null"` // Never expose in JSON
	AcceptedAt     *time.Time `json:"accepted_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relations
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Role         *Role         `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	Inviter      *User         `json:"inviter,omitempty" gorm:"foreignKey:InvitedBy"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (oi *OrganizationInvitation) BeforeCreate(tx *gorm.DB) error {
	if oi.ID == uuid.Nil {
		oi.ID = uuid.New()
	}
	return nil
}

// BeforeSave normalizes email to lowercase
func (oi *OrganizationInvitation) BeforeSave(tx *gorm.DB) error {
	oi.Email = strings.ToLower(strings.TrimSpace(oi.Email))
	return nil
}

// Organization role constants
const (
	OrganizationRoleAdmin   = "admin"
	OrganizationRoleIssuer  = "issuer"
	OrganizationRoleRTO     = "rto"
	OrganizationRoleStudent = "student"
)

// Membership status constants
const (
	MembershipStatusActive    = "active"
	MembershipStatusInvited   = "invited"
	MembershipStatusPending   = "pending"
	MembershipStatusSuspended = "suspended"
)

// Invitation status constants
const (
	InvitationStatusPending   = "pending"
	InvitationStatusAccepted  = "accepted"
	InvitationStatusExpired   = "expired"
	InvitationStatusCancelled = "cancelled"
)

// Organization status constants
const (
	OrganizationStatusActive    = "active"
	OrganizationStatusSuspended = "suspended"
	OrganizationStatusArchived  = "archived"
)
