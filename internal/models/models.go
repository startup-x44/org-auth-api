package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID         uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	Email            string     `json:"email" gorm:"uniqueIndex:idx_users_tenant_email;not null"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at" gorm:"default:null"`
	PasswordHash     string     `json:"-" gorm:"not null"` // Never expose in JSON
	Firstname        *string    `json:"firstname" gorm:"size:100"`
	Lastname         *string    `json:"lastname" gorm:"size:100"`
	Address          *string    `json:"address" gorm:"type:text"`
	Phone            *string    `json:"phone" gorm:"size:20"`
	UserType         string     `json:"user_type" gorm:"index:idx_users_type;not null"` // Admin, Student, RTO, Issuer, Validator, badger, Non-partner, Partner
	Status           string     `json:"status" gorm:"index:idx_users_status;default:'active'"` // active, suspended, deactivated
	LastLoginAt      *time.Time `json:"last_login_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Relations
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"not null"`
	Domain    string    `json:"domain" gorm:"unique;not null"`
	Status    string    `json:"status" gorm:"default:'active'"` // active, suspended
	Settings  string    `json:"settings" gorm:"type:jsonb;default:'{}'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// UserSession represents an active user session
type UserSession struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     uuid.UUID  `json:"user_id" gorm:"type:uuid;index:idx_sessions_user;not null"`
	TenantID   uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	TokenHash  string     `json:"-" gorm:"uniqueIndex:idx_sessions_token;not null"` // Never expose in JSON
	IPAddress  string     `json:"-" gorm:"type:inet"`                               // Never expose in JSON
	UserAgent  string     `json:"-" gorm:"type:text"`                              // Never expose in JSON
	ExpiresAt  time.Time  `json:"-" gorm:"index:idx_sessions_expires;not null"`   // Never expose in JSON
	RevokedAt  *time.Time `json:"-"`                                              // Never expose in JSON
	CreatedAt  time.Time  `json:"created_at"`

	// Relations
	User   *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (us *UserSession) BeforeCreate(tx *gorm.DB) error {
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}
	return nil
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	TenantID  uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	TokenHash string     `json:"-" gorm:"unique;not null"` // Never expose in JSON
	ExpiresAt time.Time  `json:"-" gorm:"not null"`        // Never expose in JSON
	RevokedAt *time.Time `json:"-"`                        // Never expose in JSON
	CreatedAt time.Time  `json:"created_at"`

	// Relations
	User   *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

// PasswordReset represents a password reset request
type PasswordReset struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	TenantID  uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	TokenHash string     `json:"-" gorm:"unique;not null"` // Never expose in JSON
	ExpiresAt time.Time  `json:"-" gorm:"not null"`        // Never expose in JSON
	UsedAt    *time.Time `json:"-"`                        // Never expose in JSON
	CreatedAt time.Time  `json:"created_at"`

	// Relations
	User   *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (pr *PasswordReset) BeforeCreate(tx *gorm.DB) error {
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	return nil
}

// User types constants
const (
	UserTypeAdmin       = "Admin"
	UserTypeStudent     = "Student"
	UserTypeRTO         = "RTO"
	UserTypeIssuer      = "Issuer"
	UserTypeValidator   = "Validator"
	UserTypeBadger      = "badger"
	UserTypeNonPartner  = "Non-partner"
	UserTypePartner     = "Partner"
)

// User status constants
const (
	UserStatusActive      = "active"
	UserStatusSuspended   = "suspended"
	UserStatusDeactivated = "deactivated"
)

// Tenant status constants
const (
	TenantStatusActive    = "active"
	TenantStatusSuspended = "suspended"
)