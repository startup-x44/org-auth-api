package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"auth-service/internal/models"
)

// failedLoginAttemptRepository implements FailedLoginAttemptRepository interface
type failedLoginAttemptRepository struct {
	db *gorm.DB
}

// NewFailedLoginAttemptRepository creates a new failed login attempt repository
func NewFailedLoginAttemptRepository(db *gorm.DB) FailedLoginAttemptRepository {
	return &failedLoginAttemptRepository{
		db: db,
	}
}

// Create creates a new failed login attempt record
func (r *failedLoginAttemptRepository) Create(ctx context.Context, attempt *models.FailedLoginAttempt) error {
	return r.db.WithContext(ctx).Create(attempt).Error
}

// GetByEmailAndIP gets failed login attempts by email, tenant, and IP within a time window
func (r *failedLoginAttemptRepository) GetByEmailAndIP(ctx context.Context, email, tenantID, ipAddress string, since time.Time) ([]*models.FailedLoginAttempt, error) {
	var attempts []*models.FailedLoginAttempt
	err := r.db.WithContext(ctx).
		Where("email = ? AND tenant_id = ? AND ip_address = ? AND attempted_at >= ?",
			email, tenantID, ipAddress, since).
		Order("attempted_at DESC").
		Find(&attempts).Error
	return attempts, err
}

// CountByEmailAndIP counts failed login attempts by email, tenant, and IP within a time window
func (r *failedLoginAttemptRepository) CountByEmailAndIP(ctx context.Context, email, tenantID, ipAddress string, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.FailedLoginAttempt{}).
		Where("email = ? AND tenant_id = ? AND ip_address = ? AND attempted_at >= ?",
			email, tenantID, ipAddress, since).
		Count(&count).Error
	return count, err
}

// DeleteExpired deletes expired failed login attempts
func (r *failedLoginAttemptRepository) DeleteExpired(ctx context.Context, maxAge time.Duration) error {
	return r.db.WithContext(ctx).
		Where("attempted_at < ?", time.Now().Add(-maxAge)).
		Delete(&models.FailedLoginAttempt{}).Error
}

// CleanupExpired removes failed login attempts older than maxAge
func (r *failedLoginAttemptRepository) CleanupExpired(ctx context.Context, maxAge time.Duration) error {
	return r.db.WithContext(ctx).
		Where("attempted_at < ?", time.Now().Add(-maxAge)).
		Delete(&models.FailedLoginAttempt{}).Error
}
