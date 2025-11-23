package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/email"
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
	ListAllOrganizations(ctx context.Context) ([]*OrganizationResponse, error) // Admin only

	// Membership management
	InviteUser(ctx context.Context, req *InviteUserRequest) (*models.OrganizationInvitation, error)
	AcceptInvitation(ctx context.Context, token string, userID string) (*models.OrganizationMembership, error)
	GetMembership(ctx context.Context, orgID, userID string) (*models.OrganizationMembership, error)
	UpdateMembership(ctx context.Context, orgID, userID string, req *UpdateMembershipRequest) (*models.OrganizationMembership, error)
	RemoveMember(ctx context.Context, orgID, userID string) error
	ListMembers(ctx context.Context, orgID string, search ...string) ([]*OrganizationMember, error)

	// Invitation management
	CancelInvitation(ctx context.Context, invitationID string) error
	ResendInvitation(ctx context.Context, invitationID string) (*models.OrganizationInvitation, error)
	ListPendingInvitations(ctx context.Context, orgID string, search ...string) ([]*models.OrganizationInvitation, error)
	GetInvitationByToken(ctx context.Context, token string) (*InvitationDetails, error)
}

// CreateOrganizationRequest represents organization creation request
type CreateOrganizationRequest struct {
	UserID      string  `json:"user_id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug,omitempty"` // Auto-generated if not provided
	Description *string `json:"description,omitempty"`
}

// OrganizationResponse represents organization response
type OrganizationResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description"`
	Status      string     `json:"status"`
	CreatedBy   string     `json:"created_by"`
	Owner       *OwnerInfo `json:"owner,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	MemberCount int        `json:"member_count"`
}

// OwnerInfo represents organization owner information
type OwnerInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UpdateOrganizationRequest represents organization update request
type UpdateOrganizationRequest struct {
	Name        string  `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// InviteUserRequest represents user invitation request
type InviteUserRequest struct {
	OrganizationID string `json:"organization_id"`
	Email          string `json:"email" binding:"required"`
	RoleName       string `json:"role" binding:"required"` // Role name to lookup (e.g., "owner", "student")
}

// InvitationDetails represents public invitation details
type InvitationDetails struct {
	Email            string    `json:"email"`
	OrganizationName string    `json:"organization_name"`
	RoleName         string    `json:"role_name"`
	ExpiresAt        time.Time `json:"expires_at"`
	Status           string    `json:"status"`
}

// UpdateMembershipRequest represents membership update request
type UpdateMembershipRequest struct {
	RoleName string `json:"role_name,omitempty"` // Role name to update to
	Status   string `json:"status,omitempty"`
}

// OrganizationMember represents organization member with profile
type OrganizationMember struct {
	UserID         string     `json:"user_id"`
	Email          string     `json:"email"`
	FirstName      *string    `json:"first_name"`
	LastName       *string    `json:"last_name"`
	RoleName       string     `json:"role_name"` // Role name for display
	RoleID         string     `json:"role_id"`   // Role ID
	Status         string     `json:"status"`
	JoinedAt       *time.Time `json:"joined_at"`
	LastActivityAt *time.Time `json:"last_activity_at"`
}

// organizationService implements OrganizationService interface
type organizationService struct {
	repo         repository.Repository
	auditLogger  *logger.AuditLogger
	emailService email.Service
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(repo repository.Repository, emailService email.Service) OrganizationService {
	return &organizationService{
		repo:         repo,
		auditLogger:  logger.NewAuditLogger(),
		emailService: emailService,
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

	// Create default OWNER role (CUSTOM role, not system) with all permissions
	// This replaces the previous admin role creation
	ownerRole, err := s.repo.CreateDefaultAdminRole(ctx, org.ID.String(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create owner role: %w", err)
	}

	// Add creator as owner member with owner role
	membership := &models.OrganizationMembership{
		OrganizationID: org.ID,
		UserID:         uuid.MustParse(userID),
		RoleID:         ownerRole.ID, // Use the created owner role ID
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

// ListAllOrganizations lists all organizations (admin only)
func (s *organizationService) ListAllOrganizations(ctx context.Context) ([]*OrganizationResponse, error) {
	// Use a high limit to get all organizations for admin
	orgs, err := s.repo.Organization().List(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list all organizations: %w", err)
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

	// Validate invitation
	if err := s.validateInviteUserRequest(req); err != nil {
		return nil, err
	}

	// Check if user is already a member
	_, err := s.repo.OrganizationMembership().GetByOrganizationAndEmail(ctx, req.OrganizationID, req.Email)
	if err == nil {
		return nil, errors.New("user is already a member of this organization")
	}

	// Check for existing invitation (pending or cancelled)
	existingInvitation, err := s.repo.OrganizationInvitation().GetByOrganizationAndEmail(ctx, req.OrganizationID, req.Email)
	if err == nil {
		if existingInvitation.Status == models.InvitationStatusPending {
			return nil, errors.New("pending invitation already exists for this email")
		}
		// If invitation was cancelled or expired, delete it so we can create a new one
		if existingInvitation.Status == models.InvitationStatusCancelled || existingInvitation.Status == models.InvitationStatusExpired {
			if err := s.repo.OrganizationInvitation().Delete(ctx, existingInvitation.ID.String()); err != nil {
				return nil, fmt.Errorf("failed to delete old invitation: %w", err)
			}
		}
	}

	// Lookup role by name - if not provided, use a default role
	var role *models.Role

	if req.RoleName == "" {
		// Default to "student" role if no role specified
		role, err = s.repo.Role().GetByOrganizationAndName(ctx, req.OrganizationID, "student")
		if err != nil {
			// If student role doesn't exist, try to get any non-system role
			return nil, errors.New("no default role found - please specify a role name")
		}
	} else {
		role, err = s.repo.Role().GetByOrganizationAndName(ctx, req.OrganizationID, req.RoleName)
		if err != nil {
			return nil, fmt.Errorf("role not found: %w", err)
		}
	}

	// Prevent inviting users with system roles (admin)
	if role.IsSystem {
		return nil, errors.New("cannot invite users with system roles - admin role is reserved for organization owners")
	}

	// Generate invitation token
	token := generateSecureToken()

	invitation := &models.OrganizationInvitation{
		OrganizationID: uuid.MustParse(req.OrganizationID),
		Email:          req.Email,
		TokenHash:      hashToken(token),
		RoleID:         role.ID,
		InvitedBy:      uuid.MustParse(userID),
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.repo.OrganizationInvitation().Create(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Get organization and inviter details for email
	org, err := s.repo.Organization().GetByID(ctx, req.OrganizationID)
	if err != nil {
		// Log error but don't fail invitation creation
		fmt.Printf("Failed to get organization for email: %v\n", err)
	} else {
		inviter, err := s.repo.User().GetByID(ctx, userID)
		if err != nil {
			fmt.Printf("Failed to get inviter for email: %v\n", err)
		} else {
			// Build inviter name, handling nil pointers
			var inviterName string
			if inviter.Firstname != nil && inviter.Lastname != nil {
				inviterName = fmt.Sprintf("%s %s", *inviter.Firstname, *inviter.Lastname)
			} else if inviter.Firstname != nil {
				inviterName = *inviter.Firstname
			} else if inviter.Lastname != nil {
				inviterName = *inviter.Lastname
			} else {
				inviterName = inviter.Email
			}

			// Trim whitespace
			inviterName = strings.TrimSpace(inviterName)
			if inviterName == "" {
				inviterName = inviter.Email
			}

			// Send invitation email
			if err := s.emailService.SendInvitationEmail(req.Email, inviterName, org.Name, token); err != nil {
				// Log error but don't fail invitation
				fmt.Printf("Failed to send invitation email: %v\n", err)
			}
		}
	}

	s.auditLogger.LogOrganizationAction(userID, "invite_user", req.OrganizationID, "", "", true, nil, fmt.Sprintf("Invited %s as %s", req.Email, req.RoleName))

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
		RoleID:         invitation.RoleID,
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

	// Get organization to check if user is the owner
	org, err := s.repo.Organization().GetByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Prevent changing the organization owner's role
	if org.CreatedBy.String() == userID {
		return nil, errors.New("cannot change the organization owner's role")
	}

	membership, err := s.repo.OrganizationMembership().GetByOrganizationAndUser(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.RoleName != "" {
		// Lookup role by name
		role, err := s.repo.Role().GetByOrganizationAndName(ctx, orgID, req.RoleName)
		if err != nil {
			return nil, fmt.Errorf("role not found: %w", err)
		}

		// Prevent assigning system roles (admin) through membership update
		if role.IsSystem {
			return nil, errors.New("cannot assign system roles - admin role is reserved for organization owners")
		}

		membership.RoleID = role.ID
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

// RemoveMember removes a user from an organization
func (s *organizationService) RemoveMember(ctx context.Context, orgID, userID string) error {
	currentUserID, _ := ctx.Value("user_id").(string)

	// Get organization to check if user is the owner
	org, err := s.repo.Organization().GetByID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	// Prevent removing the organization owner
	if org.CreatedBy.String() == userID {
		return errors.New("cannot remove the organization owner")
	}

	if err := s.repo.OrganizationMembership().Delete(ctx, orgID, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	s.auditLogger.LogOrganizationAction(currentUserID, "remove_member", orgID, "", "", true, nil, fmt.Sprintf("Removed member %s", userID))

	return nil
}

// ListMembers lists organization members
func (s *organizationService) ListMembers(ctx context.Context, orgID string, search ...string) ([]*OrganizationMember, error) {
	var searchTerm string
	if len(search) > 0 {
		searchTerm = search[0]
	}

	memberships, err := s.repo.OrganizationMembership().GetByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}

	members := make([]*OrganizationMember, 0, len(memberships))
	for _, membership := range memberships {
		user, err := s.repo.User().GetByID(ctx, membership.UserID.String())
		if err != nil {
			continue // Skip if user not found
		}

		// Apply search filter if provided
		if searchTerm != "" {
			searchLower := strings.ToLower(searchTerm)
			emailMatch := strings.Contains(strings.ToLower(user.Email), searchLower)

			var firstnameMatch, lastnameMatch bool
			if user.Firstname != nil {
				firstnameMatch = strings.Contains(strings.ToLower(*user.Firstname), searchLower)
			}
			if user.Lastname != nil {
				lastnameMatch = strings.Contains(strings.ToLower(*user.Lastname), searchLower)
			}

			if !emailMatch && !firstnameMatch && !lastnameMatch {
				continue // Skip if search doesn't match
			}
		}

		// Load role information
		var roleName string
		var roleID string
		if membership.Role != nil {
			roleName = membership.Role.Name
			roleID = membership.Role.ID.String()
		} else {
			// If role not preloaded, fetch it
			role, err := s.repo.Role().GetByID(ctx, membership.RoleID.String())
			if err == nil {
				roleName = role.Name
				roleID = role.ID.String()
			}
		}

		members = append(members, &OrganizationMember{
			UserID:         user.ID.String(),
			Email:          user.Email,
			FirstName:      user.Firstname,
			LastName:       user.Lastname,
			RoleName:       roleName,
			RoleID:         roleID,
			Status:         membership.Status,
			JoinedAt:       membership.JoinedAt,
			LastActivityAt: membership.LastActivityAt,
		})
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
		// Permission check is now handled by middleware
		return errors.New("not authorized to cancel this invitation")
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
		// Permission check is now handled by middleware
		return nil, errors.New("not authorized to resend this invitation")
	}

	// Extend expiry and update
	invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	if err := s.repo.OrganizationInvitation().Update(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	// Send invitation email again
	if s.emailService != nil {
		// Get organization and inviter details
		org, err := s.repo.Organization().GetByID(ctx, invitation.OrganizationID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get organization: %w", err)
		}

		inviter, err := s.repo.User().GetByID(ctx, invitation.InvitedBy.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get inviter: %w", err)
		}

		// Build inviter name
		inviterName := "Someone"
		if inviter.Firstname != nil && inviter.Lastname != nil {
			firstName := strings.TrimSpace(*inviter.Firstname)
			lastName := strings.TrimSpace(*inviter.Lastname)
			if firstName != "" && lastName != "" {
				inviterName = firstName + " " + lastName
			}
		}

		// Generate new token for resend
		token := generateSecureToken()
		invitation.TokenHash = hashToken(token)
		if err := s.repo.OrganizationInvitation().Update(ctx, invitation); err != nil {
			return nil, fmt.Errorf("failed to update invitation token: %w", err)
		}

		// Send email
		if err := s.emailService.SendInvitationEmail(invitation.Email, inviterName, org.Name, token); err != nil {
			fmt.Printf("Failed to send invitation email: %v\n", err)
			// Don't fail the resend if email fails
		}
	}

	s.auditLogger.LogOrganizationAction(userID, "resend_invitation", invitation.OrganizationID.String(), "", "", true, nil, "Invitation resent")

	return invitation, nil
}

// GetInvitationByToken retrieves invitation details by token (public endpoint)
func (s *organizationService) GetInvitationByToken(ctx context.Context, token string) (*InvitationDetails, error) {
	tokenHash := hashToken(token)

	invitation, err := s.repo.OrganizationInvitation().GetByToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invitation not found: %w", err)
	}

	// Check if invitation is still valid
	if invitation.Status != models.InvitationStatusPending {
		return nil, errors.New("invitation is no longer valid")
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("invitation has expired")
	}

	// Get organization details
	org, err := s.repo.Organization().GetByID(ctx, invitation.OrganizationID.String())
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Get role details
	role, err := s.repo.Role().GetByID(ctx, invitation.RoleID.String())
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	return &InvitationDetails{
		Email:            invitation.Email,
		OrganizationName: org.Name,
		RoleName:         role.Name,
		ExpiresAt:        invitation.ExpiresAt,
		Status:           invitation.Status,
	}, nil
}

// ListPendingInvitations lists pending invitations for an organization
func (s *organizationService) ListPendingInvitations(ctx context.Context, orgID string, search ...string) ([]*models.OrganizationInvitation, error) {
	var searchTerm string
	if len(search) > 0 {
		searchTerm = search[0]
	}

	invitations, err := s.repo.OrganizationInvitation().GetPendingByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Apply search filter if provided
	if searchTerm != "" {
		searchLower := strings.ToLower(searchTerm)
		filteredInvitations := make([]*models.OrganizationInvitation, 0, len(invitations))

		for _, invitation := range invitations {
			if strings.Contains(strings.ToLower(invitation.Email), searchLower) {
				filteredInvitations = append(filteredInvitations, invitation)
			}
		}

		return filteredInvitations, nil
	}

	return invitations, nil
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

	// Role is optional - will default to a basic role if not specified

	// Basic email validation
	if !strings.Contains(req.Email, "@") {
		return errors.New("invalid email format")
	}

	return nil
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

	// Get owner information
	var owner *OwnerInfo
	if creator, err := s.repo.User().GetByID(ctx, org.CreatedBy.String()); err == nil {
		firstName := ""
		lastName := ""
		if creator.Firstname != nil {
			firstName = *creator.Firstname
		}
		if creator.Lastname != nil {
			lastName = *creator.Lastname
		}

		owner = &OwnerInfo{
			ID:        creator.ID.String(),
			Email:     creator.Email,
			FirstName: firstName,
			LastName:  lastName,
		}
	}

	return &OrganizationResponse{
		ID:          org.ID.String(),
		Name:        org.Name,
		Slug:        org.Slug,
		Description: org.Description,
		Status:      org.Status,
		CreatedBy:   org.CreatedBy.String(),
		Owner:       owner,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
		MemberCount: int(memberCount),
	}
}
