package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	KeyID          string     `gorm:"uniqueIndex;not null" json:"key_id"`              // Public identifier (ak_...)
	HashedSecret   string     `gorm:"not null" json:"-"`                               // Hashed secret, never returned in API
	Name           string     `gorm:"not null" json:"name"`                            // Human-readable name
	Description    string     `json:"description"`                                     // Optional description
	ClientAppID    *uuid.UUID `gorm:"type:uuid;index" json:"client_app_id"`            // Optional: link to OAuth client app
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`         // Owner of the API key
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"` // Organization isolation
	Scopes         string     `gorm:"type:text" json:"scopes"`                         // JSON array of allowed scopes
	ExpiresAt      *time.Time `json:"expires_at"`                                      // Optional expiration
	Revoked        bool       `gorm:"default:false" json:"revoked"`                    // Revocation status
	LastUsedAt     *time.Time `json:"last_used_at"`                                    // Track usage
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relationships
	User      User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ClientApp *ClientApp `gorm:"foreignKey:ClientAppID" json:"client_app,omitempty"`
}

// BeforeCreate sets up the API key before database insertion
func (a *APIKey) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for GORM
func (APIKey) TableName() string {
	return "api_keys"
}

// IsExpired checks if the API key has expired
func (a *APIKey) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// IsActive checks if the API key is active (not revoked and not expired)
func (a *APIKey) IsActive() bool {
	return !a.Revoked && !a.IsExpired()
}

// GetScopes returns the scopes as a slice
func (a *APIKey) GetScopes() []string {
	if a.Scopes == "" {
		return []string{}
	}
	// In a real implementation, this would parse JSON
	// For now, we'll use comma-separated values
	// TODO: Implement proper JSON parsing
	return []string{a.Scopes}
}

// APIKeyCreateRequest represents the request to create an API key
type APIKeyCreateRequest struct {
	Name        string    `json:"name" binding:"required,min=1,max=100"`
	Description string    `json:"description" binding:"max=500"`
	ClientAppID string    `json:"client_app_id,omitempty"`
	Scopes      []string  `json:"scopes"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
}

// APIKeyResponse represents the API response for an API key (without secret)
type APIKeyResponse struct {
	ID          uuid.UUID  `json:"id"`
	KeyID       string     `json:"key_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ClientAppID *uuid.UUID `json:"client_app_id"`
	Scopes      []string   `json:"scopes"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Revoked     bool       `json:"revoked"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// APIKeyCreateResponse includes the secret (only returned once)
type APIKeyCreateResponse struct {
	APIKeyResponse
	Secret string `json:"secret"` // Plain text secret, only returned on creation
}

// ToResponse converts an APIKey to APIKeyResponse
func (a *APIKey) ToResponse() APIKeyResponse {
	return APIKeyResponse{
		ID:          a.ID,
		KeyID:       a.KeyID,
		Name:        a.Name,
		Description: a.Description,
		ClientAppID: a.ClientAppID,
		Scopes:      a.GetScopes(),
		ExpiresAt:   a.ExpiresAt,
		Revoked:     a.Revoked,
		LastUsedAt:  a.LastUsedAt,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}
