package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/email"
	"auth-service/pkg/jwt"
	"auth-service/pkg/logger"
	"auth-service/pkg/password"
	"auth-service/pkg/validation"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ───────────────────────────────────────────────────────────────────────────────
// PUBLIC INTERFACE
// ───────────────────────────────────────────────────────────────────────────────

// UserService defines the interface for user business logic (Slack-style multi-org)
type UserService interface {
	// GLOBAL AUTH (Slack-style multi-organization)
	RegisterGlobal(ctx context.Context, req *RegisterGlobalRequest) (*RegisterGlobalResponse, error)
	LoginGlobal(ctx context.Context, req *LoginGlobalRequest) (*LoginGlobalResponse, error)
	SelectOrganization(ctx context.Context, req *SelectOrganizationRequest) (*SelectOrganizationResponse, error)
	CreateOrganization(ctx context.Context, userID string, req *CreateOrganizationRequest) (*CreateOrganizationResponse, error)
	GetMyOrganizations(ctx context.Context, userID string) ([]*OrganizationMembership, error)

	// ORG-SCOPED AUTH
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error)
	Logout(ctx context.Context, req *LogoutRequest) error

	// USER PROFILE
	GetProfile(ctx context.Context, userID string) (*UserProfile, error)
	UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*UserProfile, error)
	ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error

	// EMAIL VERIFICATION
	SendVerificationEmail(ctx context.Context, userID string) error
	VerifyEmail(ctx context.Context, email, code string) error
	ResendVerificationEmail(ctx context.Context, email string) error

	// ADMIN
	ListUsers(ctx context.Context, limit int, cursor string) (*UserListResponse, error)
	ActivateUser(ctx context.Context, userID string) error
	DeactivateUser(ctx context.Context, userID string) error
	DeleteUser(ctx context.Context, userID string) error

	// SESSION MANAGEMENT
	GetUserSessions(ctx context.Context, userID string) ([]*models.UserSession, error)
	RevokeUserSession(ctx context.Context, userID, sessionID, reason string) error

	// DEPENDENCY INJECTION
	SetRedisClient(client *redis.Client)
	SetEmailService(emailSvc email.Service)
	SetSessionService(sessionSvc SessionService)
}

// ───────────────────────────────────────────────────────────────────────────────
// REQUEST / RESPONSE DTOs
// ───────────────────────────────────────────────────────────────────────────────

// --- GLOBAL REGISTRATION (no organization yet) ---

type RegisterGlobalRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	FirstName       string `json:"first_name,omitempty"`
	LastName        string `json:"last_name,omitempty"`
	Phone           string `json:"phone,omitempty"`
	Address         string `json:"address,omitempty"`
	InvitationToken string `json:"invitation_token,omitempty"` // Optional: auto-accept invitation after registration
	ClientIP        string `json:"-"`
	UserAgent       string `json:"-"`
}

type RegisterGlobalResponse struct {
	User *UserProfile `json:"user"`
}

// --- GLOBAL LOGIN (returns list of organizations) ---

type LoginGlobalRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	ClientIP  string `json:"-"`
	UserAgent string `json:"-"`
}

type LoginGlobalResponse struct {
	User          *UserProfile              `json:"user"`
	Organizations []*OrganizationMembership `json:"organizations"`
}

// --- SELECT ORGANIZATION (get org-scoped token) ---

type SelectOrganizationRequest struct {
	UserID         string `json:"user_id"`         // Global auth context (from FE/global cookie)
	OrganizationID string `json:"organization_id"` // Chosen org
	ClientIP       string `json:"-"`
	UserAgent      string `json:"-"`
}

type SelectOrganizationResponse struct {
	User         *UserProfile            `json:"user"`
	Organization *OrganizationMembership `json:"organization"`
	Token        *TokenPair              `json:"token"`
}

// --- CREATE ORGANIZATION (Slack-style workspace) ---
// Note: CreateOrganizationRequest/Response defined in organization_service.go

type CreateOrganizationResponse struct {
	Organization *OrganizationMembership `json:"organization"`
	Token        *TokenPair              `json:"token"` // org-scoped token for creator
}

// --- ORGANIZATION MEMBERSHIP DTO (lightweight) ---

type OrganizationMembership struct {
	OrganizationID   string     `json:"organization_id"`
	OrganizationName string     `json:"organization_name"`
	OrganizationSlug string     `json:"organization_slug"`
	Role             string     `json:"role"`
	Status           string     `json:"status"`
	JoinedAt         *time.Time `json:"joined_at,omitempty"`
}

// --- ORG-SCOPED AUTH ---

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	Token *TokenPair `json:"token"`
}

type LogoutRequest struct {
	UserID       string `json:"user_id"`
	SessionID    string `json:"session_id,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// --- PROFILE / PASSWORD ---

type UserProfile struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	FirstName    string     `json:"first_name,omitempty"`
	LastName     string     `json:"last_name,omitempty"`
	Phone        string     `json:"phone,omitempty"`
	Address      string     `json:"address,omitempty"`
	GlobalRole   string     `json:"global_role"`
	IsSuperadmin bool       `json:"is_superadmin"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
}

type UpdateProfileRequest struct {
	FirstName *string `json:"first_name,omitempty"` // nil = no change, "" = clear, "value" = update
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Address   *string `json:"address,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

// List users (for superadmin)
type UserListResponse struct {
	Users      []*UserProfile `json:"users"`
	Total      int64          `json:"total"`
	Limit      int            `json:"limit"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

// TokenPair represents access and refresh token pair (ORG-SCOPED)
type TokenPair struct {
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token"`
	ExpiresIn      int64  `json:"expires_in"`
	TokenType      string `json:"token_type"`
	SessionID      string `json:"session_id"`
	OrganizationID string `json:"organization_id"`
}

// ───────────────────────────────────────────────────────────────────────────────
// ERRORS & CONSTANTS
// ───────────────────────────────────────────────────────────────────────────────

var (
	ErrUserAlreadyExists   = errors.New("user with this email already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrOrgNotFound         = errors.New("organization not found or inactive")
	ErrMembershipNotFound  = errors.New("user is not a member of this organization")
	ErrMembershipSuspended = errors.New("membership is not active")
)

// Account lockout constants (global / email + IP)
const (
	MaxFailedAttempts   = 5
	LockoutDuration     = 15 * time.Minute
	FailedAttemptWindow = 15 * time.Minute
)

// Email verification rate limiting constants
const (
	MaxVerificationAttempts     = 5                // Max attempts per time window
	VerificationAttemptWindow   = 5 * time.Minute  // Time window for attempts
	VerificationLockoutDuration = 15 * time.Minute // Lockout duration after max attempts
)

// ───────────────────────────────────────────────────────────────────────────────
// SERVICE IMPLEMENTATION
// ───────────────────────────────────────────────────────────────────────────────

type userService struct {
	repo            repository.Repository
	jwtService      *jwt.Service
	passwordService *password.Service
	redisClient     *redis.Client
	emailSvc        email.Service
	sessionSvc      SessionService
	auditLogger     *logger.AuditLogger
}

// NewUserService creates a new user service
func NewUserService(repo repository.Repository, jwtService *jwt.Service, passwordService *password.Service) UserService {
	return &userService{
		repo:            repo,
		jwtService:      jwtService,
		passwordService: passwordService,
		redisClient:     nil,
		emailSvc:        nil,
		sessionSvc:      nil,
		auditLogger:     logger.NewAuditLogger(),
	}
}

// DI setters
func (s *userService) SetRedisClient(client *redis.Client)    { s.redisClient = client }
func (s *userService) SetEmailService(emailSvc email.Service) { s.emailSvc = emailSvc }
func (s *userService) SetSessionService(sessionSvc SessionService) {
	s.sessionSvc = sessionSvc
}

// ───────────────────────────────────────────────────────────────────────────────
// GLOBAL REGISTRATION & LOGIN (NO ORG YET)
// ───────────────────────────────────────────────────────────────────────────────

// RegisterGlobal creates a GLOBAL user account (no organization yet)
func (s *userService) RegisterGlobal(ctx context.Context, req *RegisterGlobalRequest) (*RegisterGlobalResponse, error) {
	// Core validation
	if err := validation.ValidateUserRegistration(req.Email, req.Password, req.ConfirmPassword); err != nil {
		return nil, err
	}

	// Optional profile validation
	if req.FirstName != "" {
		if err := validation.ValidateName(req.FirstName, "first name"); err != nil {
			return nil, err
		}
	}
	if req.LastName != "" {
		if err := validation.ValidateName(req.LastName, "last name"); err != nil {
			return nil, err
		}
	}
	if req.Phone != "" {
		if err := validation.ValidatePhone(req.Phone); err != nil {
			return nil, err
		}
	}
	if req.Address != "" {
		if err := validation.ValidateAddress(req.Address); err != nil {
			return nil, err
		}
	}

	// Check for existing user (normalize email before lookup)
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))
	if existing, _ := s.repo.User().GetByEmail(ctx, normalizedEmail); existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hash, err := s.passwordService.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Prepare model
	user := &models.User{
		Email:        normalizedEmail,
		PasswordHash: hash,
		Firstname:    safeStringToPointer(req.FirstName),
		Lastname:     safeStringToPointer(req.LastName),
		Phone:        safeStringToPointer(req.Phone),
		Address:      safeStringToPointer(req.Address),
		Status:       models.UserStatusActive,
		GlobalRole:   "user",
		IsSuperadmin: false,
	}

	if err := s.repo.User().Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// If invitation token provided, automatically accept the invitation
	if req.InvitationToken != "" {
		fmt.Printf("=== AUTO-ACCEPTING INVITATION ===\n")
		fmt.Printf("User email: %s\n", user.Email)

		// Hash the token for secure lookup (tokens are stored as SHA256 hashes)
		tokenHash := hashToken(req.InvitationToken)

		// Get the invitation by token hash
		invitation, err := s.repo.OrganizationInvitation().GetByToken(ctx, tokenHash)
		if err != nil {
			fmt.Printf("ERROR: Failed to get invitation by token: %v\n", err)
		} else if invitation == nil {
			fmt.Printf("ERROR: Invitation not found for token\n")
		} else {
			fmt.Printf("Found invitation: ID=%s, Email=%s, Status=%s\n", invitation.ID, invitation.Email, invitation.Status)

			// Validate invitation status
			if invitation.Status != models.InvitationStatusPending {
				fmt.Printf("WARNING: Invitation status is not pending: %s\n", invitation.Status)
			} else if invitation.ExpiresAt.Before(time.Now()) {
				fmt.Printf("ERROR: Invitation expired at %s\n", invitation.ExpiresAt)
			} else if !strings.EqualFold(invitation.Email, user.Email) {
				fmt.Printf("ERROR: Email mismatch - invitation=%s, user=%s\n", invitation.Email, user.Email)
			} else {
				// Validate organization exists and is active
				org, err := s.repo.Organization().GetByID(ctx, invitation.OrganizationID.String())
				if err != nil || org == nil {
					fmt.Printf("ERROR: Organization not found: %v\n", err)
				} else if org.Status != models.OrganizationStatusActive {
					fmt.Printf("ERROR: Organization is not active - status=%s\n", org.Status)
				} else {
					// Validate role exists
					role, err := s.repo.Role().GetByID(ctx, invitation.RoleID.String())
					if err != nil || role == nil {
						fmt.Printf("ERROR: Role not found: %v\n", err)
					} else if role.OrganizationID == nil || *role.OrganizationID != invitation.OrganizationID {
						fmt.Printf("ERROR: Role organization mismatch - role.org=%v, invitation.org=%s\n", role.OrganizationID, invitation.OrganizationID)
					} else {
						fmt.Printf("Creating membership for user=%s in org=%s with roleID=%s\n", user.ID, invitation.OrganizationID, invitation.RoleID) // Create organization membership
						now := time.Now()
						membership := &models.OrganizationMembership{
							OrganizationID: invitation.OrganizationID,
							UserID:         user.ID,
							RoleID:         invitation.RoleID,
							Status:         models.MembershipStatusActive,
							InvitedBy:      &invitation.InvitedBy,
							InvitedAt:      &invitation.CreatedAt,
							JoinedAt:       &now,
						}

						if err := s.repo.OrganizationMembership().Create(ctx, membership); err != nil {
							fmt.Printf("ERROR: Failed to create membership: %v\n", err)
						} else {
							fmt.Printf("SUCCESS: Membership created\n")

							// Update invitation status
							invitation.Status = models.InvitationStatusAccepted
							invitation.AcceptedAt = &now
							if err := s.repo.OrganizationInvitation().Update(ctx, invitation); err != nil {
								fmt.Printf("ERROR: Failed to update invitation status: %v\n", err)
							} else {
								fmt.Printf("SUCCESS: Invitation marked as accepted\n")
							}
						}
					}
				}
			}
		}
		fmt.Printf("=== END AUTO-ACCEPT ===\n")
	}

	// Send verification email with 6-digit code
	if err := s.SendVerificationEmail(ctx, user.ID.String()); err != nil {
		// Log error but don't fail registration - user can resend later
		fmt.Printf("WARNING: Failed to send verification email to %s: %v\n", user.Email, err)
	}

	return &RegisterGlobalResponse{
		User: s.convertToUserProfile(user),
	}, nil
}

// LoginGlobal authenticates the GLOBAL account (no org yet) and returns org memberships
func (s *userService) LoginGlobal(ctx context.Context, req *LoginGlobalRequest) (*LoginGlobalResponse, error) {
	if err := validation.ValidateLogin(req.Email); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	if s.isAccountLocked(ctx, email, req.ClientIP) {
		return nil, errors.New("account temporarily locked due to failed attempts")
	}

	user, err := s.repo.User().GetByEmail(ctx, email)
	if err != nil || user == nil {
		// user may be nil / not found
		s.recordFailedAttempt(ctx, email, req.ClientIP, nil)
		return nil, ErrInvalidCredentials
	}

	if user.Status != models.UserStatusActive {
		return nil, errors.New("account is deactivated")
	}

	// Check if email is verified
	if user.EmailVerifiedAt == nil {
		return nil, errors.New("email not verified. Please check your email for the verification code")
	}

	valid, err := s.passwordService.Verify(req.Password, user.PasswordHash)
	if err != nil || !valid {
		s.recordFailedAttempt(ctx, email, req.ClientIP, &user.ID)
		return nil, ErrInvalidCredentials
	}

	// Clear lockout state
	s.clearFailedAttempts(ctx, email, req.ClientIP)

	// Update last login
	if err := s.repo.User().UpdateLastLogin(ctx, user.ID.String()); err != nil {
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	// Load org memberships (Slack-style)
	memberships, err := s.repo.OrganizationMembership().GetByUser(ctx, user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to load organizations: %w", err)
	}

	orgDTOs := make([]*OrganizationMembership, 0, len(memberships))
	for _, m := range memberships {
		org, err := s.repo.Organization().GetByID(ctx, m.OrganizationID.String())
		if err != nil || org == nil {
			continue
		}

		// Always load role by ID (GORM may not preload Role relation)
		role, err := s.repo.Role().GetByID(ctx, m.RoleID.String())
		roleName := ""
		if err == nil && role != nil {
			roleName = role.Name
		}

		orgDTOs = append(orgDTOs, &OrganizationMembership{
			OrganizationID:   org.ID.String(),
			OrganizationName: org.Name,
			OrganizationSlug: org.Slug,
			Role:             roleName,
			Status:           m.Status,
			JoinedAt:         m.JoinedAt,
		})
	}

	return &LoginGlobalResponse{
		User:          s.convertToUserProfile(user),
		Organizations: orgDTOs,
	}, nil
}

// GetMyOrganizations returns org memberships for a global user
func (s *userService) GetMyOrganizations(ctx context.Context, userID string) ([]*OrganizationMembership, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	memberships, err := s.repo.OrganizationMembership().GetByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load organizations: %w", err)
	}

	orgDTOs := make([]*OrganizationMembership, 0, len(memberships))
	for _, m := range memberships {
		org, err := s.repo.Organization().GetByID(ctx, m.OrganizationID.String())
		if err != nil || org == nil {
			continue
		}

		// Always load role by ID (GORM may not preload Role relation)
		role, err := s.repo.Role().GetByID(ctx, m.RoleID.String())
		roleName := ""
		if err == nil && role != nil {
			roleName = role.Name
		}

		orgDTOs = append(orgDTOs, &OrganizationMembership{
			OrganizationID:   org.ID.String(),
			OrganizationName: org.Name,
			OrganizationSlug: org.Slug,
			Role:             roleName,
			Status:           m.Status,
			JoinedAt:         m.JoinedAt,
		})
	}

	return orgDTOs, nil
}

// ───────────────────────────────────────────────────────────────────────────────
// ORGANIZATION CREATION & ORG-SCOPED TOKEN ISSUANCE
// ───────────────────────────────────────────────────────────────────────────────

// CreateOrganization creates a Slack-style organization and makes user=admin
func (s *userService) CreateOrganization(ctx context.Context, userID string, req *CreateOrganizationRequest) (*CreateOrganizationResponse, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	if err := validation.ValidateOrganizationName(req.Name); err != nil {
		return nil, err
	}
	if !validation.IsValidSlug(req.Slug) {
		return nil, errors.New("invalid organization slug format")
	}

	// Ensure slug is unique
	existing, err := s.repo.Organization().GetBySlug(ctx, req.Slug)
	if err == nil && existing != nil {
		return nil, errors.New("organization slug already in use")
	}
	// If error is not "record not found", return it
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check slug availability: %w", err)
	}

	creator, err := s.repo.User().GetByID(ctx, userID)
	if err != nil || creator == nil {
		return nil, errors.New("user not found")
	}
	if creator.Status != models.UserStatusActive {
		return nil, errors.New("user is not active")
	}

	org := &models.Organization{
		Name:      req.Name,
		Slug:      strings.ToLower(strings.TrimSpace(req.Slug)),
		Status:    models.OrganizationStatusActive,
		CreatedBy: creator.ID,
	}

	if req.Description != nil && *req.Description != "" {
		org.Description = req.Description
	}

	if err := s.repo.Organization().Create(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Create default admin role for this organization
	adminRole, err := s.repo.CreateDefaultAdminRole(ctx, org.ID.String(), creator.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to create default admin role: %w", err)
	}

	// Creator becomes admin member
	now := time.Now()
	m := &models.OrganizationMembership{
		OrganizationID: org.ID,
		UserID:         creator.ID,
		RoleID:         adminRole.ID,
		Status:         models.MembershipStatusActive,
		JoinedAt:       &now,
	}

	if err := s.repo.OrganizationMembership().Create(ctx, m); err != nil {
		return nil, fmt.Errorf("failed to create organization membership: %w", err)
	}

	// Create initial session + token for this org (no ClientIP in CreateOrganizationRequest)
	session, err := s.createSession(ctx, creator, org.ID, "", "org-create:"+req.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	tokenPair, refreshID, err := s.issueTokenPair(ctx, creator, org.ID, m.RoleID, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	if err := s.persistRefreshToken(ctx, refreshID, tokenPair.RefreshToken, creator.ID, org.ID, session.ID); err != nil {
		return nil, fmt.Errorf("failed to persist refresh token: %w", err)
	}

	// Load role for response
	var roleName string
	if m.Role != nil {
		roleName = m.Role.Name
	} else {
		role, err := s.repo.Role().GetByID(ctx, m.RoleID.String())
		if err == nil {
			roleName = role.Name
		}
	}

	orgDTO := &OrganizationMembership{
		OrganizationID:   org.ID.String(),
		OrganizationName: org.Name,
		OrganizationSlug: org.Slug,
		Role:             roleName,
		Status:           m.Status,
		JoinedAt:         m.JoinedAt,
	}

	return &CreateOrganizationResponse{
		Organization: orgDTO,
		Token:        tokenPair,
	}, nil
}

// SelectOrganization issues an ORG-scoped JWT for a chosen organization
func (s *userService) SelectOrganization(ctx context.Context, req *SelectOrganizationRequest) (*SelectOrganizationResponse, error) {
	if req.UserID == "" || req.OrganizationID == "" {
		return nil, errors.New("user_id and organization_id are required")
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user_id")
	}
	orgUUID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return nil, errors.New("invalid organization_id")
	}

	user, err := s.repo.User().GetByID(ctx, userUUID.String())
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	if user.Status != models.UserStatusActive {
		return nil, errors.New("account is deactivated")
	}

	org, err := s.repo.Organization().GetByID(ctx, orgUUID.String())
	if err != nil || org == nil {
		return nil, ErrOrgNotFound
	}
	if org.Status != models.OrganizationStatusActive {
		return nil, ErrOrgNotFound
	}

	membership, err := s.repo.OrganizationMembership().GetByOrganizationAndUser(ctx, orgUUID.String(), userUUID.String())
	if err != nil || membership == nil {
		return nil, ErrMembershipNotFound
	}
	if membership.Status != models.MembershipStatusActive {
		return nil, ErrMembershipSuspended
	}

	// Create session (org-scoped)
	session, err := s.createSession(ctx, user, org.ID, req.ClientIP, req.UserAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Issue org-scoped JWT + refresh
	tokenPair, refreshID, err := s.issueTokenPair(ctx, user, org.ID, membership.RoleID, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	if err := s.persistRefreshToken(ctx, refreshID, tokenPair.RefreshToken, user.ID, org.ID, session.ID); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Load role for response
	var roleName string
	if membership.Role != nil {
		roleName = membership.Role.Name
	} else {
		role, err := s.repo.Role().GetByID(ctx, membership.RoleID.String())
		if err == nil {
			roleName = role.Name
		}
	}

	orgDTO := &OrganizationMembership{
		OrganizationID:   org.ID.String(),
		OrganizationName: org.Name,
		OrganizationSlug: org.Slug,
		Role:             roleName,
		Status:           membership.Status,
		JoinedAt:         membership.JoinedAt,
	}

	return &SelectOrganizationResponse{
		User:         s.convertToUserProfile(user),
		Organization: orgDTO,
		Token:        tokenPair,
	}, nil
}

// ───────────────────────────────────────────────────────────────────────────────
// ORG-SCOPED REFRESH & LOGOUT
// ───────────────────────────────────────────────────────────────────────────────

func (s *userService) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, errors.New("refresh token is required")
	}

	claims, err := s.jwtService.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	refreshRecord, err := s.repo.RefreshToken().GetByID(ctx, claims.ID)
	if err != nil || refreshRecord == nil {
		return nil, errors.New("refresh token not recognized")
	}

	// Verify JTI matches (prevent replay attacks)
	if claims.ID != refreshRecord.ID.String() {
		_ = s.repo.RefreshToken().RevokeBySession(ctx, refreshRecord.SessionID.String(), "jti_mismatch")
		if s.sessionSvc != nil {
			_ = s.sessionSvc.RevokeSession(ctx, refreshRecord.SessionID.String(), "jti_mismatch")
		}
		return nil, errors.New("refresh token invalidated - JTI mismatch")
	}

	if refreshRecord.RevokedAt != nil {
		return nil, errors.New("refresh token already used or revoked")
	}

	if time.Now().After(refreshRecord.ExpiresAt) {
		_ = s.repo.RefreshToken().Revoke(ctx, refreshRecord.ID.String(), "expired")
		return nil, errors.New("refresh token expired")
	}

	// Verify bound session & user/org (SHA256 comparison)
	computedHash := hashToken(req.RefreshToken)

	if computedHash != refreshRecord.TokenHash {
		// Security measure: revoke entire session chain
		_ = s.repo.RefreshToken().RevokeBySession(ctx, refreshRecord.SessionID.String(), "token_hash_mismatch")
		if s.sessionSvc != nil {
			_ = s.sessionSvc.RevokeSession(ctx, refreshRecord.SessionID.String(), "token_hash_mismatch")
		}
		return nil, errors.New("refresh token invalidated")
	}

	user, err := s.repo.User().GetByID(ctx, refreshRecord.UserID.String())
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if user.Status != models.UserStatusActive {
		return nil, errors.New("account is deactivated")
	}

	// Optional: ensure membership still valid
	membership, err := s.repo.OrganizationMembership().GetByOrganizationAndUser(ctx, refreshRecord.OrganizationID.String(), user.ID.String())
	if err != nil || membership == nil || membership.Status != models.MembershipStatusActive {
		return nil, errors.New("organization membership is not active")
	}

	// Rotate refresh token (revoke old)
	if err := s.repo.RefreshToken().Revoke(ctx, refreshRecord.ID.String(), "rotated"); err != nil {
		return nil, fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	tokenPair, newRefreshID, err := s.issueTokenPair(ctx, user, refreshRecord.OrganizationID, membership.RoleID, refreshRecord.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	if err := s.persistRefreshToken(ctx, newRefreshID, tokenPair.RefreshToken, user.ID, refreshRecord.OrganizationID, refreshRecord.SessionID); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &RefreshTokenResponse{Token: tokenPair}, nil
}

func (s *userService) Logout(ctx context.Context, req *LogoutRequest) error {
	if req.UserID == "" {
		return errors.New("user ID is required")
	}

	// Revoke provided refresh token
	if req.RefreshToken != "" {
		if claims, err := s.jwtService.ParseRefreshToken(req.RefreshToken); err == nil {
			_ = s.repo.RefreshToken().Revoke(ctx, claims.ID, "logout")
		}
	}

	if req.SessionID != "" {
		if s.sessionSvc != nil {
			_ = s.sessionSvc.RevokeSession(ctx, req.SessionID, "logout")
		}
		_ = s.repo.UserSession().Revoke(ctx, req.SessionID, "logout")
		_ = s.repo.RefreshToken().RevokeBySession(ctx, req.SessionID, "logout")
	} else {
		// Kill all sessions for user
		_ = s.repo.UserSession().DeleteByUserID(ctx, req.UserID)
		_ = s.repo.RefreshToken().DeleteByUserID(ctx, req.UserID)
	}

	return nil
}

// ───────────────────────────────────────────────────────────────────────────────
// PROFILE & PASSWORD
// ───────────────────────────────────────────────────────────────────────────────

func (s *userService) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	user, err := s.repo.User().GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return s.convertToUserProfile(user), nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*UserProfile, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	// Validate provided fields (only if not nil)
	if req.FirstName != nil && *req.FirstName != "" {
		if err := validation.ValidateName(*req.FirstName, "first name"); err != nil {
			return nil, err
		}
	}
	if req.LastName != nil && *req.LastName != "" {
		if err := validation.ValidateName(*req.LastName, "last name"); err != nil {
			return nil, err
		}
	}
	if req.Phone != nil && *req.Phone != "" {
		if err := validation.ValidatePhone(*req.Phone); err != nil {
			return nil, err
		}
	}
	if req.Address != nil && *req.Address != "" {
		if err := validation.ValidateAddress(*req.Address); err != nil {
			return nil, err
		}
	}

	user, err := s.repo.User().GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields (nil = no change, empty string = clear field, value = update)
	if req.FirstName != nil {
		user.Firstname = req.FirstName
	}
	if req.LastName != nil {
		user.Lastname = req.LastName
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.Address != nil {
		user.Address = req.Address
	}

	if err := s.repo.User().Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return s.convertToUserProfile(user), nil
}

func (s *userService) ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	if err := validation.ValidatePassword(req.NewPassword); err != nil {
		return fmt.Errorf("invalid new password: %w", err)
	}
	if err := validation.ValidatePasswordsMatch(req.NewPassword, req.ConfirmPassword); err != nil {
		return err
	}

	user, err := s.repo.User().GetByID(ctx, userID)
	if err != nil || user == nil {
		return fmt.Errorf("user not found: %w", err)
	}

	valid, err := s.passwordService.Verify(req.CurrentPassword, user.PasswordHash)
	if err != nil || !valid {
		return errors.New("current password is incorrect")
	}

	passwordHashChan := s.hashPasswordAsync(req.NewPassword)
	res := <-passwordHashChan
	if res.Error != nil {
		return fmt.Errorf("failed to hash password: %w", res.Error)
	}

	if err := s.repo.User().UpdatePassword(ctx, userID, res.Hash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all refresh tokens
	_ = s.repo.RefreshToken().DeleteByUserID(ctx, userID)

	return nil
}

func (s *userService) ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error {
	if err := validation.ValidateForgotPassword(req.Email); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Normalize email before lookup
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.repo.User().GetByEmail(ctx, email)
	if err != nil || user == nil {
		// Don't reveal if user exists
		return nil
	}

	if user.Status != models.UserStatusActive {
		return nil
	}

	// Create secure random token (string) for email link
	rawToken := generateCryptographicallySecureToken()

	// Hash token for DB using SHA256 (same as refresh tokens - bcrypt is too slow for tokens)
	tokenHash := hashToken(rawToken)

	reset := &models.PasswordReset{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.repo.PasswordReset().Create(ctx, reset); err != nil {
		return fmt.Errorf("failed to create password reset: %w", err)
	}

	if s.emailSvc != nil {
		if err := s.emailSvc.SendPasswordResetEmail(email, rawToken); err != nil {
			fmt.Printf("Failed to send password reset email: %v\n", err)
		}
	} else {
		fmt.Printf("Password reset token for %s: %s\n", email, rawToken)
	}

	return nil
}

func (s *userService) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	if err := validation.ValidatePasswordReset(req.Token, req.NewPassword, req.ConfirmPassword); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Look up reset record via repository, which should verify token hash
	reset, err := s.repo.PasswordReset().GetByToken(ctx, req.Token)
	if err != nil || reset == nil {
		return errors.New("invalid or expired reset token")
	}

	if time.Now().After(reset.ExpiresAt) {
		return errors.New("reset token expired")
	}

	user, err := s.repo.User().GetByID(ctx, reset.UserID.String())
	if err != nil || user == nil {
		return fmt.Errorf("user not found: %w", err)
	}

	passwordHashChan := s.hashPasswordAsync(req.NewPassword)
	res := <-passwordHashChan
	if res.Error != nil {
		return fmt.Errorf("failed to hash password: %w", res.Error)
	}

	if err := s.repo.User().UpdatePassword(ctx, user.ID.String(), res.Hash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete reset token by ID (not raw token - DB stores hash)
	_ = s.repo.PasswordReset().DeleteByID(ctx, reset.ID.String())
	_ = s.repo.RefreshToken().DeleteByUserID(ctx, user.ID.String())

	return nil
}

// ───────────────────────────────────────────────────────────────────────────────
// EMAIL VERIFICATION
// ───────────────────────────────────────────────────────────────────────────────

func (s *userService) SendVerificationEmail(ctx context.Context, userID string) error {
	user, err := s.repo.User().GetByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// Already verified
	if user.EmailVerifiedAt != nil {
		return errors.New("email already verified")
	}

	// Generate 6-digit verification code
	code := generateVerificationCode()

	// Store code (plain text is OK for short-lived codes)
	expiresAt := time.Now().Add(15 * time.Minute) // 15 minutes expiry
	user.EmailVerificationToken = &code
	user.EmailVerificationExpiresAt = &expiresAt

	if err := s.repo.User().Update(ctx, user); err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	// Send verification email
	if s.emailSvc != nil {
		if err := s.emailSvc.SendVerificationEmail(user.Email, code); err != nil {
			fmt.Printf("Failed to send verification email: %v\n", err)
			return fmt.Errorf("failed to send verification email: %w", err)
		}
	} else {
		// Dev mode: print to console
		fmt.Printf("📧 Verification code for %s: %s\n", user.Email, code)
		fmt.Printf("🔗 Code expires at: %s\n", expiresAt.Format(time.RFC3339))
	}

	return nil
}

func (s *userService) VerifyEmail(ctx context.Context, email, code string) error {
	if code == "" {
		return errors.New("verification code is required")
	}

	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	// Check rate limiting FIRST (before any DB queries to prevent enumeration)
	if s.isVerificationRateLimited(ctx, email) {
		return errors.New("too many verification attempts - please wait 15 minutes before trying again")
	}

	// Find user by email
	user, err := s.repo.User().GetByEmail(ctx, email)
	if err != nil || user == nil {
		// Record failed attempt even for non-existent users to prevent enumeration
		s.recordVerificationAttempt(ctx, email)
		return errors.New("invalid verification code")
	}

	// Check if already verified
	if user.EmailVerifiedAt != nil {
		return errors.New("email already verified")
	}

	// Check if code exists
	if user.EmailVerificationToken == nil {
		s.recordVerificationAttempt(ctx, email)
		return errors.New("no verification code found - please request a new code")
	}

	// Check if token expired
	if user.EmailVerificationExpiresAt == nil || time.Now().After(*user.EmailVerificationExpiresAt) {
		s.recordVerificationAttempt(ctx, email)
		return errors.New("verification code expired - please request a new code")
	}

	// Verify code matches
	if *user.EmailVerificationToken != code {
		s.recordVerificationAttempt(ctx, email)
		return errors.New("invalid verification code")
	}

	// SUCCESS: Mark email as verified
	now := time.Now()
	user.EmailVerifiedAt = &now
	user.EmailVerificationToken = nil
	user.EmailVerificationExpiresAt = nil

	if err := s.repo.User().Update(ctx, user); err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	// Clear verification attempts on success
	s.clearVerificationAttempts(ctx, email)

	fmt.Printf("✅ Email verified for user: %s\n", user.Email)
	return nil
}

func (s *userService) ResendVerificationEmail(ctx context.Context, email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.repo.User().GetByEmail(ctx, email)
	if err != nil || user == nil {
		// Don't reveal if user exists
		return nil
	}

	// Already verified
	if user.EmailVerifiedAt != nil {
		return errors.New("email already verified")
	}

	// Rate limiting: check if code was sent recently (within last 60 seconds)
	if user.EmailVerificationExpiresAt != nil {
		timeRemaining := time.Until(user.EmailVerificationExpiresAt.Add(-14 * time.Minute)) // 15min - 14min = 1min since sent
		if timeRemaining > 0 {
			return fmt.Errorf("please wait %d seconds before requesting a new code", int(timeRemaining.Seconds()))
		}
	}

	// Send new verification email
	return s.SendVerificationEmail(ctx, user.ID.String())
}

// generateVerificationCode generates a random 6-digit code
func generateVerificationCode() string {
	b := make([]byte, 3) // 3 bytes = 6 hex digits
	rand.Read(b)
	code := fmt.Sprintf("%06d", (int(b[0])<<16|int(b[1])<<8|int(b[2]))%1000000)
	return code
}

// ───────────────────────────────────────────────────────────────────────────────
// ADMIN (GLOBAL SUPERADMIN ONLY)
// ───────────────────────────────────────────────────────────────────────────────

func (s *userService) ListUsers(ctx context.Context, limit int, cursor string) (*UserListResponse, error) {
	isSuperadmin, _ := ctx.Value("is_superadmin").(bool)
	if !isSuperadmin {
		return nil, errors.New("insufficient permissions")
	}

	users, err := s.repo.User().List(ctx, limit, cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.repo.User().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	profiles := make([]*UserProfile, len(users))
	for i, u := range users {
		profiles[i] = s.convertToUserProfile(u)
	}

	var nextCursor string
	if len(users) == limit {
		nextCursor = users[len(users)-1].ID.String()
	}

	return &UserListResponse{
		Users:      profiles,
		Total:      total,
		Limit:      limit,
		NextCursor: nextCursor,
	}, nil
}

func (s *userService) ActivateUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	isSuperadmin, _ := ctx.Value("is_superadmin").(bool)
	if !isSuperadmin {
		return errors.New("superadmin permission required")
	}

	err := s.repo.User().Activate(ctx, userID)
	if err != nil {
		s.auditLogger.LogAdminAction("system", "activate_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), false, err, "Failed to activate user account")
		return err
	}

	s.auditLogger.LogAdminAction("system", "activate_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), true, nil, "User account activated successfully")
	return nil
}

func (s *userService) DeactivateUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	isSuperadmin, _ := ctx.Value("is_superadmin").(bool)
	if !isSuperadmin {
		return errors.New("superadmin permission required")
	}

	_ = s.repo.UserSession().DeleteByUserID(ctx, userID)
	_ = s.repo.RefreshToken().DeleteByUserID(ctx, userID)

	err := s.repo.User().Deactivate(ctx, userID)
	if err != nil {
		s.auditLogger.LogAdminAction("system", "deactivate_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), false, err, "Failed to deactivate user account")
		return err
	}

	s.auditLogger.LogAdminAction("system", "deactivate_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), true, nil, "User account deactivated successfully")
	return nil
}

func (s *userService) DeleteUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	isSuperadmin, _ := ctx.Value("is_superadmin").(bool)
	if !isSuperadmin {
		return errors.New("superadmin permission required")
	}

	tx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		s.auditLogger.LogAdminAction("system", "delete_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), false, err, "Failed to start transaction for user deletion")
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if err := tx.UserSession().DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}
	if err := tx.RefreshToken().DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete refresh tokens: %w", err)
	}
	if err := tx.PasswordReset().DeleteExpired(ctx); err != nil {
		fmt.Printf("Failed to cleanup password resets: %v\n", err)
	}

	if err := tx.User().Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.auditLogger.LogAdminAction("system", "delete_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), true, nil, "User account deleted successfully")
	return nil
}

// ───────────────────────────────────────────────────────────────────────────────
// SESSION MANAGEMENT
// ───────────────────────────────────────────────────────────────────────────────

func (s *userService) GetUserSessions(ctx context.Context, userID string) ([]*models.UserSession, error) {
	if s.sessionSvc != nil {
		return s.sessionSvc.GetUserSessions(ctx, userID)
	}
	return s.repo.UserSession().GetByUserID(ctx, userID)
}

func (s *userService) RevokeUserSession(ctx context.Context, userID, sessionID, reason string) error {
	if userID == "" || sessionID == "" {
		return errors.New("user ID and session ID are required")
	}

	session, err := s.repo.UserSession().GetByID(ctx, sessionID)
	if err != nil || session == nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if session.UserID.String() != userID {
		return errors.New("session does not belong to user")
	}

	if s.sessionSvc != nil {
		return s.sessionSvc.RevokeSession(ctx, sessionID, reason)
	}

	return s.repo.UserSession().Revoke(ctx, sessionID, reason)
}

// ───────────────────────────────────────────────────────────────────────────────
// INTERNAL HELPERS
// ───────────────────────────────────────────────────────────────────────────────

// createSession creates a persistent session record (ORG-SCOPED)
func (s *userService) createSession(ctx context.Context, user *models.User, organizationID uuid.UUID, clientIP, userAgent string) (*models.UserSession, error) {
	if s.sessionSvc != nil {
		return s.sessionSvc.CreateSession(ctx, user.ID.String(), organizationID.String(), clientIP, userAgent)
	}

	session := &models.UserSession{
		UserID:         user.ID,
		OrganizationID: organizationID,
		TokenHash:      generateCryptographicallySecureToken(),
		IPAddress:      clientIP,
		UserAgent:      userAgent,
		IsActive:       true,
		LastActivity:   time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}

	if err := s.repo.UserSession().Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// issueTokenPair generates org- & session-bound JWTs and returns refresh token ID
func (s *userService) issueTokenPair(ctx context.Context, user *models.User, organizationID uuid.UUID, roleID uuid.UUID, sessionID uuid.UUID) (*TokenPair, string, error) {
	// Load role to get name and organization context
	role, err := s.repo.Role().GetByID(ctx, roleID.String())
	if err != nil {
		return nil, "", fmt.Errorf("failed to load role: %w", err)
	}

	// Get user permissions for this role (filtered by user type)
	// Superadmin: gets system + org permissions
	// Org admin/user: gets ONLY org permissions (custom roles)
	permissions, err := s.getRolePermissionsFiltered(ctx, user, role, organizationID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load permissions: %w", err)
	}

	tokenCtx := &jwt.TokenContext{
		UserID:           user.ID,
		OrganizationID:   organizationID,
		SessionID:        sessionID,
		RoleID:           roleID,
		Email:            user.Email,
		GlobalRole:       user.GlobalRole,
		OrganizationRole: role.Name,
		Permissions:      permissions,
		IsSuperadmin:     user.IsSuperadmin,
	}

	accessToken, err := s.jwtService.GenerateAccessToken(tokenCtx)
	if err != nil {
		return nil, "", err
	}

	refreshToken, refreshID, err := s.jwtService.GenerateRefreshToken(tokenCtx)
	if err != nil {
		return nil, "", err
	}

	tokenPair := &TokenPair{
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		ExpiresIn:      3600,
		TokenType:      "Bearer",
		SessionID:      sessionID.String(),
		OrganizationID: organizationID.String(),
	}

	return tokenPair, refreshID, nil
}

// getRolePermissionsFiltered fetches permissions filtered by user type
// Superadmin: system permissions + organization permissions
// Regular user: ONLY organization custom permissions (no system permissions)
func (s *userService) getRolePermissionsFiltered(ctx context.Context, user *models.User, role *models.Role, orgID uuid.UUID) ([]string, error) {
	// CRITICAL: System roles should NOT be assigned to organizations
	// If somehow a system role is found, handle it carefully
	if role.IsSystem && role.OrganizationID == nil {
		// System role detected - only superadmin should have these
		if user.IsSuperadmin {
			// Superadmin with system role: return all permissions
			return models.DefaultAdminPermissions(), nil
		} else {
			// ERROR: Non-superadmin should never have a system role
			return nil, errors.New("invalid role assignment: system role assigned to non-superadmin")
		}
	}

	// For custom organization roles (IsSystem=false, OrganizationID != nil)
	// Fetch permissions from DB (filtered by organization context)
	perms, err := s.repo.Permission().GetRolePermissions(ctx, role.ID)
	if err != nil {
		return nil, err
	}

	permissions := make([]string, 0, len(perms))
	for _, p := range perms {
		// Filter permissions based on user type and organization context
		if user.IsSuperadmin {
			// Superadmin sees ALL permissions (system + org-specific)
			permissions = append(permissions, p.Name)
		} else {
			// Regular users: ONLY include permissions that are:
			// - System permissions (IsSystem=true, OrganizationID=nil), OR
			// - Custom permissions for THIS organization (OrganizationID = orgID)
			if p.IsSystem && p.OrganizationID == nil {
				// Include system permission
				permissions = append(permissions, p.Name)
			} else if p.OrganizationID != nil && *p.OrganizationID == orgID {
				// Include org-specific permission
				permissions = append(permissions, p.Name)
			}
			// Skip permissions from other organizations
		}
	}

	return permissions, nil
}

// DEPRECATED: getRolePermissionsWithCache - replaced with getRolePermissionsFiltered
// This method did not properly filter permissions by user type
func (s *userService) getRolePermissionsWithCache(ctx context.Context, role *models.Role) ([]string, error) {
	// Admin gets all permissions
	if role.Name == models.RoleNameAdmin && role.IsSystem {
		return models.DefaultAdminPermissions(), nil
	}

	// Try cache first (if Redis available)
	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("role:permissions:%s", role.ID.String())
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			// Parse cached permissions (JSON array)
			var permissions []string
			if err := json.Unmarshal([]byte(cached), &permissions); err == nil {
				return permissions, nil
			}
		}
	}

	// Fetch from DB
	perms, err := s.repo.Permission().GetRolePermissions(ctx, role.ID)
	if err != nil {
		return nil, err
	}

	permissions := make([]string, len(perms))
	for i, p := range perms {
		permissions[i] = p.Name
	}

	// Cache for 5 minutes (if Redis available)
	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("role:permissions:%s", role.ID.String())
		if jsonPerms, err := json.Marshal(permissions); err == nil {
			_ = s.redisClient.Set(ctx, cacheKey, string(jsonPerms), 5*time.Minute).Err()
		}
	}

	return permissions, nil
}

// persistRefreshToken hashes and stores a refresh token tied to a session & org
func (s *userService) persistRefreshToken(ctx context.Context, refreshID string, refreshToken string, userID, organizationID, sessionID uuid.UUID) error {
	if refreshID == "" {
		return errors.New("refresh token ID is required")
	}

	refreshUUID, err := uuid.Parse(refreshID)
	if err != nil {
		return fmt.Errorf("invalid refresh token id: %w", err)
	}

	// Use SHA256 for hashing refresh tokens (they're too long for password hashing)
	tokenHash := hashToken(refreshToken)

	claims, err := s.jwtService.ParseRefreshToken(refreshToken)
	if err != nil {
		return fmt.Errorf("failed to parse refresh token claims: %w", err)
	}

	record := &models.RefreshToken{
		ID:             refreshUUID,
		UserID:         userID,
		OrganizationID: organizationID,
		SessionID:      sessionID,
		TokenHash:      tokenHash,
		ExpiresAt:      claims.ExpiresAt.Time,
	}

	return s.repo.RefreshToken().Create(ctx, record)
}

// convertToUserProfile maps model to DTO
func (s *userService) convertToUserProfile(user *models.User) *UserProfile {
	p := &UserProfile{
		ID:           user.ID.String(),
		Email:        user.Email,
		FirstName:    safeStringDereference(user.Firstname),
		LastName:     safeStringDereference(user.Lastname),
		Phone:        safeStringDereference(user.Phone),
		Address:      safeStringDereference(user.Address),
		GlobalRole:   user.GlobalRole,
		IsSuperadmin: user.IsSuperadmin,
		IsActive:     user.Status == models.UserStatusActive,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	if user.LastLoginAt != nil {
		p.LastLogin = user.LastLoginAt
	}

	return p
}

// generateCryptographicallySecureToken generates a secure random token string
func generateCryptographicallySecureToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback (should not happen normally)
		return uuid.New().String()
	}
	return hex.EncodeToString(b)
}

func safeStringDereference(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func safeStringToPointer(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

// hashPasswordAsync hashes a password asynchronously
type PasswordHashResult struct {
	Hash  string
	Error error
}

func (s *userService) hashPasswordAsync(pw string) <-chan PasswordHashResult {
	ch := make(chan PasswordHashResult, 1)
	go func() {
		defer close(ch)
		hash, err := s.passwordService.Hash(pw)
		ch <- PasswordHashResult{Hash: hash, Error: err}
	}()
	return ch
}

// ───────────────────────────────────────────────────────────────────────────────
// ACCOUNT LOCKOUT (GLOBAL, EMAIL ONLY - prevents IP-based lockout abuse)
// ───────────────────────────────────────────────────────────────────────────────

func (s *userService) isAccountLocked(ctx context.Context, email, ipAddress string) bool {
	if s.redisClient == nil {
		return s.isAccountLockedDB(ctx, email)
	}
	// Use email-only key to prevent attacker from locking out user with different IPs
	lockKey := fmt.Sprintf("lockout:%s", email)
	exists, err := s.redisClient.Exists(ctx, lockKey).Result()
	return err == nil && exists > 0
}

func (s *userService) isAccountLockedDB(ctx context.Context, email string) bool {
	since := time.Now().Add(-FailedAttemptWindow)
	// Count all failed attempts for this email (regardless of IP)
	count, err := s.repo.FailedLoginAttempt().CountByEmailAndIP(ctx, email, "", "", since)
	return err == nil && count >= MaxFailedAttempts
}

func (s *userService) recordFailedAttempt(ctx context.Context, email string, ipAddress string, userID *uuid.UUID) {
	// Normalize email for consistent storage and lookup
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	attempt := &models.FailedLoginAttempt{
		UserID:      userID,
		Email:       normalizedEmail,
		IPAddress:   ipAddress,
		UserAgent:   getUserAgent(ctx),
		AttemptedAt: time.Now(),
	}

	if err := s.repo.FailedLoginAttempt().Create(ctx, attempt); err != nil {
		fmt.Printf("Failed to record failed attempt: %v\n", err)
	}

	since := time.Now().Add(-FailedAttemptWindow)
	// Count by email only (not IP - prevents lockout abuse)
	count, err := s.repo.FailedLoginAttempt().CountByEmailAndIP(ctx, normalizedEmail, "", "", since)
	if err == nil && count >= MaxFailedAttempts {
		s.lockAccount(ctx, normalizedEmail)
	}
}

func (s *userService) lockAccount(ctx context.Context, email string) {
	if s.redisClient == nil {
		fmt.Printf("Account locked for email %s\n", email)
		return
	}
	// Use email-only key (prevents IP-based lockout abuse)
	lockKey := fmt.Sprintf("lockout:%s", email)
	if err := s.redisClient.Set(ctx, lockKey, "locked", LockoutDuration).Err(); err != nil {
		fmt.Printf("Failed to set account lockout in Redis: %v\n", err)
	}
}

func (s *userService) clearFailedAttempts(ctx context.Context, email, ipAddress string) {
	if s.redisClient != nil {
		lockKey := fmt.Sprintf("lockout:%s", email)
		_ = s.redisClient.Del(ctx, lockKey)
	}
}

// ───────────────────────────────────────────────────────────────────────────────
// EMAIL VERIFICATION RATE LIMITING (Redis-based)
// ───────────────────────────────────────────────────────────────────────────────

// isVerificationRateLimited checks if email has exceeded verification attempts
func (s *userService) isVerificationRateLimited(ctx context.Context, email string) bool {
	if s.redisClient == nil {
		return false // No rate limiting if Redis unavailable
	}

	lockKey := fmt.Sprintf("verify_lockout:%s", email)

	// Check if account is locked out
	exists, err := s.redisClient.Exists(ctx, lockKey).Result()
	if err == nil && exists > 0 {
		return true
	}

	// Check attempt count in current window
	attemptKey := fmt.Sprintf("verify_attempts:%s", email)
	count, err := s.redisClient.Get(ctx, attemptKey).Int()
	if err == nil && count >= MaxVerificationAttempts {
		// Lock out the account
		_ = s.redisClient.Set(ctx, lockKey, "locked", VerificationLockoutDuration).Err()
		return true
	}

	return false
}

// recordVerificationAttempt increments the verification attempt counter
func (s *userService) recordVerificationAttempt(ctx context.Context, email string) {
	if s.redisClient == nil {
		return
	}

	attemptKey := fmt.Sprintf("verify_attempts:%s", email)

	// Increment counter
	count, err := s.redisClient.Incr(ctx, attemptKey).Result()
	if err != nil {
		fmt.Printf("Failed to record verification attempt: %v\n", err)
		return
	}

	// Set expiry on first attempt
	if count == 1 {
		_ = s.redisClient.Expire(ctx, attemptKey, VerificationAttemptWindow).Err()
	}

	// Log attempts for monitoring
	fmt.Printf("🔐 Verification attempt %d/%d for email: %s\n", count, MaxVerificationAttempts, email)

	// If max attempts reached, lock account
	if count >= MaxVerificationAttempts {
		lockKey := fmt.Sprintf("verify_lockout:%s", email)
		_ = s.redisClient.Set(ctx, lockKey, "locked", VerificationLockoutDuration).Err()
		fmt.Printf("⚠️  Email verification locked for %s (too many attempts)\n", email)
	}
}

// clearVerificationAttempts clears attempt counter on successful verification
func (s *userService) clearVerificationAttempts(ctx context.Context, email string) {
	if s.redisClient == nil {
		return
	}

	attemptKey := fmt.Sprintf("verify_attempts:%s", email)
	lockKey := fmt.Sprintf("verify_lockout:%s", email)

	_ = s.redisClient.Del(ctx, attemptKey)
	_ = s.redisClient.Del(ctx, lockKey)
}

// ───────────────────────────────────────────────────────────────────────────────
// CONTEXT HELPERS (IP & UA) – adjust based on your middleware
// ───────────────────────────────────────────────────────────────────────────────

func getClientIP(ctx context.Context) string {
	if v, ok := ctx.Value("client_ip").(string); ok {
		return v
	}
	return ""
}

func getUserAgent(ctx context.Context) string {
	if v, ok := ctx.Value("user_agent").(string); ok {
		return v
	}
	return ""
}
