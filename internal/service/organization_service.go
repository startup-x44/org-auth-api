package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/logger"
	"auth-service/pkg/validation"

	"github.com/google/uuid"
)

// OrganizationService defines the interface for organization business logic
type OrganizationService interface {
	CreateOrganization(ctx context.Context, req *CreateOrganizationRequest) (*OrganizationResponse, error)
	GetOrganization(ctx context.Context, orgID string) (*OrganizationResponse, error)
	GetOrganizationBySlug(ctx context.Context, slug string) (*OrganizationResponse, error)
	UpdateOrganization(ctx context.Context, orgID string, req *UpdateOrganizationRequest) (*OrganizationResponse, error)
	DeleteOrganization(ctx context.Context, orgID string) error
	ListUserOrganizations(ctx context.Context, userID string) ([]*OrganizationResponse, error)

	// Membership management
	InviteUser(ctx context.Context, req *InviteUserRequest) (*models.OrganizationInvitation, error)
	AcceptInvitation(ctx context.Context, token string, userID string) (*models.OrganizationMembership, error)
	GetMembership(ctx context.Context, orgID, userID string) (*models.OrganizationMembership, error)
	UpdateMembership(ctx context.Context, orgID, userID string, req *UpdateMembershipRequest) (*models.OrganizationMembership, error)
	RemoveMember(ctx context.Context, orgID, userID string) error
	ListMembers(ctx context.Context, orgID string) ([]*OrganizationMember, error)

	// Invitation management
	CancelInvitation(ctx context.Context, invitationID string) error
	ResendInvitation(ctx context.Context, invitationID string) (*models.OrganizationInvitation, error)
	ListPendingInvitations(ctx context.Context, orgID string) ([]*models.OrganizationInvitation, error)
}

// CreateOrganizationRequest represents organization creation request
type CreateOrganizationRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug,omitempty"` // Auto-generated if not provided
	Description *string `json:"description,omitempty"`
}

// OrganizationResponse represents organization response
type OrganizationResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	Status      string    `json:"status"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	MemberCount int       `json:"member_count"`
}

// UpdateOrganizationRequest represents organization update request
type UpdateOrganizationRequest struct {
	Name        string  `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// InviteUserRequest represents user invitation request
type InviteUserRequest struct {
	OrganizationID string `json:"organization_id"`
	Email          string `json:"email"`
	Role           string `json:"role"`
}

// UpdateMembershipRequest represents membership update request
type UpdateMembershipRequest struct {
	Role   string `json:"role,omitempty"`
	Status string `json:"status,omitempty"`
}

// OrganizationMember represents organization member with profile
type OrganizationMember struct {
	UserID         string     `json:"user_id"`
	Email          string     `json:"email"`
	FirstName      *string    `json:"first_name"`
	LastName       *string    `json:"last_name"`
	Role           string     `json:"role"`
	Status         string     `json:"status"`
	JoinedAt       *time.Time `json:"joined_at"`
	LastActivityAt *time.Time `json:"last_activity_at"`
}

// organizationService implements OrganizationService interface
type organizationService struct {
	repo        repository.Repository
	auditLogger *logger.AuditLogger
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(repo repository.Repository) OrganizationService {
	return &organizationService{
		repo:        repo,
		auditLogger: logger.NewAuditLogger(),
	}
}

// CreateOrganization creates a new organization
func (s *organizationService) CreateOrganization(ctx context.Context, req *CreateOrganizationRequest) (*OrganizationResponse, error) {
	// Get user ID from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user context required")
	}

	// Validate request
	if err := s.validateCreateOrganizationRequest(req); err != nil {
		return nil, err
	}

	// Generate slug if not provided
	if req.Slug == "" {
		req.Slug = s.generateSlug(req.Name)
	}

	// Check if slug is unique
	existing, _ := s.repo.Organization().GetBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, errors.New("organization slug already exists")
	}

	// Create organization
	org := &models.Organization{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		CreatedBy:   uuid.MustParse(userID),
	}

	if err := s.repo.Organization().Create(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Add creator as admin member
	membership := &models.OrganizationMembership{
		OrganizationID: org.ID,
		UserID:         uuid.MustParse(userID),
		Role:           models.OrganizationRoleAdmin,
		Status:         models.MembershipStatusActive,
		JoinedAt:       &org.CreatedAt,
	}

	if err := s.repo.OrganizationMembership().Create(ctx, membership); err != nil {
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	s.auditLogger.LogOrganizationAction(userID, "create_organization", org.ID.String(), "", "", true, nil, "Organization created")

	return s.convertToOrganizationResponse(ctx, org), nil
}

// GetOrganization gets an organization by ID
func (s *organizationService) GetOrganization(ctx context.Context, orgID string) (*OrganizationResponse, error) {
	org, err := s.repo.Organization().GetByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	return s.convertToOrganizationResponse(ctx, org), nil
}

// GetOrganizationBySlug gets an organization by slug
func (s *organizationService) GetOrganizationBySlug(ctx context.Context, slug string) (*OrganizationResponse, error) {
	org, err := s.repo.Organization().GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	return s.convertToOrganizationResponse(ctx, org), nil
}

// UpdateOrganization updates an organization
func (s *organizationService) UpdateOrganization(ctx context.Context, orgID string, req *UpdateOrganizationRequest) (*OrganizationResponse, error) {
	userID, _ := ctx.Value("user_id").(string)

	// Check admin permission
	if err := s.checkAdminPermission(ctx, orgID, userID); err != nil {
		return nil, err
	}

	org, err := s.repo.Organization().GetByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Update fields
	if req.Name != "" {
		org.Name = req.Name
	}
	if req.Description != nil {
		org.Description = req.Description
	}

	if err := s.repo.Organization().Update(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	s.auditLogger.LogOrganizationAction(userID, "update_organization", orgID, "", "", true, nil, "Organization updated")

	return s.convertToOrganizationResponse(ctx, org), nil
}

// DeleteOrganization deletes an organization
func (s *organizationService) DeleteOrganization(ctx context.Context, orgID string) error {
	userID, _ := ctx.Value("user_id").(string)

	// Check admin permission
	if err := s.checkAdminPermission(ctx, orgID, userID); err != nil {
		return err
	}

	// Check if organization has members
	memberCount, err := s.repo.OrganizationMembership().CountByOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to check members: %w", err)
	}

	if memberCount > 1 {
		return errors.New("cannot delete organization with multiple members")
	}

	if err := s.repo.Organization().Delete(ctx, orgID); err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	s.auditLogger.LogOrganizationAction(userID, "delete_organization", orgID, "", "", true, nil, "Organization deleted")

	return nil
}

// ListUserOrganizations lists organizations for a user
func (s *organizationService) ListUserOrganizations(ctx context.Context, userID string) ([]*OrganizationResponse, error) {
	orgs, err := s.repo.Organization().GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	responses := make([]*OrganizationResponse, len(orgs))
	for i, org := range orgs {
		responses[i] = s.convertToOrganizationResponse(ctx, org)
	}

	return responses, nil
}

// InviteUser invites a user to join an organization
func (s *organizationService) InviteUser(ctx context.Context, req *InviteUserRequest) (*models.OrganizationInvitation, error) {
	userID, _ := ctx.Value("user_id").(string)

	// Check admin permission
	if err := s.checkAdminPermission(ctx, req.OrganizationID, userID); err != nil {
		return nil, err
	}

	// Validate invitation
	if err := s.validateInviteUserRequest(req); err != nil {
		return nil, err
	}

	// Check if user is already a member
	_, err := s.repo.OrganizationMembership().GetByOrganizationAndEmail(ctx, req.OrganizationID, req.Email)
	if err == nil {
		return nil, errors.New("user is already a member of this organization")
	}

	// Check for existing pending invitation
	existingInvitation, err := s.repo.OrganizationInvitation().GetByOrganizationAndEmail(ctx, req.OrganizationID, req.Email)
	if err == nil && existingInvitation.Status == models.InvitationStatusPending {
		return nil, errors.New("pending invitation already exists for this email")
	}

	// Generate invitation token
	token := generateSecureToken()

	invitation := &models.OrganizationInvitation{
		OrganizationID: uuid.MustParse(req.OrganizationID),
		Email:          req.Email,
		TokenHash:      hashToken(token),
		Role:           req.Role,
		InvitedBy:      uuid.MustParse(userID),
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.repo.OrganizationInvitation().Create(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// TODO: Send invitation email
	s.auditLogger.LogOrganizationAction(userID, "invite_user", req.OrganizationID, "", "", true, nil, fmt.Sprintf("Invited %s as %s", req.Email, req.Role))

	return invitation, nil
}

// AcceptInvitation accepts an organization invitation
func (s *organizationService) AcceptInvitation(ctx context.Context, token string, userID string) (*models.OrganizationMembership, error) {
	// Get invitation by token
	invitation, err := s.repo.OrganizationInvitation().GetByToken(ctx, hashToken(token))
	if err != nil {
		return nil, errors.New("invalid invitation token")
	}

	// Check if invitation is still valid
	if invitation.Status != models.InvitationStatusPending {
		return nil, errors.New("invitation is no longer valid")
	}

	if time.Now().After(invitation.ExpiresAt) {
		return nil, errors.New("invitation has expired")
	}

	// Check if user email matches invitation
	user, err := s.repo.User().GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.Email != invitation.Email {
		return nil, errors.New("invitation email does not match user email")
	}

	// Create membership
	now := time.Now()
	membership := &models.OrganizationMembership{
		OrganizationID: invitation.OrganizationID,
		UserID:         uuid.MustParse(userID),
		Role:           invitation.Role,
		Status:         models.MembershipStatusActive,
		InvitedBy:      &invitation.InvitedBy,
		InvitedAt:      &invitation.CreatedAt,
		JoinedAt:       &now,
	}

	if err := s.repo.OrganizationMembership().Create(ctx, membership); err != nil {
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	// Update invitation status
	invitation.Status = models.InvitationStatusAccepted
	invitation.AcceptedAt = &now
	if err := s.repo.OrganizationInvitation().Update(ctx, invitation); err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to update invitation status: %v\n", err)
	}

	s.auditLogger.LogOrganizationAction(userID, "accept_invitation", invitation.OrganizationID.String(), "", "", true, nil, "Invitation accepted")

	return membership, nil
}

// GetMembership gets a user's membership in an organization
func (s *organizationService) GetMembership(ctx context.Context, orgID, userID string) (*models.OrganizationMembership, error) {
	membership, err := s.repo.OrganizationMembership().GetByOrganizationAndUser(ctx, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("membership not found: %w", err)
	}

	return membership, nil
}

// UpdateMembership updates a user's membership
func (s *organizationService) UpdateMembership(ctx context.Context, orgID, userID string, req *UpdateMembershipRequest) (*models.OrganizationMembership, error) {
	currentUserID, _ := ctx.Value("user_id").(string)

	// Check admin permission
	if err := s.checkAdminPermission(ctx, orgID, currentUserID); err != nil {
		return nil, err
	}

	membership, err := s.repo.OrganizationMembership().GetByOrganizationAndUser(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Role != "" {
		membership.Role = req.Role
	}
	if req.Status != "" {
		membership.Status = req.Status
	}

	if err := s.repo.OrganizationMembership().Update(ctx, membership); err != nil {
		return nil, fmt.Errorf("failed to update membership: %w", err)
	}

	s.auditLogger.LogOrganizationAction(currentUserID, "update_membership", orgID, "", "", true, nil, fmt.Sprintf("Updated membership for user %s", userID))

	return membership, nil
}

// RemoveMember removes a member from organization
func (s *organizationService) RemoveMember(ctx context.Context, orgID, userID string) error {
	currentUserID, _ := ctx.Value("user_id").(string)

	// Check admin permission
	if err := s.checkAdminPermission(ctx, orgID, currentUserID); err != nil {
		return err
	}

	if err := s.repo.OrganizationMembership().Delete(ctx, orgID, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	s.auditLogger.LogOrganizationAction(currentUserID, "remove_member", orgID, "", "", true, nil, fmt.Sprintf("Removed member %s", userID))

	return nil
}

// ListMembers lists organization members
func (s *organizationService) ListMembers(ctx context.Context, orgID string) ([]*OrganizationMember, error) {
	memberships, err := s.repo.OrganizationMembership().GetByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}

	members := make([]*OrganizationMember, len(memberships))
	for i, membership := range memberships {
		user, err := s.repo.User().GetByID(ctx, membership.UserID.String())
		if err != nil {
			continue // Skip if user not found
		}

		members[i] = &OrganizationMember{
			UserID:         user.ID.String(),
			Email:          user.Email,
			FirstName:      user.Firstname,
			LastName:       user.Lastname,
			Role:           membership.Role,
			Status:         membership.Status,
			JoinedAt:       membership.JoinedAt,
			LastActivityAt: membership.LastActivityAt,
		}
	}

	return members, nil
}

// CancelInvitation cancels a pending invitation
func (s *organizationService) CancelInvitation(ctx context.Context, invitationID string) error {
	userID, _ := ctx.Value("user_id").(string)

	invitation, err := s.repo.OrganizationInvitation().GetByID(ctx, invitationID)
	if err != nil {
		return err
	}

	// Check permission (only invited by user or org admin can cancel)
	if invitation.InvitedBy.String() != userID {
		if err := s.checkAdminPermission(ctx, invitation.OrganizationID.String(), userID); err != nil {
			return errors.New("not authorized to cancel this invitation")
		}
	}

	invitation.Status = models.InvitationStatusCancelled
	if err := s.repo.OrganizationInvitation().Update(ctx, invitation); err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}

	s.auditLogger.LogOrganizationAction(userID, "cancel_invitation", invitation.OrganizationID.String(), "", "", true, nil, "Invitation cancelled")

	return nil
}

// ResendInvitation resends a pending invitation
func (s *organizationService) ResendInvitation(ctx context.Context, invitationID string) (*models.OrganizationInvitation, error) {
	userID, _ := ctx.Value("user_id").(string)

	invitation, err := s.repo.OrganizationInvitation().GetByID(ctx, invitationID)
	if err != nil {
		return nil, err
	}

	// Check permission
	if invitation.InvitedBy.String() != userID {
		if err := s.checkAdminPermission(ctx, invitation.OrganizationID.String(), userID); err != nil {
			return nil, errors.New("not authorized to resend this invitation")
		}
	}

	// Extend expiry and update
	invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	if err := s.repo.OrganizationInvitation().Update(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	// TODO: Send invitation email again
	s.auditLogger.LogOrganizationAction(userID, "resend_invitation", invitation.OrganizationID.String(), "", "", true, nil, "Invitation resent")

	return invitation, nil
}

// ListPendingInvitations lists pending invitations for an organization
func (s *organizationService) ListPendingInvitations(ctx context.Context, orgID string) ([]*models.OrganizationInvitation, error) {
	userID, _ := ctx.Value("user_id").(string)

	// Check admin permission
	if err := s.checkAdminPermission(ctx, orgID, userID); err != nil {
		return nil, err
	}

	return s.repo.OrganizationInvitation().GetPendingByOrganization(ctx, orgID)
}

// Helper methods

func (s *organizationService) validateCreateOrganizationRequest(req *CreateOrganizationRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("organization name is required")
	}

	if len(req.Name) > 100 {
		return errors.New("organization name too long")
	}

	if req.Slug != "" {
		if !validation.IsValidSlug(req.Slug) {
			return errors.New("invalid slug format")
		}
	}

	return nil
}

func (s *organizationService) validateInviteUserRequest(req *InviteUserRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}

	if req.Role == "" {
		req.Role = models.OrganizationRoleStudent // Default role
	}

	validRoles := []string{
		models.OrganizationRoleAdmin,
		models.OrganizationRoleIssuer,
		models.OrganizationRoleRTO,
		models.OrganizationRoleStudent,
	}

	for _, role := range validRoles {
		if req.Role == role {
			return nil
		}
	}

	return errors.New("invalid role")
}

func (s *organizationService) generateSlug(name string) string {
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	slug = result.String()

	// Ensure uniqueness by appending number if needed
	baseSlug := slug
	counter := 1
	for {
		// Check if slug exists
		existing, _ := s.repo.Organization().GetBySlug(context.Background(), slug)
		if existing == nil {
			break
		}
		counter++
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
	}

	return slug
}

func (s *organizationService) convertToOrganizationResponse(ctx context.Context, org *models.Organization) *OrganizationResponse {
	memberCount, _ := s.repo.OrganizationMembership().CountByOrganization(ctx, org.ID.String())

	return &OrganizationResponse{
		ID:          org.ID.String(),
		Name:        org.Name,
		Slug:        org.Slug,
		Description: org.Description,
		Status:      org.Status,
		CreatedBy:   org.CreatedBy.String(),
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
		MemberCount: int(memberCount),
	}
}

func (s *organizationService) checkAdminPermission(ctx context.Context, orgID, userID string) error {
	isSuperadmin, _ := ctx.Value("is_superadmin").(bool)
	if isSuperadmin {
		return nil
	}

	membership, err := s.repo.OrganizationMembership().GetByOrganizationAndUser(ctx, orgID, userID)
	if err != nil {
		return errors.New("not a member of this organization")
	}

	if membership.Role != models.OrganizationRoleAdmin || membership.Status != models.MembershipStatusActive {
		return errors.New("admin permission required")
	}

	return nil
}

// TODO: Implement these utility functions
func hashToken(token string) string {
	// TODO: Implement secure token hashing
	return token // Placeholder
}
