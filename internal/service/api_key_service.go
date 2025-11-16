package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/password"

	"github.com/google/uuid"
)

// ErrInvalidAPIKeyID is returned when an API key ID format is invalid
var ErrInvalidAPIKeyID = errors.New("invalid API key ID format")

// ValidatePublicKeyID validates the format of a public API key ID
func ValidatePublicKeyID(keyID string) error {
	// Must start with "ak_"
	if !strings.HasPrefix(keyID, "ak_") {
		return ErrInvalidAPIKeyID
	}

	// Must be exactly 35 characters total (ak_ + 32 hex chars)
	if len(keyID) != 35 {
		return ErrInvalidAPIKeyID
	}

	// Check that the part after "ak_" contains only hex characters
	hexPart := keyID[3:]
	for _, r := range hexPart {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return ErrInvalidAPIKeyID
		}
	}

	return nil
}

// APIKeyService handles business logic for API keys
type APIKeyService interface {
	CreateAPIKey(ctx context.Context, userID, tenantID uuid.UUID, req *models.APIKeyCreateRequest) (*models.APIKeyCreateResponse, error)
	GetAPIKey(ctx context.Context, keyID string, userID, tenantID uuid.UUID) (*models.APIKeyResponse, error)
	ListAPIKeys(ctx context.Context, userID, tenantID uuid.UUID) ([]*models.APIKeyResponse, error)
	RevokeAPIKey(ctx context.Context, keyID string, userID, tenantID uuid.UUID) error
	ValidateAPIKey(ctx context.Context, keyWithSecret string) (*models.APIKey, error)
	UpdateLastUsed(ctx context.Context, keyID string) error
}

type apiKeyService struct {
	apiKeyRepo repository.APIKeyRepository
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(apiKeyRepo repository.APIKeyRepository) APIKeyService {
	return &apiKeyService{
		apiKeyRepo: apiKeyRepo,
	}
}

// CreateAPIKey creates a new API key for a user
func (s *apiKeyService) CreateAPIKey(ctx context.Context, userID, tenantID uuid.UUID, req *models.APIKeyCreateRequest) (*models.APIKeyCreateResponse, error) {
	// Generate unique key ID and secret
	keyID := repository.GenerateAPIKeyID()
	secret := repository.GenerateAPIKeySecret()

	// Hash the secret for storage (using Argon2id directly without password validation)
	pwdService := password.NewService()
	hashedSecret, err := pwdService.HashWithoutValidation(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to hash API key secret: %w", err)
	}

	// Parse client app ID if provided
	var clientAppID *uuid.UUID
	if req.ClientAppID != "" {
		parsedID, err := uuid.Parse(req.ClientAppID)
		if err != nil {
			return nil, fmt.Errorf("invalid client_app_id: %w", err)
		}
		clientAppID = &parsedID
	}

	// Parse expiration if provided
	var expiresAt *time.Time
	if !req.ExpiresAt.IsZero() {
		expiresAt = &req.ExpiresAt
	}

	// Convert scopes to string (simple comma-separated for now)
	scopesStr := strings.Join(req.Scopes, ",")

	// Create API key model
	apiKey := &models.APIKey{
		KeyID:        keyID,
		HashedSecret: hashedSecret,
		Name:         req.Name,
		Description:  req.Description,
		ClientAppID:  clientAppID,
		UserID:       userID,
		TenantID:     tenantID,
		Scopes:       scopesStr,
		ExpiresAt:    expiresAt,
		Revoked:      false,
	}

	// Save to database
	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Return response with secret (only time it's exposed)
	return &models.APIKeyCreateResponse{
		APIKeyResponse: apiKey.ToResponse(),
		Secret:         repository.FormatAPIKey(keyID, secret),
	}, nil
}

// GetAPIKey retrieves an API key by its public key ID
func (s *apiKeyService) GetAPIKey(ctx context.Context, keyID string, userID, tenantID uuid.UUID) (*models.APIKeyResponse, error) {
	// Validate public key ID format
	if err := ValidatePublicKeyID(keyID); err != nil {
		return nil, err
	}

	apiKey, err := s.apiKeyRepo.GetByKeyIDWithOwnership(ctx, keyID, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	response := apiKey.ToResponse()
	return &response, nil
}

// ListAPIKeys retrieves all API keys for a user
func (s *apiKeyService) ListAPIKeys(ctx context.Context, userID, tenantID uuid.UUID) ([]*models.APIKeyResponse, error) {
	apiKeys, err := s.apiKeyRepo.ListByUser(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	responses := make([]*models.APIKeyResponse, len(apiKeys))
	for i, apiKey := range apiKeys {
		response := apiKey.ToResponse()
		responses[i] = &response
	}

	return responses, nil
}

// RevokeAPIKey revokes an API key
func (s *apiKeyService) RevokeAPIKey(ctx context.Context, keyID string, userID, tenantID uuid.UUID) error {
	// Validate public key ID format
	if err := ValidatePublicKeyID(keyID); err != nil {
		return err
	}

	// First get the API key to get its database UUID
	apiKey, err := s.apiKeyRepo.GetByKeyIDWithOwnership(ctx, keyID, userID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}

	// Revoke using the database UUID
	if err := s.apiKeyRepo.Revoke(ctx, apiKey.ID, userID, tenantID); err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	return nil
}

// ValidateAPIKey validates an API key and returns the key details if valid
func (s *apiKeyService) ValidateAPIKey(ctx context.Context, keyWithSecret string) (*models.APIKey, error) {
	// Parse the API key format: "ak_xxx.secret"
	parts := strings.SplitN(keyWithSecret, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid API key format")
	}

	keyID := parts[0]
	secret := parts[1]

	// Validate key ID format
	if err := ValidatePublicKeyID(keyID); err != nil {
		return nil, err
	}

	// Get API key from database (no ownership check for validation - this is for authentication)
	apiKey, err := s.apiKeyRepo.GetByKeyID(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("API key not found")
	}

	// Check if key is active
	if !apiKey.IsActive() {
		return nil, fmt.Errorf("API key is revoked or expired")
	}

	// Verify secret
	pwdService := password.NewService()
	isValid, err := pwdService.Verify(secret, apiKey.HashedSecret)
	if err != nil || !isValid {
		return nil, fmt.Errorf("invalid API key secret")
	}

	return apiKey, nil
}

// UpdateLastUsed updates the last used timestamp for an API key
func (s *apiKeyService) UpdateLastUsed(ctx context.Context, keyID string) error {
	// Get API key by public keyID to get the database UUID
	apiKey, err := s.apiKeyRepo.GetByKeyID(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}

	// Update using the database UUID
	return s.apiKeyRepo.UpdateLastUsed(ctx, apiKey.ID)
}
