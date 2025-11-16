package repository

import (
	"context"
	"fmt"

	"auth-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type organizationMembershipRepository struct {
	db *gorm.DB
}

func NewOrganizationMembershipRepository(db *gorm.DB) OrganizationMembershipRepository {
	return &organizationMembershipRepository{db: db}
}

/* ------------------------
      CRUD
------------------------ */

func (r *organizationMembershipRepository) Create(ctx context.Context, m *models.OrganizationMembership) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *organizationMembershipRepository) Update(ctx context.Context, m *models.OrganizationMembership) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *organizationMembershipRepository) Delete(ctx context.Context, orgID, userID string) error {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization ID: %w", err)
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	return r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgUUID, userUUID).
		Delete(&models.OrganizationMembership{}).Error
}

/* ------------------------
   GET BY ID
------------------------ */

func (r *organizationMembershipRepository) GetByID(ctx context.Context, id string) (*models.OrganizationMembership, error) {
	mID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid membership ID: %w", err)
	}

	var m models.OrganizationMembership
	err = r.db.WithContext(ctx).
		Preload("User").
		Preload("Organization").
		Preload("Role").
		Where("id = ?", mID).
		First(&m).Error

	if err != nil {
		return nil, err
	}

	// Load permissions with organization-aware filtering AFTER we have the membership
	if m.RoleID != uuid.Nil {
		err = r.db.WithContext(ctx).
			Preload("Permissions", "is_system = true OR organization_id = ? OR organization_id IS NULL", m.OrganizationID).
			First(&m.Role, "id = ?", m.RoleID).Error
		if err != nil {
			return nil, err
		}
	}

	return &m, nil
}

/* ------------------------
   GET BY ORG + USER
------------------------ */

func (r *organizationMembershipRepository) GetByOrganizationAndUser(ctx context.Context, orgID, userID string) (*models.OrganizationMembership, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var m models.OrganizationMembership
	err = r.db.WithContext(ctx).
		Preload("Role").
		Preload("Role.Permissions", "is_system = true OR organization_id = ? OR organization_id IS NULL", orgUUID).
		Where("organization_id = ? AND user_id = ?", orgUUID, userUUID).
		First(&m).Error

	return &m, err
}

/* ------------------------
   GET MEMBERS BY ORG
------------------------ */

func (r *organizationMembershipRepository) GetByOrganization(ctx context.Context, orgID string) ([]*models.OrganizationMembership, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var list []*models.OrganizationMembership
	err = r.db.WithContext(ctx).
		Preload("User").
		Preload("Role").
		Preload("Role.Permissions", "is_system = true OR organization_id = ? OR organization_id IS NULL", orgUUID).
		Where("organization_id = ?", orgUUID).
		Find(&list).Error

	return list, err
}

/* ------------------------
   GET ORGS BY USER
------------------------ */

func (r *organizationMembershipRepository) GetByUser(ctx context.Context, userID string) ([]*models.OrganizationMembership, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var list []*models.OrganizationMembership
	err = r.db.WithContext(ctx).
		Preload("Organization").
		Preload("Role").
		Find(&list, "user_id = ?", userUUID).Error

	if err != nil {
		return nil, err
	}

	// Now load permissions per role
	for _, m := range list {
		if m.RoleID == uuid.Nil {
			continue
		}
		err = r.db.WithContext(ctx).
			Preload("Permissions", "is_system = true OR organization_id = ? OR organization_id IS NULL", m.OrganizationID).
			First(&m.Role, "id = ?", m.RoleID).Error

		if err != nil {
			return nil, err
		}
	}

	return list, nil
}

/* ------------------------
   GET BY ORG + EMAIL
------------------------ */

func (r *organizationMembershipRepository) GetByOrganizationAndEmail(ctx context.Context, orgID, email string) (*models.OrganizationMembership, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var m models.OrganizationMembership
	err = r.db.WithContext(ctx).
		Joins("JOIN users ON users.id = organization_memberships.user_id").
		Where("organization_memberships.organization_id = ? AND users.email = ?", orgUUID, email).
		Preload("Role").
		Preload("Role.Permissions", "is_system = true OR organization_id = ? OR organization_id IS NULL", orgUUID).
		First(&m).Error

	return &m, err
}

/* ------------------------
   COUNT
------------------------ */

func (r *organizationMembershipRepository) CountByOrganization(ctx context.Context, orgID string) (int64, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return 0, fmt.Errorf("invalid organization ID: %w", err)
	}

	var count int64
	err = r.db.WithContext(ctx).
		Model(&models.OrganizationMembership{}).
		Where("organization_id = ?", orgUUID).
		Count(&count).Error

	return count, err
}
