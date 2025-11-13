package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"auth-service/internal/models"
)

var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantAlreadyExists = errors.New("tenant already exists")
	ErrInvalidTenantID     = errors.New("invalid tenant ID")
)

// tenantRepository implements TenantRepository interface
type tenantRepository struct {
	db *gorm.DB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

// Create creates a new tenant
func (r *tenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	if tenant == nil {
		return errors.New("tenant cannot be nil")
	}

	// Check if tenant with same domain already exists
	var existingTenant models.Tenant
	result := r.db.WithContext(ctx).Where("domain = ?", tenant.Domain).First(&existingTenant)
	if result.Error == nil {
		return ErrTenantAlreadyExists
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	return r.db.WithContext(ctx).Create(tenant).Error
}

// GetByID retrieves a tenant by ID
func (r *tenantRepository) GetByID(ctx context.Context, id string) (*models.Tenant, error) {
	if id == "" {
		return nil, ErrInvalidTenantID
	}

	var tenant models.Tenant
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTenantNotFound
	}
	return &tenant, err
}

// GetByDomain retrieves a tenant by domain
func (r *tenantRepository) GetByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	if domain == "" {
		return nil, errors.New("domain is required")
	}

	var tenant models.Tenant
	err := r.db.WithContext(ctx).Where("domain = ?", domain).First(&tenant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTenantNotFound
	}
	return &tenant, err
}

// Update updates a tenant
func (r *tenantRepository) Update(ctx context.Context, tenant *models.Tenant) error {
	if tenant == nil || tenant.ID == "" {
		return errors.New("tenant and tenant ID are required")
	}

	result := r.db.WithContext(ctx).Model(tenant).Updates(tenant)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTenantNotFound
	}
	return nil
}

// Delete deletes a tenant
func (r *tenantRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidTenantID
	}

	result := r.db.WithContext(ctx).Delete(&models.Tenant{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTenantNotFound
	}
	return nil
}

// List retrieves tenants with pagination
func (r *tenantRepository) List(ctx context.Context, limit, offset int) ([]*models.Tenant, error) {
	var tenants []*models.Tenant
	query := r.db.WithContext(ctx)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("created_at DESC").Find(&tenants).Error
	return tenants, err
}

// Count counts all tenants
func (r *tenantRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Tenant{}).Count(&count).Error
	return count, err
}