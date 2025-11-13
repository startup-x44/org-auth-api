package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/logger"
)

// SessionService defines the interface for advanced session management
type SessionService interface {
	CreateSession(ctx context.Context, userID, tenantID string, ipAddress, userAgent string) (*models.UserSession, error)
	ValidateSession(ctx context.Context, token string) (*models.UserSession, error)
	UpdateSessionActivity(ctx context.Context, sessionID string) error
	RevokeSession(ctx context.Context, sessionID, reason string) error
	RevokeUserSessions(ctx context.Context, userID, reason string) error
	GetUserSessions(ctx context.Context, userID string) ([]*models.UserSession, error)
	GetActiveSessionCount(ctx context.Context, userID string) (int64, error)
	EnforceSessionLimits(ctx context.Context, userID string, maxSessions int) error
	DetectSuspiciousActivity(ctx context.Context, session *models.UserSession) (bool, string)
	CleanupExpiredSessions(ctx context.Context) error
	CleanupInactiveSessions(ctx context.Context, maxInactive time.Duration) error
	GetSessionsByIPAddress(ctx context.Context, ipAddress string) ([]*models.UserSession, error)
	GetSessionsByDevice(ctx context.Context, fingerprint string) ([]*models.UserSession, error)
}

// SessionConfig holds configuration for session management
type SessionConfig struct {
	MaxSessionsPerUser    int           `json:"max_sessions_per_user"`
	SessionTimeout        time.Duration `json:"session_timeout"`
	MaxInactiveTime       time.Duration `json:"max_inactive_time"`
	EnableGeoTracking    bool          `json:"enable_geo_tracking"`
	EnableDeviceTracking bool          `json:"enable_device_tracking"`
	SuspiciousActivityThreshold int     `json:"suspicious_activity_threshold"`
}

// sessionService implements SessionService interface
type sessionService struct {
	repo        repository.Repository
	config      *SessionConfig
	auditLogger *logger.AuditLogger
}

// NewSessionService creates a new session service
func NewSessionService(repo repository.Repository, config *SessionConfig) SessionService {
	return &sessionService{
		repo:        repo,
		config:      config,
		auditLogger: logger.NewAuditLogger(),
	}
}

// CreateSession creates a new user session with device fingerprinting
func (s *sessionService) CreateSession(ctx context.Context, userID, tenantID string, ipAddress, userAgent string) (*models.UserSession, error) {
	// Generate session token
	sessionToken := generateSecureToken()

	// Create device fingerprint
	deviceFingerprint := s.generateDeviceFingerprint(ipAddress, userAgent)

	// Get location if enabled
	var location string
	if s.config.EnableGeoTracking {
		location = s.getLocationFromIP(ipAddress)
	}

	session := &models.UserSession{
		UserID:             uuid.MustParse(userID),
		TenantID:           uuid.MustParse(tenantID),
		TokenHash:          sessionToken,
		IPAddress:          ipAddress,
		UserAgent:          userAgent,
		DeviceFingerprint:  deviceFingerprint,
		Location:           location,
		IsActive:           true,
		LastActivity:       time.Now(),
		ExpiresAt:          time.Now().Add(s.config.SessionTimeout),
	}

	if err := s.repo.UserSession().Create(ctx, session); err != nil {
		s.auditLogger.LogSecurityEvent("session_creation_failed", "", session.IPAddress, false, err, "Failed to create user session")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.auditLogger.LogSecurityEvent("session_created", "", session.IPAddress, true, nil, "New user session created")
	return session, nil
}

// ValidateSession validates a session token and updates activity
func (s *sessionService) ValidateSession(ctx context.Context, token string) (*models.UserSession, error) {
	session, err := s.repo.UserSession().GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// Check if session is active
	if !session.IsActive {
		return nil, fmt.Errorf("session is inactive: %s", session.RevokedReason)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Auto-revoke expired session
		s.repo.UserSession().Revoke(ctx, session.ID.String(), "expired")
		return nil, fmt.Errorf("session expired")
	}

	// Update activity timestamp
	if err := s.repo.UserSession().UpdateActivity(ctx, session.ID.String()); err != nil {
		// Log error but don't fail validation
		fmt.Printf("Failed to update session activity: %v\n", err)
	}

	return session, nil
}

// UpdateSessionActivity updates the last activity timestamp for a session
func (s *sessionService) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	return s.repo.UserSession().UpdateActivity(ctx, sessionID)
}

// RevokeSession revokes a specific session
func (s *sessionService) RevokeSession(ctx context.Context, sessionID, reason string) error {
	session, err := s.repo.UserSession().GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if err := s.repo.UserSession().Revoke(ctx, sessionID, reason); err != nil {
		s.auditLogger.LogSecurityEvent("session_revocation_failed", "", session.IPAddress, false, err, fmt.Sprintf("Failed to revoke session: %s", sessionID))
		return err
	}

	s.auditLogger.LogSecurityEvent("session_revoked", "", session.IPAddress, true, nil, fmt.Sprintf("Session revoked: %s", reason))
	return nil
}

// RevokeUserSessions revokes all sessions for a user
func (s *sessionService) RevokeUserSessions(ctx context.Context, userID, reason string) error {
	sessions, err := s.repo.UserSession().GetActiveByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	if err := s.repo.UserSession().RevokeByUserID(ctx, userID, reason); err != nil {
		s.auditLogger.LogSecurityEvent("bulk_session_revocation_failed", "", "", false, err, fmt.Sprintf("Failed to revoke sessions for user: %s", userID))
		return err
	}

	s.auditLogger.LogSecurityEvent("bulk_session_revoked", "", "", true, nil, fmt.Sprintf("Revoked %d sessions: %s", len(sessions), reason))
	return nil
}

// GetUserSessions retrieves all sessions for a user
func (s *sessionService) GetUserSessions(ctx context.Context, userID string) ([]*models.UserSession, error) {
	return s.repo.UserSession().GetByUserID(ctx, userID)
}

// GetActiveSessionCount returns the count of active sessions for a user
func (s *sessionService) GetActiveSessionCount(ctx context.Context, userID string) (int64, error) {
	return s.repo.UserSession().GetActiveCountByUserID(ctx, userID)
}

// EnforceSessionLimits enforces maximum concurrent sessions per user
func (s *sessionService) EnforceSessionLimits(ctx context.Context, userID string, maxSessions int) error {
	activeCount, err := s.GetActiveSessionCount(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get active session count: %w", err)
	}

	if activeCount >= int64(maxSessions) {
		// Get oldest sessions to revoke
		sessions, err := s.repo.UserSession().GetActiveByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user sessions: %w", err)
		}

		// Revoke oldest sessions to make room for new one
		sessionsToRevoke := int(activeCount) - maxSessions + 1
		for i := 0; i < sessionsToRevoke && i < len(sessions); i++ {
			if err := s.RevokeSession(ctx, sessions[len(sessions)-1-i].ID.String(), "session_limit_exceeded"); err != nil {
				fmt.Printf("Failed to revoke old session: %v\n", err)
			}
		}
	}

	return nil
}

// DetectSuspiciousActivity detects suspicious session activity
func (s *sessionService) DetectSuspiciousActivity(ctx context.Context, session *models.UserSession) (bool, string) {
	reasons := []string{}

	// Check for rapid location changes
	if s.config.EnableGeoTracking && session.Location != "" {
		recentSessions, err := s.repo.UserSession().GetSessionsByDeviceFingerprint(ctx, session.DeviceFingerprint)
		if err == nil && len(recentSessions) > 0 {
			for _, s := range recentSessions {
				if s.Location != "" && s.Location != session.Location {
					reasons = append(reasons, "location_change")
					break
				}
			}
		}
	}

	// Check for unusual login times (basic check)
	now := time.Now()
	hour := now.Hour()
	if hour < 6 || hour > 22 { // Outside normal business hours
		reasons = append(reasons, "unusual_time")
	}

	// Check for multiple failed attempts from same IP recently
	since := time.Now().Add(-time.Hour)
	ipSessions, err := s.repo.UserSession().GetSessionsByIPAddress(ctx, session.IPAddress, since)
	if err == nil && len(ipSessions) > s.config.SuspiciousActivityThreshold {
		reasons = append(reasons, "high_frequency_logins")
	}

	if len(reasons) > 0 {
		return true, strings.Join(reasons, ",")
	}

	return false, ""
}

// CleanupExpiredSessions removes expired sessions
func (s *sessionService) CleanupExpiredSessions(ctx context.Context) error {
	if err := s.repo.UserSession().RevokeExpired(ctx); err != nil {
		return fmt.Errorf("failed to revoke expired sessions: %w", err)
	}

	// Actually delete old expired sessions after a grace period
	gracePeriod := 7 * 24 * time.Hour // 7 days
	if err := s.repo.UserSession().CleanupExpired(ctx, gracePeriod); err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	return nil
}

// CleanupInactiveSessions removes inactive sessions
func (s *sessionService) CleanupInactiveSessions(ctx context.Context, maxInactive time.Duration) error {
	return s.repo.UserSession().RevokeInactive(ctx, maxInactive)
}

// GetSessionsByIPAddress retrieves sessions by IP address
func (s *sessionService) GetSessionsByIPAddress(ctx context.Context, ipAddress string) ([]*models.UserSession, error) {
	since := time.Now().Add(-24 * time.Hour) // Last 24 hours
	return s.repo.UserSession().GetSessionsByIPAddress(ctx, ipAddress, since)
}

// GetSessionsByDevice retrieves sessions by device fingerprint
func (s *sessionService) GetSessionsByDevice(ctx context.Context, fingerprint string) ([]*models.UserSession, error) {
	return s.repo.UserSession().GetSessionsByDeviceFingerprint(ctx, fingerprint)
}

// generateDeviceFingerprint creates a device fingerprint from IP and user agent
func (s *sessionService) generateDeviceFingerprint(ipAddress, userAgent string) string {
	if !s.config.EnableDeviceTracking {
		return ""
	}

	// Create a simple fingerprint from IP and user agent
	fingerprint := fmt.Sprintf("%s|%s", ipAddress, userAgent)
	hash := sha256.Sum256([]byte(fingerprint))
	return fmt.Sprintf("%x", hash)[:16] // First 16 chars of hash
}

// getLocationFromIP attempts to get location from IP address
func (s *sessionService) getLocationFromIP(ipAddress string) string {
	if !s.config.EnableGeoTracking {
		return ""
	}

	// Parse IP address
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return ""
	}

	// This is a placeholder - in a real implementation, you would use a GeoIP service
	// For now, return a basic classification
	if ip.IsPrivate() {
		return "private_network"
	}

	// Basic geographic hints (very simplified)
	if strings.HasPrefix(ipAddress, "192.168.") || strings.HasPrefix(ipAddress, "10.") {
		return "local_network"
	}

	return "unknown"
}

// generateSecureToken generates a cryptographically secure session token
func generateSecureToken() string {
	// This should use crypto/rand in a real implementation
	// For now, return a UUID-based token
	return uuid.New().String() + "-" + uuid.New().String()
}