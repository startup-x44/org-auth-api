package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/jwt"
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
	ListUsers(ctx context.Context, tenantID string, limit, offset int) (*UserListResponse, error)
	ActivateUser(ctx context.Context, userID string) error
	DeactivateUser(ctx context.Context, userID string) error
	DeleteUser(ctx context.Context, userID string) error
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
	Users  []*UserProfile `json:"users"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
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
	repo         repository.Repository
	jwtService   jwt.Service
	passwordSvc  password.Service
}

// NewUserService creates a new user service
func NewUserService(repo repository.Repository, jwtService jwt.Service, passwordSvc password.Service) UserService {
	return &userService{
		repo:        repo,
		jwtService:  jwtService,
		passwordSvc: passwordSvc,
	}
}

// Register registers a new user
func (s *userService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// Validate request
	if err := validation.ValidateUserRegistration(req.Email, req.Password, req.ConfirmPassword, req.UserType, req.TenantID); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
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

	// Check if tenant exists
	_, err := s.repo.Tenant().GetByID(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant: %w", err)
	}

	// Hash password
	hashedPassword, err := s.passwordSvc.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:     req.Email,
		Password:  hashedPassword,
		Type:      req.UserType,
		TenantID:  req.TenantID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		IsActive:  true, // Users are active by default
	}

	if err := s.repo.User().Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := &models.UserSession{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Session expires in 24 hours
	}
	if err := s.repo.UserSession().Create(ctx, session); err != nil {
		// Log error but don't fail registration
		fmt.Printf("Failed to create session: %v\n", err)
	}

	// Create refresh token
	refreshToken := &models.RefreshToken{
		UserID:    user.ID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Refresh token expires in 7 days
	}
	if err := s.repo.RefreshToken().Create(ctx, refreshToken); err != nil {
		// Log error but don't fail registration
		fmt.Printf("Failed to create refresh token: %v\n", err)
	}

	return &RegisterResponse{
		User:  s.convertToUserProfile(user),
		Token: tokenPair,
	}, nil
}

// Login authenticates a user
func (s *userService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Validate request
	if err := validation.ValidateLogin(req.Email, req.TenantID); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get user
	var user *models.User
	var err error
	if req.UserType != "" {
		user, err = s.repo.User().GetByEmailAndType(ctx, req.Email, req.UserType, req.TenantID)
	} else {
		user, err = s.repo.User().GetByEmail(ctx, req.Email, req.TenantID)
	}
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	if err := s.passwordSvc.Verify(req.Password, user.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Update last login
	if err := s.repo.User().UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail login
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	// Generate tokens
	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := &models.UserSession{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := s.repo.UserSession().Create(ctx, session); err != nil {
		fmt.Printf("Failed to create session: %v\n", err)
	}

	// Create refresh token
	refreshToken := &models.RefreshToken{
		UserID:    user.ID,
		Token:     tokenPair.RefreshToken,
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
	user, err := s.repo.User().GetByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Generate new token pair
	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update refresh token
	refreshToken.Token = tokenPair.RefreshToken
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
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
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
	if err := s.passwordSvc.Verify(req.CurrentPassword, user.Password); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := s.passwordSvc.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.repo.User().UpdatePassword(ctx, userID, hashedPassword); err != nil {
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

	// Check if user exists
	user, err := s.repo.User().GetByEmail(ctx, req.Email, req.TenantID)
	if err != nil {
		// Don't reveal if user exists or not for security
		return nil
	}

	// Check if user is active
	if !user.IsActive {
		return nil // Don't reveal deactivated accounts
	}

	// Generate reset token
	resetToken := generateSecureToken()

	// Create password reset record
	reset := &models.PasswordReset{
		Email:     req.Email,
		TenantID:  req.TenantID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Token expires in 1 hour
	}

	if err := s.repo.PasswordReset().Create(ctx, reset); err != nil {
		return fmt.Errorf("failed to create password reset: %w", err)
	}

	// TODO: Send email with reset token
	// This would integrate with an email service

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
	user, err := s.repo.User().GetByEmail(ctx, reset.Email, reset.TenantID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Hash new password
	hashedPassword, err := s.passwordSvc.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.repo.User().UpdatePassword(ctx, user.ID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete the reset token
	if err := s.repo.PasswordReset().Delete(ctx, req.Token); err != nil {
		fmt.Printf("Failed to delete reset token: %v\n", err)
	}

	// Invalidate all refresh tokens for security
	if err := s.repo.RefreshToken().DeleteByUserID(ctx, user.ID); err != nil {
		fmt.Printf("Failed to invalidate refresh tokens: %v\n", err)
	}

	return nil
}

// ListUsers lists users with pagination
func (s *userService) ListUsers(ctx context.Context, tenantID string, limit, offset int) (*UserListResponse, error) {
	if tenantID == "" {
		return nil, errors.New("tenant ID is required")
	}

	users, err := s.repo.User().List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.repo.User().Count(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	userProfiles := make([]*UserProfile, len(users))
	for i, user := range users {
		userProfiles[i] = s.convertToUserProfile(user)
	}

	return &UserListResponse{
		Users:  userProfiles,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// ActivateUser activates a user account
func (s *userService) ActivateUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	return s.repo.User().Activate(ctx, userID)
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

	return s.repo.User().Deactivate(ctx, userID)
}

// DeleteUser deletes a user account
func (s *userService) DeleteUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	// Start transaction
	tx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete all related data
	if err := tx.UserSession().DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}
	if err := tx.RefreshToken().DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete refresh tokens: %w", err)
	}
	if err := tx.PasswordReset().DeleteExpired(ctx); err != nil {
		// This is a cleanup, don't fail if it errors
		fmt.Printf("Failed to cleanup password resets: %v\n", err)
	}

	// Delete user
	if err := tx.User().Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Helper methods

// generateTokenPair generates access and refresh tokens
func (s *userService) generateTokenPair(user *models.User) (*TokenPair, error) {
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Type, user.TenantID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, user.Type, user.TenantID)
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
		ID:        user.ID,
		Email:     user.Email,
		UserType:  user.Type,
		TenantID:  user.TenantID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if user.LastLoginAt != nil {
		profile.LastLogin = user.LastLoginAt
	}

	return profile
}

// generateSecureToken generates a secure random token
func generateSecureToken() string {
	// TODO: Implement secure token generation
	// For now, return a placeholder
	return "secure-reset-token-placeholder"
}