package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"auth-service/internal/models"
)

var (
	ErrSessionNotFound = errors.New("session not found")
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
	err := r.db.WithContext(ctx).Where("user_id = ? AND expires_at > ?", userID, time.Now()).Order("created_at DESC").Find(&sessions).Error
	return sessions, err
}

// Update updates a session
func (r *userSessionRepository) Update(ctx context.Context, session *models.UserSession) error {
	if session == nil || session.ID == "" {
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