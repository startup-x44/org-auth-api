package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ClientApp represents an OAuth2 client application
type ClientApp struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string         `gorm:"type:varchar(255);not null" json:"name"`
	ClientID       string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"client_id"`
	ClientSecret   string         `gorm:"type:varchar(255);not null" json:"-"` // bcrypt hashed
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	RedirectURIs   pq.StringArray `gorm:"type:text[]" json:"redirect_uris"`
	AllowedOrigins pq.StringArray `gorm:"type:text[]" json:"allowed_origins"`
	AllowedScopes  pq.StringArray `gorm:"type:text[]" json:"allowed_scopes"`
	IsConfidential bool           `gorm:"default:true" json:"is_confidential"` // true = requires secret, false = public (PKCE only)
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// TableName specifies the table name for ClientApp
func (ClientApp) TableName() string {
	return "client_apps"
}

// AuthorizationCode represents an OAuth2 authorization code
type AuthorizationCode struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CodeHash            string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"` // HMAC-SHA256 hash of code
	ClientID            string     `gorm:"type:varchar(255);not null;index" json:"client_id"`
	UserID              uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	OrganizationID      *uuid.UUID `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	RedirectURI         string     `gorm:"type:text;not null" json:"redirect_uri"`
	Scope               string     `gorm:"type:text" json:"scope"`
	CodeChallenge       string     `gorm:"type:varchar(255)" json:"code_challenge,omitempty"`
	CodeChallengeMethod string     `gorm:"type:varchar(10)" json:"code_challenge_method,omitempty"`
	ExpiresAt           time.Time  `gorm:"not null;index" json:"expires_at"`
	Used                bool       `gorm:"default:false;index" json:"used"`
	CreatedAt           time.Time  `json:"created_at"`
}

// TableName specifies the table name for AuthorizationCode
func (AuthorizationCode) TableName() string {
	return "authorization_codes"
}

// OAuthRefreshToken represents a refresh token for OAuth2
type OAuthRefreshToken struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TokenHash      string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"` // HMAC-SHA256 hash of token
	FamilyID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"family_id"`      // Groups all tokens in rotation chain
	ClientID       string     `gorm:"type:varchar(255);not null;index" json:"client_id"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	OrganizationID *uuid.UUID `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	Scope          string     `gorm:"type:text" json:"scope"`
	UserAgentHash  string     `gorm:"type:varchar(64);index" json:"-"`                    // SHA256 hash of user agent for binding
	IPHash         string     `gorm:"type:varchar(64);index" json:"-"`                    // SHA256 hash of IP for binding
	DeviceID       string     `gorm:"type:varchar(255);index" json:"device_id,omitempty"` // Optional device identifier
	ExpiresAt      time.Time  `gorm:"not null;index" json:"expires_at"`
	Revoked        bool       `gorm:"default:false;index" json:"revoked"`
	UsedAt         *time.Time `gorm:"index" json:"used_at,omitempty"`                  // Track when token was rotated
	ReplacedByID   *uuid.UUID `gorm:"type:uuid;index" json:"replaced_by_id,omitempty"` // ID of new token after rotation
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TableName specifies the table name for OAuthRefreshToken
func (OAuthRefreshToken) TableName() string {
	return "oauth_refresh_tokens"
}
