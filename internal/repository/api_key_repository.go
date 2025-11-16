package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"auth-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// APIKeyRepository handles database operations for API keys
type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *models.APIKey) error
	GetByKeyID(ctx context.Context, keyID string) (*models.APIKey, error)
	GetByKeyIDWithOwnership(ctx context.Context, keyID string, userID, tenantID uuid.UUID) (*models.APIKey, error)
	GetByID(ctx context.Context, id uuid.UUID, userID, tenantID uuid.UUID) (*models.APIKey, error)
	ListByUser(ctx context.Context, userID, tenantID uuid.UUID) ([]*models.APIKey, error)
	ListByClientApp(ctx context.Context, clientAppID, tenantID uuid.UUID) ([]*models.APIKey, error)
	Update(ctx context.Context, apiKey *models.APIKey) error
	Revoke(ctx context.Context, id uuid.UUID, userID, tenantID uuid.UUID) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type apiKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository {
	return &apiKeyRepository{db: db}
}

// Create inserts a new API key into the database
func (r *apiKeyRepository) Create(ctx context.Context, apiKey *models.APIKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

// GetByKeyID retrieves an API key by its public key ID (for validation purposes)
func (r *apiKeyRepository) GetByKeyID(ctx context.Context, keyID string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.WithContext(ctx).
		Where("key_id = ? AND revoked = false", keyID).
		First(&apiKey).Error

	if err != nil {
		return nil, err
	}

	return &apiKey, nil
}

// GetByKeyIDWithOwnership retrieves an API key by its public key ID with ownership validation
func (r *apiKeyRepository) GetByKeyIDWithOwnership(ctx context.Context, keyID string, userID, tenantID uuid.UUID) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.WithContext(ctx).
		Where("key_id = ? AND user_id = ? AND tenant_id = ?", keyID, userID, tenantID).
		First(&apiKey).Error

	if err != nil {
		return nil, err
	}

	return &apiKey, nil
}

// GetByID retrieves an API key by its ID (with tenant isolation)
func (r *apiKeyRepository) GetByID(ctx context.Context, id uuid.UUID, userID, tenantID uuid.UUID) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND tenant_id = ?", id, userID, tenantID).
		First(&apiKey).Error

	if err != nil {
		return nil, err
	}

	return &apiKey, nil
}

// ListByUser retrieves all API keys for a specific user (with tenant isolation)
func (r *apiKeyRepository) ListByUser(ctx context.Context, userID, tenantID uuid.UUID) ([]*models.APIKey, error) {
	var apiKeys []*models.APIKey
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Order("created_at DESC").
		Find(&apiKeys).Error

	return apiKeys, err
}

// ListByClientApp retrieves all API keys for a specific client app (with tenant isolation)
func (r *apiKeyRepository) ListByClientApp(ctx context.Context, clientAppID, tenantID uuid.UUID) ([]*models.APIKey, error) {
	var apiKeys []*models.APIKey
	err := r.db.WithContext(ctx).
		Where("client_app_id = ? AND tenant_id = ?", clientAppID, tenantID).
		Order("created_at DESC").
		Find(&apiKeys).Error

	return apiKeys, err
}

// Update updates an existing API key
func (r *apiKeyRepository) Update(ctx context.Context, apiKey *models.APIKey) error {
	return r.db.WithContext(ctx).Save(apiKey).Error
}

// Revoke marks an API key as revoked (with tenant isolation)
func (r *apiKeyRepository) Revoke(ctx context.Context, id uuid.UUID, userID, tenantID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.APIKey{}).
		Where("id = ? AND user_id = ? AND tenant_id = ?", id, userID, tenantID).
		Updates(map[string]interface{}{
			"revoked":    true,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// UpdateLastUsed updates the last_used_at timestamp for an API key using its database UUID
func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.APIKey{}).
		Where("id = ?", id).
		Update("last_used_at", time.Now()).Error
}

// DeleteExpired removes expired API keys from the database
func (r *apiKeyRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&models.APIKey{}).Error
}

// GenerateAPIKeyID generates a unique public key ID with ak_ prefix
func GenerateAPIKeyID() string {
	// Generate 16 random bytes (32 hex characters)
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "ak_" + hex.EncodeToString(bytes)
}

// GenerateAPIKeySecret generates a secure random secret for API keys
func GenerateAPIKeySecret() string {
	// Generate 32 random bytes (64 hex characters)
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// FormatAPIKey formats the key ID and secret as a complete API key
func FormatAPIKey(keyID, secret string) string {
	return fmt.Sprintf("%s.%s", keyID, secret)
}
