package repository

import (
	"context"

	"auth-service/internal/models"

	"gorm.io/gorm"
)

// organizationMembershipRepository implements OrganizationMembershipRepository
type organizationMembershipRepository struct {
	db *gorm.DB
}

// NewOrganizationMembershipRepository creates a new organization membership repository
func NewOrganizationMembershipRepository(db *gorm.DB) OrganizationMembershipRepository {
	return &organizationMembershipRepository{db: db}
}

// Create creates a new organization membership
func (r *organizationMembershipRepository) Create(ctx context.Context, membership *models.OrganizationMembership) error {
	return r.db.WithContext(ctx).Create(membership).Error
}

// GetByID gets a membership by ID
func (r *organizationMembershipRepository) GetByID(ctx context.Context, id string) (*models.OrganizationMembership, error) {
	var membership models.OrganizationMembership
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&membership).Error
	return &membership, err
}

// GetByOrganizationAndUser gets membership by organization and user
func (r *organizationMembershipRepository) GetByOrganizationAndUser(ctx context.Context, orgID, userID string) (*models.OrganizationMembership, error) {
	var membership models.OrganizationMembership
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		First(&membership).Error
	return &membership, err
}

// GetByOrganizationAndEmail gets membership by organization and email
func (r *organizationMembershipRepository) GetByOrganizationAndEmail(ctx context.Context, orgID, email string) (*models.OrganizationMembership, error) {
	var membership models.OrganizationMembership
	err := r.db.WithContext(ctx).
		Joins("JOIN users u ON u.id = organization_memberships.user_id").
		Where("organization_memberships.organization_id = ? AND u.email = ?", orgID, email).
		First(&membership).Error
	return &membership, err
}

// GetByOrganization gets all memberships for an organization
func (r *organizationMembershipRepository) GetByOrganization(ctx context.Context, orgID string) ([]*models.OrganizationMembership, error) {
	var memberships []*models.OrganizationMembership
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Find(&memberships).Error
	return memberships, err
}

// GetByUser gets all memberships for a user
func (r *organizationMembershipRepository) GetByUser(ctx context.Context, userID string) ([]*models.OrganizationMembership, error) {
	var memberships []*models.OrganizationMembership
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&memberships).Error
	return memberships, err
}

// Update updates a membership
func (r *organizationMembershipRepository) Update(ctx context.Context, membership *models.OrganizationMembership) error {
	return r.db.WithContext(ctx).Save(membership).Error
}

// Delete deletes a membership
func (r *organizationMembershipRepository) Delete(ctx context.Context, orgID, userID string) error {
	return r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Delete(&models.OrganizationMembership{}).Error
}

// CountByOrganization counts memberships for an organization
func (r *organizationMembershipRepository) CountByOrganization(ctx context.Context, orgID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.OrganizationMembership{}).
		Where("organization_id = ?", orgID).
		Count(&count).Error
	return count, err
}
