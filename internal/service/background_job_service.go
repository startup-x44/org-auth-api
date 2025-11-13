package service

import (
	"context"
	"sync"
	"time"

	"auth-service/internal/repository"
	"auth-service/pkg/logger"
)

// BackgroundJobService defines the interface for background job operations
type BackgroundJobService interface {
	Start()
	Stop()
	CleanupExpiredSessions(ctx context.Context) error
	CleanupInactiveSessions(ctx context.Context) error
	CleanupExpiredTokens(ctx context.Context) error
	CleanupFailedAttempts(ctx context.Context) error
}

// BackgroundJobConfig holds configuration for background jobs
type BackgroundJobConfig struct {
	SessionCleanupInterval    time.Duration `json:"session_cleanup_interval"`
	TokenCleanupInterval      time.Duration `json:"token_cleanup_interval"`
	FailedAttemptCleanupInterval time.Duration `json:"failed_attempt_cleanup_interval"`
	MaxInactiveSessionTime    time.Duration `json:"max_inactive_session_time"`
	MaxFailedAttemptAge       time.Duration `json:"max_failed_attempt_age"`
}

// backgroundJobService implements BackgroundJobService interface
type backgroundJobService struct {
	repo        repository.Repository
	sessionSvc  SessionService
	config      *BackgroundJobConfig
	logger      *logger.AuditLogger
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// NewBackgroundJobService creates a new background job service
func NewBackgroundJobService(repo repository.Repository, sessionSvc SessionService, config *BackgroundJobConfig) BackgroundJobService {
	return &backgroundJobService{
		repo:       repo,
		sessionSvc: sessionSvc,
		config:     config,
		logger:     logger.NewAuditLogger(),
		stopChan:   make(chan struct{}),
	}
}

// Start starts the background job service
func (s *backgroundJobService) Start() {
	s.logger.LogSystemEvent("system", "background_jobs_started", "service", "", "", "", true, nil, "Background job service started")

	// Start session cleanup job
	s.wg.Add(1)
	go s.sessionCleanupJob()

	// Start token cleanup job
	s.wg.Add(1)
	go s.tokenCleanupJob()

	// Start failed attempts cleanup job
	s.wg.Add(1)
	go s.failedAttemptsCleanupJob()
}

// Stop stops the background job service
func (s *backgroundJobService) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	s.logger.LogSystemEvent("system", "background_jobs_stopped", "service", "", "", "", true, nil, "Background job service stopped")
}

// sessionCleanupJob runs periodic session cleanup
func (s *backgroundJobService) sessionCleanupJob() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.SessionCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			ctx := context.Background()
			if err := s.CleanupExpiredSessions(ctx); err != nil {
				s.logger.LogSystemEvent("system", "session_cleanup_failed", "cleanup", "", "", "", false, err, "Failed to cleanup expired sessions")
			}
			if err := s.CleanupInactiveSessions(ctx); err != nil {
				s.logger.LogSystemEvent("system", "inactive_session_cleanup_failed", "cleanup", "", "", "", false, err, "Failed to cleanup inactive sessions")
			}
		}
	}
}

// tokenCleanupJob runs periodic token cleanup
func (s *backgroundJobService) tokenCleanupJob() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.TokenCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			ctx := context.Background()
			if err := s.CleanupExpiredTokens(ctx); err != nil {
				s.logger.LogSystemEvent("system", "token_cleanup_failed", "cleanup", "", "", "", false, err, "Failed to cleanup expired tokens")
			}
		}
	}
}

// failedAttemptsCleanupJob runs periodic failed attempts cleanup
func (s *backgroundJobService) failedAttemptsCleanupJob() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.FailedAttemptCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			ctx := context.Background()
			if err := s.CleanupFailedAttempts(ctx); err != nil {
				s.logger.LogSystemEvent("system", "failed_attempts_cleanup_failed", "cleanup", "", "", "", false, err, "Failed to cleanup failed attempts")
			}
		}
	}
}

// CleanupExpiredSessions cleans up expired sessions
func (s *backgroundJobService) CleanupExpiredSessions(ctx context.Context) error {
	if s.sessionSvc != nil {
		return s.sessionSvc.CleanupExpiredSessions(ctx)
	}

	// Fallback to repository method
	return s.repo.UserSession().DeleteExpired(ctx)
}

// CleanupInactiveSessions cleans up inactive sessions
func (s *backgroundJobService) CleanupInactiveSessions(ctx context.Context) error {
	if s.sessionSvc != nil {
		return s.sessionSvc.CleanupInactiveSessions(ctx, s.config.MaxInactiveSessionTime)
	}

	// Fallback to repository method
	return s.repo.UserSession().RevokeInactive(ctx, s.config.MaxInactiveSessionTime)
}

// CleanupExpiredTokens cleans up expired refresh tokens and password resets
func (s *backgroundJobService) CleanupExpiredTokens(ctx context.Context) error {
	// Cleanup expired refresh tokens
	if err := s.repo.RefreshToken().DeleteExpired(ctx); err != nil {
		return err
	}

	// Cleanup expired password resets
	if err := s.repo.PasswordReset().DeleteExpired(ctx); err != nil {
		return err
	}

	return nil
}

// CleanupFailedAttempts cleans up old failed login attempts
func (s *backgroundJobService) CleanupFailedAttempts(ctx context.Context) error {
	return s.repo.FailedLoginAttempt().DeleteExpired(ctx, s.config.MaxFailedAttemptAge)
}