package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"auth-service/internal/models"
)

var (
	ErrPasswordResetNotFound = errors.New("password reset not found")
)

// passwordResetRepository implements PasswordResetRepository interface
type passwordResetRepository struct {
	db *gorm.DB
}

// NewPasswordResetRepository creates a new password reset repository
func NewPasswordResetRepository(db *gorm.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

// Create creates a new password reset request
func (r *passwordResetRepository) Create(ctx context.Context, reset *models.PasswordReset) error {
	if reset == nil {
		return errors.New("password reset cannot be nil")
	}

	return r.db.WithContext(ctx).Create(reset).Error
}

// GetByToken retrieves a password reset by token
func (r *passwordResetRepository) GetByToken(ctx context.Context, token string) (*models.PasswordReset, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	var reset models.PasswordReset
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&reset).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPasswordResetNotFound
	}
	return &reset, err
}

// GetByEmail retrieves the latest password reset for an email and tenant
func (r *passwordResetRepository) GetByEmail(ctx context.Context, email, tenantID string) (*models.PasswordReset, error) {
	if email == "" || tenantID == "" {
		return nil, errors.New("email and tenant ID are required")
	}

	var reset models.PasswordReset
	err := r.db.WithContext(ctx).Where("email = ? AND tenant_id = ?", email, tenantID).Order("created_at DESC").First(&reset).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPasswordResetNotFound
	}
	return &reset, err
}

// Update updates a password reset
func (r *passwordResetRepository) Update(ctx context.Context, reset *models.PasswordReset) error {
	if reset == nil || reset.TokenHash == "" {
		return errors.New("password reset and token are required")
	}

	result := r.db.WithContext(ctx).Model(reset).Where("token_hash = ?", reset.TokenHash).Updates(reset)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPasswordResetNotFound
	}
	return nil
}

// Delete deletes a password reset by token
func (r *passwordResetRepository) Delete(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token is required")
	}

	result := r.db.WithContext(ctx).Delete(&models.PasswordReset{}, "token = ?", token)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPasswordResetNotFound
	}
	return nil
}

// DeleteExpired deletes expired password resets
func (r *passwordResetRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Delete(&models.PasswordReset{}, "expires_at <= ?", time.Now()).Error
}

// CleanupExpired removes expired password resets older than maxAge
func (r *passwordResetRepository) CleanupExpired(ctx context.Context, maxAge time.Duration) error {
	cutoffTime := time.Now().Add(-maxAge)
	return r.db.WithContext(ctx).Delete(&models.PasswordReset{}, "expires_at <= ?", cutoffTime).Error
}