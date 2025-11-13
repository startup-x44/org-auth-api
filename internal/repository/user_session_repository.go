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
	ErrSessionNotFound  = errors.New("session not found")
	ErrInvalidSessionID = errors.New("invalid session ID")
)

// userSessionRepository implements UserSessionRepository interface
type userSessionRepository struct {
	db *gorm.DB
}

// NewUserSessionRepository creates a new user session repository
func NewUserSessionRepository(db *gorm.DB) UserSessionRepository {
	return &userSessionRepository{db: db}
}

// Create creates a new user session
func (r *userSessionRepository) Create(ctx context.Context, session *models.UserSession) error {
	if session == nil {
		return errors.New("session cannot be nil")
	}

	return r.db.WithContext(ctx).Create(session).Error
}

// GetByID retrieves a session by ID
func (r *userSessionRepository) GetByID(ctx context.Context, id string) (*models.UserSession, error) {
	if id == "" {
		return nil, ErrInvalidSessionID
	}

	var session models.UserSession
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSessionNotFound
	}
	return &session, err
}

// GetByUserID retrieves all sessions for a user
func (r *userSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*models.UserSession, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	var sessions []*models.UserSession
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&sessions).Error
	return sessions, err
}

// GetActiveByUserID retrieves active sessions for a user
func (r *userSessionRepository) GetActiveByUserID(ctx context.Context, userID string) ([]*models.UserSession, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	var sessions []*models.UserSession
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = true AND expires_at > ?", userID, time.Now()).Order("created_at DESC").Find(&sessions).Error
	return sessions, err
}

// Update updates a session
func (r *userSessionRepository) Update(ctx context.Context, session *models.UserSession) error {
	if session == nil || session.ID == uuid.Nil {
		return errors.New("session and session ID are required")
	}

	result := r.db.WithContext(ctx).Model(session).Updates(session)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// Delete deletes a session
func (r *userSessionRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidSessionID
	}

	result := r.db.WithContext(ctx).Delete(&models.UserSession{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// DeleteByUserID deletes all sessions for a user
func (r *userSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	return r.db.WithContext(ctx).Delete(&models.UserSession{}, "user_id = ?", userID).Error
}

// DeleteExpired deletes expired sessions
func (r *userSessionRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Delete(&models.UserSession{}, "expires_at <= ?", time.Now()).Error
}

// CleanupExpired removes expired sessions older than maxAge
func (r *userSessionRepository) CleanupExpired(ctx context.Context, maxAge time.Duration) error {
	cutoffTime := time.Now().Add(-maxAge)
	return r.db.WithContext(ctx).Delete(&models.UserSession{}, "expires_at <= ?", cutoffTime).Error
}

// GetByToken retrieves a session by token hash
func (r *userSessionRepository) GetByToken(ctx context.Context, token string) (*models.UserSession, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	var session models.UserSession
	err := r.db.WithContext(ctx).Where("token_hash = ?", token).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSessionNotFound
	}
	return &session, err
}

// GetActiveCountByUserID counts active sessions for a user
func (r *userSessionRepository) GetActiveCountByUserID(ctx context.Context, userID string) (int64, error) {
	if userID == "" {
		return 0, errors.New("user ID is required")
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserSession{}).
		Where("user_id = ? AND is_active = true AND expires_at > ?", userID, time.Now()).
		Count(&count).Error
	return count, err
}

// UpdateActivity updates the last activity timestamp for a session
func (r *userSessionRepository) UpdateActivity(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidSessionID
	}

	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.UserSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_activity": now,
			"updated_at":    now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// Revoke revokes a specific session
func (r *userSessionRepository) Revoke(ctx context.Context, id string, reason string) error {
	if id == "" {
		return ErrInvalidSessionID
	}

	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.UserSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":      false,
			"revoked_at":     &now,
			"revoked_reason": reason,
			"updated_at":     now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// RevokeByUserID revokes all sessions for a user
func (r *userSessionRepository) RevokeByUserID(ctx context.Context, userID string, reason string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.UserSession{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"is_active":      false,
			"revoked_at":     &now,
			"revoked_reason": reason,
			"updated_at":     now,
		}).Error
}

// RevokeExpired revokes all expired sessions
func (r *userSessionRepository) RevokeExpired(ctx context.Context) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.UserSession{}).
		Where("expires_at <= ? AND is_active = true", now).
		Updates(map[string]interface{}{
			"is_active":      false,
			"revoked_at":     &now,
			"revoked_reason": "expired",
			"updated_at":     now,
		}).Error
}

// RevokeInactive revokes sessions inactive for longer than maxInactive duration
func (r *userSessionRepository) RevokeInactive(ctx context.Context, maxInactive time.Duration) error {
	cutoffTime := time.Now().Add(-maxInactive)
	now := time.Now()

	return r.db.WithContext(ctx).Model(&models.UserSession{}).
		Where("last_activity <= ? AND is_active = true", cutoffTime).
		Updates(map[string]interface{}{
			"is_active":      false,
			"revoked_at":     &now,
			"revoked_reason": "inactive",
			"updated_at":     now,
		}).Error
}

// GetSessionsByIPAddress retrieves sessions by IP address within a time window
func (r *userSessionRepository) GetSessionsByIPAddress(ctx context.Context, ipAddress string, since time.Time) ([]*models.UserSession, error) {
	if ipAddress == "" {
		return nil, errors.New("IP address is required")
	}

	var sessions []*models.UserSession
	err := r.db.WithContext(ctx).
		Where("ip_address = ? AND created_at >= ?", ipAddress, since).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// GetSessionsByDeviceFingerprint retrieves sessions by device fingerprint
func (r *userSessionRepository) GetSessionsByDeviceFingerprint(ctx context.Context, fingerprint string) ([]*models.UserSession, error) {
	if fingerprint == "" {
		return nil, errors.New("device fingerprint is required")
	}

	var sessions []*models.UserSession
	err := r.db.WithContext(ctx).
		Where("device_fingerprint = ?", fingerprint).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}