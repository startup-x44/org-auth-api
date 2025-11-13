package repository

import (
	"context"
	"errors"
	"time"

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

// GetByToken retrieves a refresh token by token string
func (r *refreshTokenRepository) GetByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	var refreshToken models.RefreshToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&refreshToken).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRefreshTokenNotFound
	}
	return &refreshToken, err
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

// Update updates a refresh token
func (r *refreshTokenRepository) Update(ctx context.Context, token *models.RefreshToken) error {
	if token == nil || token.Token == "" {
		return errors.New("refresh token and token string are required")
	}

	result := r.db.WithContext(ctx).Model(token).Where("token = ?", token.Token).Updates(token)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRefreshTokenNotFound
	}
	return nil
}

// Delete deletes a refresh token
func (r *refreshTokenRepository) Delete(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token is required")
	}

	result := r.db.WithContext(ctx).Delete(&models.RefreshToken{}, "token = ?", token)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRefreshTokenNotFound
	}
	return nil
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