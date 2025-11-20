package service

import (
	"context"
	"runtime"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/logger"
)

// AuditService handles audit logging operations
type AuditService interface {
	// Core logging methods
	LogAuth(ctx context.Context, action string, userID *uuid.UUID, success bool, details map[string]interface{}, err error)
	LogRole(ctx context.Context, action string, userID uuid.UUID, roleID *uuid.UUID, orgID *uuid.UUID, success bool, details map[string]interface{}, err error)
	LogPermission(ctx context.Context, action string, userID uuid.UUID, permissionID *uuid.UUID, orgID *uuid.UUID, success bool, details map[string]interface{}, err error)
	LogOrganization(ctx context.Context, action string, userID uuid.UUID, orgID *uuid.UUID, success bool, details map[string]interface{}, err error)
	LogUser(ctx context.Context, action string, actorUserID uuid.UUID, targetUserID *uuid.UUID, orgID *uuid.UUID, success bool, details map[string]interface{}, err error)
	LogSession(ctx context.Context, action string, userID uuid.UUID, sessionID *uuid.UUID, success bool, details map[string]interface{}, err error)
	LogOAuth(ctx context.Context, action string, userID *uuid.UUID, clientID *uuid.UUID, success bool, details map[string]interface{}, err error)
	LogAPIKey(ctx context.Context, action string, userID uuid.UUID, apiKeyID *uuid.UUID, success bool, details map[string]interface{}, err error)

	// Query methods
	GetUserAuditLogs(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.AuditLog, error)
	GetOrganizationAuditLogs(ctx context.Context, orgID uuid.UUID, limit int, offset int) ([]*models.AuditLog, error)
	GetAuditLogsByAction(ctx context.Context, action string, limit int, offset int) ([]*models.AuditLog, error)
	GetAuditLogsByRequestID(ctx context.Context, requestID string) ([]*models.AuditLog, error)

	// Retention policy
	CleanupOldLogs(ctx context.Context, retentionDays int) (int64, error)
}

type auditService struct {
	repo repository.AuditLogRepository
	db   *gorm.DB
}

// NewAuditService creates a new audit service
func NewAuditService(db *gorm.DB) AuditService {
	return &auditService{
		repo: repository.NewAuditLogRepository(db),
		db:   db,
	}
}

// createAuditLog is a helper to create audit log entries
func (s *auditService) createAuditLog(ctx context.Context, action, resource string, userID, resourceID, orgID *uuid.UUID, success bool, details map[string]interface{}, err error) {
	// Extract context values
	requestID, _ := ctx.Value(logger.RequestIDKey).(string)
	ipAddress, _ := ctx.Value(logger.IPAddressKey).(string)
	userAgent, _ := ctx.Value(logger.UserAgentKey).(string)

	// Get calling method name
	method := ""
	if pc, _, _, ok := runtime.Caller(2); ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			method = fn.Name()
		}
	}

	// Create audit log entry
	auditLog := &models.AuditLog{
		Timestamp:      time.Now(),
		UserID:         userID,
		OrganizationID: orgID,
		Action:         action,
		Resource:       resource,
		ResourceID:     resourceID,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		RequestID:      requestID,
		Details:        details,
		Success:        success,
		Service:        "auth-service",
		Method:         method,
	}

	if err != nil {
		auditLog.Error = err.Error()
	}

	// Persist to database (fire and forget - don't block on audit logging)
	go func() {
		if createErr := s.repo.Create(context.Background(), auditLog); createErr != nil {
			// Log to structured logger if DB insert fails
			logger.Error(ctx).
				Str("action", action).
				Str("resource", resource).
				Bool("success", success).
				Err(createErr).
				Msg("Failed to persist audit log to database")
		}
	}()

	// Also log to structured logger for real-time visibility
	logEvent := logger.Info(ctx).
		Str("audit_action", action).
		Str("audit_resource", resource).
		Bool("audit_success", success)

	if userID != nil {
		logEvent = logEvent.Str("audit_user_id", userID.String())
	}
	if resourceID != nil {
		logEvent = logEvent.Str("audit_resource_id", resourceID.String())
	}
	if orgID != nil {
		logEvent = logEvent.Str("audit_org_id", orgID.String())
	}
	if len(details) > 0 {
		logEvent = logEvent.Interface("audit_details", details)
	}
	if err != nil {
		logEvent = logEvent.Err(err)
	}

	logEvent.Msg("Audit event")
}

// LogAuth logs authentication-related events
func (s *auditService) LogAuth(ctx context.Context, action string, userID *uuid.UUID, success bool, details map[string]interface{}, err error) {
	s.createAuditLog(ctx, action, models.ResourceAuth, userID, nil, nil, success, details, err)
}

// LogRole logs role-related events
func (s *auditService) LogRole(ctx context.Context, action string, userID uuid.UUID, roleID *uuid.UUID, orgID *uuid.UUID, success bool, details map[string]interface{}, err error) {
	s.createAuditLog(ctx, action, models.ResourceRole, &userID, roleID, orgID, success, details, err)
}

// LogPermission logs permission-related events
func (s *auditService) LogPermission(ctx context.Context, action string, userID uuid.UUID, permissionID *uuid.UUID, orgID *uuid.UUID, success bool, details map[string]interface{}, err error) {
	s.createAuditLog(ctx, action, models.ResourcePermission, &userID, permissionID, orgID, success, details, err)
}

// LogOrganization logs organization-related events
func (s *auditService) LogOrganization(ctx context.Context, action string, userID uuid.UUID, orgID *uuid.UUID, success bool, details map[string]interface{}, err error) {
	s.createAuditLog(ctx, action, models.ResourceOrganization, &userID, orgID, orgID, success, details, err)
}

// LogUser logs user management events
func (s *auditService) LogUser(ctx context.Context, action string, actorUserID uuid.UUID, targetUserID *uuid.UUID, orgID *uuid.UUID, success bool, details map[string]interface{}, err error) {
	s.createAuditLog(ctx, action, models.ResourceUser, &actorUserID, targetUserID, orgID, success, details, err)
}

// LogSession logs session-related events
func (s *auditService) LogSession(ctx context.Context, action string, userID uuid.UUID, sessionID *uuid.UUID, success bool, details map[string]interface{}, err error) {
	s.createAuditLog(ctx, action, models.ResourceSession, &userID, sessionID, nil, success, details, err)
}

// LogOAuth logs OAuth2-related events
func (s *auditService) LogOAuth(ctx context.Context, action string, userID *uuid.UUID, clientID *uuid.UUID, success bool, details map[string]interface{}, err error) {
	s.createAuditLog(ctx, action, models.ResourceOAuthClient, userID, clientID, nil, success, details, err)
}

// LogAPIKey logs API key-related events
func (s *auditService) LogAPIKey(ctx context.Context, action string, userID uuid.UUID, apiKeyID *uuid.UUID, success bool, details map[string]interface{}, err error) {
	s.createAuditLog(ctx, action, models.ResourceAPIKey, &userID, apiKeyID, nil, success, details, err)
}

// GetUserAuditLogs retrieves audit logs for a specific user
func (s *auditService) GetUserAuditLogs(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.AuditLog, error) {
	return s.repo.FindByUserID(ctx, userID, limit, offset)
}

// GetOrganizationAuditLogs retrieves audit logs for a specific organization
func (s *auditService) GetOrganizationAuditLogs(ctx context.Context, orgID uuid.UUID, limit int, offset int) ([]*models.AuditLog, error) {
	return s.repo.FindByOrganizationID(ctx, orgID, limit, offset)
}

// GetAuditLogsByAction retrieves audit logs for a specific action
func (s *auditService) GetAuditLogsByAction(ctx context.Context, action string, limit int, offset int) ([]*models.AuditLog, error) {
	return s.repo.FindByAction(ctx, action, limit, offset)
}

// GetAuditLogsByRequestID retrieves all audit logs for a specific request
func (s *auditService) GetAuditLogsByRequestID(ctx context.Context, requestID string) ([]*models.AuditLog, error) {
	return s.repo.FindByRequestID(ctx, requestID)
}

// CleanupOldLogs deletes audit logs older than the retention period
func (s *auditService) CleanupOldLogs(ctx context.Context, retentionDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	return s.repo.DeleteOlderThan(ctx, cutoffDate)
}
