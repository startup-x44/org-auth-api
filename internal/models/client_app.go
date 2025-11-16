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
	Code                string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"code"`
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
	Token          string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"` // hashed
	ClientID       string     `gorm:"type:varchar(255);not null;index" json:"client_id"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	OrganizationID *uuid.UUID `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	Scope          string     `gorm:"type:text" json:"scope"`
	ExpiresAt      time.Time  `gorm:"not null;index" json:"expires_at"`
	Revoked        bool       `gorm:"default:false;index" json:"revoked"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TableName specifies the table name for OAuthRefreshToken
func (OAuthRefreshToken) TableName() string {
	return "oauth_refresh_tokens"
}
