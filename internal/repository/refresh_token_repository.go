package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"auth-service/internal/models"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
)

// refreshTokenRepository implements RefreshTokenRepository interface
type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// Create creates a new refresh token
func (r *refreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	if token == nil {
		return errors.New("refresh token cannot be nil")
	}

	return r.db.WithContext(ctx).Create(token).Error
}

// GetByID retrieves a refresh token by its primary key
func (r *refreshTokenRepository) GetByID(ctx context.Context, id string) (*models.RefreshToken, error) {
	if id == "" {
		return nil, errors.New("token ID is required")
	}

	var refreshToken models.RefreshToken
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&refreshToken).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRefreshTokenNotFound
	}
	return &refreshToken, err
}

// GetActiveBySession retrieves non-revoked refresh tokens for a session
func (r *refreshTokenRepository) GetActiveBySession(ctx context.Context, sessionID string) ([]*models.RefreshToken, error) {
	if sessionID == "" {
		return nil, errors.New("session ID is required")
	}

	var tokens []*models.RefreshToken
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND revoked_at IS NULL", sessionID).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}

// GetByUserID retrieves all refresh tokens for a user
func (r *refreshTokenRepository) GetByUserID(ctx context.Context, userID string) ([]*models.RefreshToken, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	var tokens []*models.RefreshToken
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&tokens).Error
	return tokens, err
}

// Update updates a refresh token record (used for rotation metadata)
func (r *refreshTokenRepository) Update(ctx context.Context, token *models.RefreshToken) error {
	if token == nil || token.ID == uuid.Nil {
		return errors.New("refresh token and ID are required")
	}

	result := r.db.WithContext(ctx).Model(&models.RefreshToken{}).Where("id = ?", token.ID).Updates(token)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRefreshTokenNotFound
	}
	return nil
}

// Revoke marks a refresh token as revoked with reason
func (r *refreshTokenRepository) Revoke(ctx context.Context, id string, reason string) error {
	if id == "" {
		return errors.New("token ID is required")
	}

	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"revoked_at":     &now,
			"revoked_reason": reason,
			"updated_at":     now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRefreshTokenNotFound
	}
	return nil
}

// RevokeBySession revokes all refresh tokens tied to a session
func (r *refreshTokenRepository) RevokeBySession(ctx context.Context, sessionID string, reason string) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}

	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"revoked_at":     &now,
			"revoked_reason": reason,
			"updated_at":     now,
		}).Error
}

// DeleteByUserID deletes all refresh tokens for a user
func (r *refreshTokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	return r.db.WithContext(ctx).Delete(&models.RefreshToken{}, "user_id = ?", userID).Error
}

// DeleteExpired deletes expired refresh tokens
func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Delete(&models.RefreshToken{}, "expires_at <= ?", time.Now()).Error
}

// CleanupExpired removes expired refresh tokens older than maxAge
func (r *refreshTokenRepository) CleanupExpired(ctx context.Context, maxAge time.Duration) error {
	cutoffTime := time.Now().Add(-maxAge)
	return r.db.WithContext(ctx).Delete(&models.RefreshToken{}, "expires_at <= ?", cutoffTime).Error
}
