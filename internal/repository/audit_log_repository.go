package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"auth-service/internal/models"
)

// AuditLogRepository handles audit log database operations
type AuditLogRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.AuditLog, error)
	FindByOrganizationID(ctx context.Context, orgID uuid.UUID, limit int, offset int) ([]*models.AuditLog, error)
	FindByAction(ctx context.Context, action string, limit int, offset int) ([]*models.AuditLog, error)
	FindByRequestID(ctx context.Context, requestID string) ([]*models.AuditLog, error)
	FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit int, offset int) ([]*models.AuditLog, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	CountByOrganizationID(ctx context.Context, orgID uuid.UUID) (int64, error)
	DeleteOlderThan(ctx context.Context, date time.Time) (int64, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

// Create creates a new audit log entry
func (r *auditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindByID finds an audit log by ID
func (r *auditLogRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) {
	var log models.AuditLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// FindByUserID finds audit logs for a specific user
func (r *auditLogRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// FindByOrganizationID finds audit logs for a specific organization
func (r *auditLogRepository) FindByOrganizationID(ctx context.Context, orgID uuid.UUID, limit int, offset int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// FindByAction finds audit logs for a specific action
func (r *auditLogRepository) FindByAction(ctx context.Context, action string, limit int, offset int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	err := r.db.WithContext(ctx).
		Where("action = ?", action).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// FindByRequestID finds all audit logs for a specific request ID
func (r *auditLogRepository) FindByRequestID(ctx context.Context, requestID string) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	err := r.db.WithContext(ctx).
		Where("request_id = ?", requestID).
		Order("timestamp DESC").
		Find(&logs).Error
	return logs, err
}

// FindByDateRange finds audit logs within a date range
func (r *auditLogRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit int, offset int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	err := r.db.WithContext(ctx).
		Where("timestamp >= ? AND timestamp <= ?", startDate, endDate).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// CountByUserID counts audit logs for a specific user
func (r *auditLogRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountByOrganizationID counts audit logs for a specific organization
func (r *auditLogRepository) CountByOrganizationID(ctx context.Context, orgID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("organization_id = ?", orgID).
		Count(&count).Error
	return count, err
}

// DeleteOlderThan deletes audit logs older than the specified date (for retention policy)
func (r *auditLogRepository) DeleteOlderThan(ctx context.Context, date time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", date).
		Delete(&models.AuditLog{})
	return result.RowsAffected, result.Error
}
