package repository

import (
	"context"

	"auth-service/internal/models"

	"gorm.io/gorm"
)

// organizationRepository implements OrganizationRepository
type organizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &organizationRepository{db: db}
}

// Create creates a new organization
func (r *organizationRepository) Create(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

// GetByID gets an organization by ID
func (r *organizationRepository) GetByID(ctx context.Context, id string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&org).Error
	return &org, err
}

// GetBySlug gets an organization by slug
func (r *organizationRepository) GetBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&org).Error
	return &org, err
}

// GetByUserID gets organizations for a user
func (r *organizationRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Organization, error) {
	var orgs []*models.Organization
	err := r.db.WithContext(ctx).
		Joins("JOIN organization_memberships om ON om.organization_id = organizations.id").
		Where("om.user_id = ? AND om.status = ?", userID, models.MembershipStatusActive).
		Find(&orgs).Error
	return orgs, err
}

// Update updates an organization
func (r *organizationRepository) Update(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

// Delete deletes an organization
func (r *organizationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Organization{}, "id = ?", id).Error
}

// List lists organizations with pagination
func (r *organizationRepository) List(ctx context.Context, limit, offset int) ([]*models.Organization, error) {
	var orgs []*models.Organization
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&orgs).Error
	return orgs, err
}

// Count counts total organizations
func (r *organizationRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Organization{}).Count(&count).Error
	return count, err
}
