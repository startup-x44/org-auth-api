package repository

import (
	"context"
	"time"

	"auth-service/internal/models"

	"gorm.io/gorm"
)

// organizationInvitationRepository implements OrganizationInvitationRepository
type organizationInvitationRepository struct {
	db *gorm.DB
}

// NewOrganizationInvitationRepository creates a new organization invitation repository
func NewOrganizationInvitationRepository(db *gorm.DB) OrganizationInvitationRepository {
	return &organizationInvitationRepository{db: db}
}

// Create creates a new organization invitation
func (r *organizationInvitationRepository) Create(ctx context.Context, invitation *models.OrganizationInvitation) error {
	return r.db.WithContext(ctx).Create(invitation).Error
}

// GetByID gets an invitation by ID
func (r *organizationInvitationRepository) GetByID(ctx context.Context, id string) (*models.OrganizationInvitation, error) {
	var invitation models.OrganizationInvitation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&invitation).Error
	return &invitation, err
}

// GetByToken gets an invitation by token hash
func (r *organizationInvitationRepository) GetByToken(ctx context.Context, tokenHash string) (*models.OrganizationInvitation, error) {
	var invitation models.OrganizationInvitation
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&invitation).Error
	return &invitation, err
}

// GetByOrganizationAndEmail gets invitation by organization and email
func (r *organizationInvitationRepository) GetByOrganizationAndEmail(ctx context.Context, orgID, email string) (*models.OrganizationInvitation, error) {
	var invitation models.OrganizationInvitation
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND email = ?", orgID, email).
		First(&invitation).Error
	return &invitation, err
}

// GetPendingByOrganization gets pending invitations for an organization
func (r *organizationInvitationRepository) GetPendingByOrganization(ctx context.Context, orgID string) ([]*models.OrganizationInvitation, error) {
	var invitations []*models.OrganizationInvitation
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND status = ?", orgID, models.InvitationStatusPending).
		Find(&invitations).Error
	return invitations, err
}

// Update updates an invitation
func (r *organizationInvitationRepository) Update(ctx context.Context, invitation *models.OrganizationInvitation) error {
	return r.db.WithContext(ctx).Save(invitation).Error
}

// Delete deletes an invitation
func (r *organizationInvitationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.OrganizationInvitation{}, "id = ?", id).Error
}

// CleanupExpired removes expired invitations
func (r *organizationInvitationRepository) CleanupExpired(ctx context.Context, maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)
	return r.db.WithContext(ctx).
		Where("expires_at < ? AND status = ?", cutoff, models.InvitationStatusPending).
		Delete(&models.OrganizationInvitation{}).Error
}
