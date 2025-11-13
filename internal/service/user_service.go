package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/email"
	"auth-service/pkg/jwt"
	"auth-service/pkg/logger"
	"auth-service/pkg/password"
	"auth-service/pkg/validation"
)

// UserService defines the interface for user business logic
type UserService interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error)
	Logout(ctx context.Context, req *LogoutRequest) error
	GetProfile(ctx context.Context, userID string) (*UserProfile, error)
	UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*UserProfile, error)
	ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
	ListUsers(ctx context.Context, tenantID string, limit int, cursor string) (*UserListResponse, error)
	ActivateUser(ctx context.Context, userID string) error
	DeactivateUser(ctx context.Context, userID string) error
	DeleteUser(ctx context.Context, userID string) error
	SetRedisClient(client *redis.Client)
	SetEmailService(emailSvc email.Service)
	SetSessionService(sessionSvc SessionService)
	GetUserSessions(ctx context.Context, userID string) ([]*models.UserSession, error)
	RevokeUserSession(ctx context.Context, userID, sessionID, reason string) error
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	UserType        string `json:"user_type"`
	TenantID        string `json:"tenant_id"`
	FirstName       string `json:"first_name,omitempty"`
	LastName        string `json:"last_name,omitempty"`
	Phone           string `json:"phone,omitempty"`
}

// RegisterResponse represents user registration response
type RegisterResponse struct {
	User  *UserProfile `json:"user"`
	Token *TokenPair   `json:"token"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TenantID string `json:"tenant_id"`
	UserType string `json:"user_type,omitempty"`
}

// LoginResponse represents user login response
type LoginResponse struct {
	User  *UserProfile `json:"user"`
	Token *TokenPair   `json:"token"`
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenResponse represents token refresh response
type RefreshTokenResponse struct {
	Token *TokenPair `json:"token"`
}

// LogoutRequest represents logout request
type LogoutRequest struct {
	UserID       string `json:"user_id"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// UserProfile represents user profile information
type UserProfile struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	UserType  string    `json:"user_type"`
	TenantID  string    `json:"tenant_id"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
	Email    string `json:"email"`
	TenantID string `json:"tenant_id"`
}

// ResetPasswordRequest represents password reset request
type ResetPasswordRequest struct {
	Token           string `json:"token"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

// UserListResponse represents paginated user list response
type UserListResponse struct {
	Users      []*UserProfile `json:"users"`
	Total      int64          `json:"total"`
	Limit      int            `json:"limit"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// userService implements UserService interface
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
		sessionSvc:      nil, // Will be set later
		auditLogger:     logger.NewAuditLogger(),
	}
}

// SetRedisClient sets the Redis client for account lockout functionality
func (s *userService) SetRedisClient(client *redis.Client) {
	s.redisClient = client
}

// SetEmailService sets the email service for password reset functionality
func (s *userService) SetEmailService(emailSvc email.Service) {
	s.emailSvc = emailSvc
}

// SetSessionService sets the session service for advanced session management
func (s *userService) SetSessionService(sessionSvc SessionService) {
	s.sessionSvc = sessionSvc
}

// Register registers a new user
func (s *userService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// Validate registration request
	if err := s.validateRegistrationRequest(req); err != nil {
		return nil, err
	}

	// Resolve tenant for user
	tenantID, err := s.resolveTenantForRegistration(ctx, req)
	if err != nil {
		return nil, err
	}

	// Create user account
	user, err := s.createUserAccount(ctx, req, tenantID)
	if err != nil {
		return nil, err
	}

	// Generate authentication tokens
	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// Create user session and refresh token
	if err := s.createUserSessionAndTokens(ctx, user, tokenPair); err != nil {
		// Log error but don't fail registration
		fmt.Printf("Failed to create session/tokens: %v\n", err)
	}

	return &RegisterResponse{
		User:  s.convertToUserProfile(user),
		Token: tokenPair,
	}, nil
}

// validateRegistrationRequest validates the registration request data
func (s *userService) validateRegistrationRequest(req *RegisterRequest) error {
	// Validate core registration fields
	if err := validation.ValidateUserRegistration(req.Email, req.Password, req.ConfirmPassword, req.UserType, req.TenantID); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Validate optional profile fields
	if err := s.validateOptionalProfileFields(req); err != nil {
		return err
	}

	return nil
}

// validateOptionalProfileFields validates optional user profile fields
func (s *userService) validateOptionalProfileFields(req *RegisterRequest) error {
	if req.FirstName != "" {
		if err := validation.ValidateName(req.FirstName, "first name"); err != nil {
			return err
		}
	}
	if req.LastName != "" {
		if err := validation.ValidateName(req.LastName, "last name"); err != nil {
			return err
		}
	}
	if req.Phone != "" {
		if err := validation.ValidatePhone(req.Phone); err != nil {
			return err
		}
	}
	return nil
}

// resolveTenantForRegistration determines the tenant for the new user
func (s *userService) resolveTenantForRegistration(ctx context.Context, req *RegisterRequest) (string, error) {
	// If tenant ID is provided, validate it
	if req.TenantID != "" {
		return s.validateProvidedTenantID(ctx, req.TenantID)
	}

	// Auto-assign tenant based on email domain
	return s.autoAssignTenantByEmailDomain(ctx, req.Email)
}

// validateProvidedTenantID validates a tenant ID provided in the request
func (s *userService) validateProvidedTenantID(ctx context.Context, tenantID string) (string, error) {
	_, err := s.repo.Tenant().GetByID(ctx, tenantID)
	// _, err := s.repo.Tenant().GetByID(ctx, tenantID)
	if err != nil {
		return "", fmt.Errorf("invalid tenant: %w", err)
	}
	return tenantID, nil
}

// autoAssignTenantByEmailDomain automatically assigns tenant based on email domain
func (s *userService) autoAssignTenantByEmailDomain(ctx context.Context, email string) (string, error) {
	emailDomain := extractDomainFromEmail(email)
	if emailDomain != "" {
		// Try to find tenant by domain
		tenant, err := s.repo.Tenant().GetByDomain(ctx, emailDomain)
		if err == nil {
			return tenant.ID.String(), nil
		}
	}

	// Fall back to default tenant
	defaultTenant, err := s.getOrCreateDefaultTenant(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get default tenant: %w", err)
	}
	return defaultTenant.ID.String(), nil
}

// createUserAccount creates the user account with hashed password
func (s *userService) createUserAccount(ctx context.Context, req *RegisterRequest, tenantID string) (*models.User, error) {
	// Hash password asynchronously for better performance
	passwordHashChan := s.hashPasswordAsync(req.Password)
	passwordHashResult := <-passwordHashChan
	if passwordHashResult.Error != nil {
		return nil, fmt.Errorf("failed to hash password: %w", passwordHashResult.Error)
	}

	// Create user model
	user := &models.User{
		Email:        req.Email,
		PasswordHash: passwordHashResult.Hash,
		UserType:     req.UserType,
		TenantID:     uuid.MustParse(tenantID),
		Firstname:    safeStringToPointer(req.FirstName),
		Lastname:     safeStringToPointer(req.LastName),
		Phone:        safeStringToPointer(req.Phone),
	}

	// Save user to database
	if err := s.repo.User().Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// createUserSessionAndTokens creates session and refresh token for the user
func (s *userService) createUserSessionAndTokens(ctx context.Context, user *models.User, tokenPair *TokenPair) error {
	// Create session
	session := &models.UserSession{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		TokenHash: generateCryptographicallySecureToken(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := s.repo.UserSession().Create(ctx, session); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Create refresh token
	refreshToken := &models.RefreshToken{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		TokenHash: tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.repo.RefreshToken().Create(ctx, refreshToken); err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// Login authenticates a user
func (s *userService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Validate request
	if err := validation.ValidateLogin(req.Email, req.TenantID); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Resolve tenant ID (could be domain or UUID)
	resolvedTenantID, err := s.resolveTenantID(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant: %w", err)
	}

	// Check if account is locked
	clientIP := getClientIP(ctx)
	if s.isAccountLocked(ctx, req.Email, clientIP, resolvedTenantID) {
		return nil, errors.New("account is temporarily locked due to too many failed login attempts")
	}

	// Get user
	var user *models.User
	if req.UserType != "" {
		user, err = s.repo.User().GetByEmailAndType(ctx, req.Email, req.UserType, resolvedTenantID)
	} else {
		user, err = s.repo.User().GetByEmail(ctx, req.Email, resolvedTenantID)
	}
	if err != nil {
		// Record failed attempt for non-existent user
		s.recordFailedAttempt(ctx, req.Email, clientIP, resolvedTenantID, nil)
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	// Check if user is active
	if user.Status != models.UserStatusActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	valid, err := s.passwordService.Verify(req.Password, user.PasswordHash)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if !valid {
		// Record failed attempt
		s.recordFailedAttempt(ctx, req.Email, clientIP, resolvedTenantID, &user.ID)
		return nil, errors.New("invalid credentials")
	}

	// Clear failed attempts on successful login
	s.clearFailedAttempts(ctx, req.Email, clientIP, resolvedTenantID)

	// Update last login
	if err := s.repo.User().UpdateLastLogin(ctx, user.ID.String()); err != nil {
		// Log error but don't fail login
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	// Generate tokens
	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Enforce session limits if session service is available
	if s.sessionSvc != nil {
		maxSessions := 5 // Default max sessions per user
		if err := s.sessionSvc.EnforceSessionLimits(ctx, user.ID.String(), maxSessions); err != nil {
			return nil, fmt.Errorf("failed to enforce session limits: %w", err)
		}
	}

	// Create session using session service if available
	var session *models.UserSession
	if s.sessionSvc != nil {
		session, err = s.sessionSvc.CreateSession(ctx, user.ID.String(), resolvedTenantID, clientIP, getUserAgent(ctx))
		if err != nil {
			fmt.Printf("Failed to create session: %v\n", err)
		}

		// Check for suspicious activity
		if session != nil {
			if suspicious, reasons := s.sessionSvc.DetectSuspiciousActivity(ctx, session); suspicious {
				s.auditLogger.LogSecurityEvent("suspicious_login_detected", req.Email, clientIP, false, nil, fmt.Sprintf("Suspicious activity detected: %s", reasons))
				// Optionally revoke the session or require additional verification
			}
		}
	} else {
		// Fallback to old session creation logic
		session = &models.UserSession{
			UserID:    user.ID,
			TenantID:  user.TenantID,
			TokenHash: generateCryptographicallySecureToken(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		if err := s.repo.UserSession().Create(ctx, session); err != nil {
			fmt.Printf("Failed to create session: %v\n", err)
		}
	}

	// Create refresh token
	refreshToken := &models.RefreshToken{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		TokenHash: tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.repo.RefreshToken().Create(ctx, refreshToken); err != nil {
		fmt.Printf("Failed to create refresh token: %v\n", err)
	}

	return &LoginResponse{
		User:  s.convertToUserProfile(user),
		Token: tokenPair,
	}, nil
}

// RefreshToken refreshes access token using refresh token
func (s *userService) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, errors.New("refresh token is required")
	}

	// Get refresh token from database
	refreshToken, err := s.repo.RefreshToken().GetByToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if refresh token is expired
	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	// Get user
	user, err := s.repo.User().GetByID(ctx, refreshToken.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user is active
	if user.Status != models.UserStatusActive {
		return nil, errors.New("account is deactivated")
	}

	// Generate new token pair
	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update refresh token
	refreshToken.TokenHash = tokenPair.RefreshToken
	refreshToken.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	if err := s.repo.RefreshToken().Update(ctx, refreshToken); err != nil {
		fmt.Printf("Failed to update refresh token: %v\n", err)
	}

	return &RefreshTokenResponse{
		Token: tokenPair,
	}, nil
}

// Logout logs out a user
func (s *userService) Logout(ctx context.Context, req *LogoutRequest) error {
	if req.UserID == "" {
		return errors.New("user ID is required")
	}

	// Delete refresh token if provided
	if req.RefreshToken != "" {
		if err := s.repo.RefreshToken().Delete(ctx, req.RefreshToken); err != nil {
			fmt.Printf("Failed to delete refresh token: %v\n", err)
		}
	}

	// Delete all sessions for user (optional - could be selective)
	if err := s.repo.UserSession().DeleteByUserID(ctx, req.UserID); err != nil {
		fmt.Printf("Failed to delete user sessions: %v\n", err)
	}

	return nil
}

// GetProfile gets user profile
func (s *userService) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	user, err := s.repo.User().GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return s.convertToUserProfile(user), nil
}

// UpdateProfile updates user profile
func (s *userService) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*UserProfile, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	// Validate optional fields
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

	// Get current user
	user, err := s.repo.User().GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields
	if req.FirstName != "" {
		user.Firstname = &req.FirstName
	}
	if req.LastName != "" {
		user.Lastname = &req.LastName
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}

	// Save changes
	if err := s.repo.User().Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return s.convertToUserProfile(user), nil
}

// ChangePassword changes user password
func (s *userService) ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	// Validate new password
	if err := validation.ValidatePassword(req.NewPassword); err != nil {
		return fmt.Errorf("invalid new password: %w", err)
	}
	if err := validation.ValidatePasswordsMatch(req.NewPassword, req.ConfirmPassword); err != nil {
		return err
	}

	// Get user
	user, err := s.repo.User().GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify current password
	valid, err := s.passwordService.Verify(req.CurrentPassword, user.PasswordHash)
	if err != nil {
		return errors.New("current password is incorrect")
	}
	if !valid {
		return errors.New("current password is incorrect")
	}

	// Hash new password asynchronously for better performance
	passwordHashChan := s.hashPasswordAsync(req.NewPassword)
	passwordHashResult := <-passwordHashChan
	if passwordHashResult.Error != nil {
		return fmt.Errorf("failed to hash password: %w", passwordHashResult.Error)
	}

	// Update password
	if err := s.repo.User().UpdatePassword(ctx, userID, passwordHashResult.Hash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all refresh tokens for security
	if err := s.repo.RefreshToken().DeleteByUserID(ctx, userID); err != nil {
		fmt.Printf("Failed to invalidate refresh tokens: %v\n", err)
	}

	return nil
}

// ForgotPassword initiates password reset
func (s *userService) ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error {
	if err := validation.ValidateForgotPassword(req.Email, req.TenantID); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Resolve tenant ID (could be domain or UUID)
	resolvedTenantID, err := s.resolveTenantID(ctx, req.TenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant: %w", err)
	}

	// Check if user exists
	user, err := s.repo.User().GetByEmail(ctx, req.Email, resolvedTenantID)
	if err != nil {
		// Don't reveal if user exists or not for security
		return nil
	}

	// Check if user is active
	if user.Status != models.UserStatusActive {
		return nil // Don't reveal deactivated accounts
	}

	// Generate reset token (15 minutes expiry)
	resetToken := generateCryptographicallySecureToken()

	// Create password reset record
	reset := &models.PasswordReset{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		TokenHash: resetToken,
		ExpiresAt: time.Now().Add(15 * time.Minute), // 15 minutes
	}

	if err := s.repo.PasswordReset().Create(ctx, reset); err != nil {
		return fmt.Errorf("failed to create password reset: %w", err)
	}

	// Send password reset email
	if s.emailSvc != nil {
		if err := s.emailSvc.SendPasswordResetEmail(req.Email, resetToken); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to send password reset email: %v\n", err)
		}
	} else {
		// Log for development
		fmt.Printf("Password reset token for %s: %s\n", req.Email, resetToken)
	}

	return nil
}

// ResetPassword resets user password using token
func (s *userService) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	if err := validation.ValidatePasswordReset(req.Token, req.NewPassword, req.ConfirmPassword); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get password reset record
	reset, err := s.repo.PasswordReset().GetByToken(ctx, req.Token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	// Check if token is expired
	if time.Now().After(reset.ExpiresAt) {
		return errors.New("reset token expired")
	}

	// Get user
	user, err := s.repo.User().GetByID(ctx, reset.UserID.String())
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Hash new password asynchronously for better performance
	passwordHashChan := s.hashPasswordAsync(req.NewPassword)
	passwordHashResult := <-passwordHashChan
	if passwordHashResult.Error != nil {
		return fmt.Errorf("failed to hash password: %w", passwordHashResult.Error)
	}

	// Update password
	if err := s.repo.User().UpdatePassword(ctx, user.ID.String(), passwordHashResult.Hash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete the reset token
	if err := s.repo.PasswordReset().Delete(ctx, req.Token); err != nil {
		fmt.Printf("Failed to delete reset token: %v\n", err)
	}

	// Invalidate all refresh tokens for security
	if err := s.repo.RefreshToken().DeleteByUserID(ctx, user.ID.String()); err != nil {
		fmt.Printf("Failed to invalidate refresh tokens: %v\n", err)
	}

	return nil
}

// ListUsers lists users with cursor-based pagination
func (s *userService) ListUsers(ctx context.Context, tenantID string, limit int, cursor string) (*UserListResponse, error) {
	if tenantID == "" {
		return nil, errors.New("tenant ID is required")
	}

	// Resolve tenant ID (could be domain or UUID)
	resolvedTenantID, err := s.resolveTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant: %w", err)
	}

	users, err := s.repo.User().List(ctx, resolvedTenantID, limit, cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.repo.User().Count(ctx, resolvedTenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	userProfiles := make([]*UserProfile, len(users))
	for i, user := range users {
		userProfiles[i] = s.convertToUserProfile(user)
	}

	// Determine next cursor (ID of last user if we got a full page)
	var nextCursor string
	if len(users) == limit {
		nextCursor = users[len(users)-1].ID.String()
	}

	return &UserListResponse{
		Users:      userProfiles,
		Total:      total,
		Limit:      limit,
		NextCursor: nextCursor,
	}, nil
}

// ActivateUser activates a user account
func (s *userService) ActivateUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	err := s.repo.User().Activate(ctx, userID)
	if err != nil {
		s.auditLogger.LogAdminAction("system", "activate_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), false, err, "Failed to activate user account")
		return err
	}

	s.auditLogger.LogAdminAction("system", "activate_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), true, nil, "User account activated successfully")
	return nil
}

// DeactivateUser deactivates a user account
func (s *userService) DeactivateUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	// Invalidate all sessions and refresh tokens
	if err := s.repo.UserSession().DeleteByUserID(ctx, userID); err != nil {
		fmt.Printf("Failed to delete user sessions: %v\n", err)
	}
	if err := s.repo.RefreshToken().DeleteByUserID(ctx, userID); err != nil {
		fmt.Printf("Failed to delete refresh tokens: %v\n", err)
	}

	err := s.repo.User().Deactivate(ctx, userID)
	if err != nil {
		s.auditLogger.LogAdminAction("system", "deactivate_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), false, err, "Failed to deactivate user account")
		return err
	}

	s.auditLogger.LogAdminAction("system", "deactivate_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), true, nil, "User account deactivated successfully")
	return nil
}

// DeleteUser deletes a user account
func (s *userService) DeleteUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	// Start transaction
	tx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		s.auditLogger.LogAdminAction("system", "delete_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), false, err, "Failed to start transaction for user deletion")
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete all related data
	if err := tx.UserSession().DeleteByUserID(ctx, userID); err != nil {
		s.auditLogger.LogAdminAction("system", "delete_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), false, err, "Failed to delete user sessions")
		return fmt.Errorf("failed to delete sessions: %w", err)
	}
	if err := tx.RefreshToken().DeleteByUserID(ctx, userID); err != nil {
		s.auditLogger.LogAdminAction("system", "delete_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), false, err, "Failed to delete refresh tokens")
		return fmt.Errorf("failed to delete refresh tokens: %w", err)
	}
	if err := tx.PasswordReset().DeleteExpired(ctx); err != nil {
		// This is a cleanup, don't fail if it errors
		fmt.Printf("Failed to cleanup password resets: %v\n", err)
	}

	// Delete user
	if err := tx.User().Delete(ctx, userID); err != nil {
		s.auditLogger.LogAdminAction("system", "delete_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), false, err, "Failed to delete user account")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.auditLogger.LogAdminAction("system", "delete_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), false, err, "Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.auditLogger.LogAdminAction("system", "delete_user", "user", userID, getClientIP(ctx), getUserAgent(ctx), true, nil, "User account deleted successfully")
	return nil
}

// GetUserSessions retrieves all sessions for a user
func (s *userService) GetUserSessions(ctx context.Context, userID string) ([]*models.UserSession, error) {
	if s.sessionSvc != nil {
		return s.sessionSvc.GetUserSessions(ctx, userID)
	}
	return s.repo.UserSession().GetByUserID(ctx, userID)
}

// RevokeUserSession revokes a specific user session
func (s *userService) RevokeUserSession(ctx context.Context, userID, sessionID, reason string) error {
	if userID == "" || sessionID == "" {
		return errors.New("user ID and session ID are required")
	}

	// Verify the session belongs to the user
	session, err := s.repo.UserSession().GetByID(ctx, sessionID)
	if err != nil {
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

// Helper methods

// generateTokenPair generates access and refresh tokens
func (s *userService) generateTokenPair(user *models.User) (*TokenPair, error) {
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
		TokenType:    "Bearer",
	}, nil
}

// convertToUserProfile converts model to profile
func (s *userService) convertToUserProfile(user *models.User) *UserProfile {
	profile := &UserProfile{
		ID:        user.ID.String(),
		Email:     user.Email,
		UserType:  user.UserType,
		TenantID:  user.TenantID.String(),
		FirstName: safeStringDereference(user.Firstname),
		LastName:  safeStringDereference(user.Lastname),
		Phone:     safeStringDereference(user.Phone),
		IsActive:  user.Status == models.UserStatusActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if user.LastLoginAt != nil {
		profile.LastLogin = user.LastLoginAt
	}

	return profile
}

// generateCryptographicallySecureToken generates a secure random token
func generateCryptographicallySecureToken() string {
	// TODO: Implement secure token generation
	// For now, return a placeholder
	return "secure-reset-token-placeholder"
}

// extractDomainFromEmail extracts domain from email address
func extractDomainFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// getOrCreateDefaultTenant gets or creates a default tenant
func (s *userService) getOrCreateDefaultTenant(ctx context.Context) (*models.Tenant, error) {
	// Try to find existing default tenant
	defaultTenant, err := s.repo.Tenant().GetByDomain(ctx, "default.local")
	if err == nil {
		return defaultTenant, nil
	}

	// Create default tenant if it doesn't exist
	defaultTenant = &models.Tenant{
		Name:   "Default Organization",
		Domain: "default.local",
	}

	if err := s.repo.Tenant().Create(ctx, defaultTenant); err != nil {
		return nil, fmt.Errorf("failed to create default tenant: %w", err)
	}

	return defaultTenant, nil
}

// getStringValue safely gets string value from pointer
func safeStringDereference(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// resolveTenantID resolves tenant ID from either UUID or domain
func (s *userService) resolveTenantID(ctx context.Context, tenantID string) (string, error) {
	if tenantID == "" {
		return "", errors.New("tenant ID is required")
	}

	// Try to parse as UUID first
	if _, err := uuid.Parse(tenantID); err == nil {
		// It's a valid UUID, use it directly
		return tenantID, nil
	}

	// Not a UUID, treat as domain
	tenant, err := s.repo.Tenant().GetByDomain(ctx, tenantID)
	if err != nil {
		return "", fmt.Errorf("invalid tenant: %w", err)
	}

	return tenant.ID.String(), nil
}

// Account lockout constants
const (
	MaxFailedAttempts = 5
	LockoutDuration   = 15 * time.Minute
	FailedAttemptWindow = 15 * time.Minute
)

// isAccountLocked checks if an account is currently locked
func (s *userService) isAccountLocked(ctx context.Context, email, ipAddress, tenantID string) bool {
	if s.redisClient == nil {
		// Fallback to database-based lockout
		return s.isAccountLockedDB(ctx, email, ipAddress, tenantID)
	}

	// Check Redis for lockout
	lockKey := fmt.Sprintf("lockout:%s:%s:%s", email, ipAddress, tenantID)
	exists, err := s.redisClient.Exists(ctx, lockKey).Result()
	return err == nil && exists > 0
}

// isAccountLockedDB checks account lockout using database
func (s *userService) isAccountLockedDB(ctx context.Context, email, ipAddress, tenantID string) bool {
	since := time.Now().Add(-FailedAttemptWindow)
	count, err := s.repo.FailedLoginAttempt().CountByEmailAndIP(ctx, email, ipAddress, tenantID, since)
	return err == nil && count >= MaxFailedAttempts
}

// recordFailedAttempt records a failed login attempt
func (s *userService) recordFailedAttempt(ctx context.Context, email, ipAddress, tenantID string, userID *uuid.UUID) {
	// Record in database
	attempt := &models.FailedLoginAttempt{
		UserID:     userID,
		TenantID:   uuid.MustParse(tenantID),
		Email:      email,
		IPAddress:  ipAddress,
		UserAgent:  getUserAgent(ctx),
		AttemptedAt: time.Now(),
	}

	if err := s.repo.FailedLoginAttempt().Create(ctx, attempt); err != nil {
		fmt.Printf("Failed to record failed attempt: %v\n", err)
	}

	// Check if we should lock the account
	since := time.Now().Add(-FailedAttemptWindow)
	count, err := s.repo.FailedLoginAttempt().CountByEmailAndIP(ctx, email, ipAddress, tenantID, since)
	if err == nil && count >= MaxFailedAttempts {
		s.lockAccount(ctx, email, ipAddress, tenantID)
	}
}

// lockAccount locks an account for a period of time
func (s *userService) lockAccount(ctx context.Context, email, ipAddress, tenantID string) {
	if s.redisClient == nil {
		// No Redis, just log that account would be locked
		fmt.Printf("Account locked for email %s from IP %s\n", email, ipAddress)
		return
	}

	lockKey := fmt.Sprintf("lockout:%s:%s:%s", email, ipAddress, tenantID)
	if err := s.redisClient.Set(ctx, lockKey, "locked", LockoutDuration).Err(); err != nil {
		fmt.Printf("Failed to set account lockout in Redis: %v\n", err)
	}
}

// clearFailedAttempts clears failed attempts for successful login
func (s *userService) clearFailedAttempts(ctx context.Context, email, ipAddress, tenantID string) {
	// Clear from database (cleanup will happen periodically)
	// For now, just clear Redis if available
	if s.redisClient != nil {
		lockKey := fmt.Sprintf("lockout:%s:%s:%s", email, ipAddress, tenantID)
		s.redisClient.Del(ctx, lockKey)
	}
}

// hashPasswordAsync hashes a password asynchronously for better performance
func (s *userService) hashPasswordAsync(password string) <-chan PasswordHashResult {
	resultChan := make(chan PasswordHashResult, 1)

	go func() {
		defer close(resultChan)
		hash, err := s.passwordService.Hash(password)
		resultChan <- PasswordHashResult{
			Hash:  hash,
			Error: err,
		}
	}()

	return resultChan
}

// PasswordHashResult holds the result of an asynchronous password hash operation
type PasswordHashResult struct {
	Hash  string
	Error error
}

// safeStringToPointer converts a string to a pointer, returning nil for empty strings
func safeStringToPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}