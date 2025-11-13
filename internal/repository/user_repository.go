package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"auth-service/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUserID     = errors.New("invalid user ID")
)

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	// Check if user with same email and type already exists in tenant
	var existingUser models.User
	result := r.db.WithContext(ctx).Where("email = ? AND user_type = ? AND tenant_id = ?", user.Email, user.UserType, user.TenantID).First(&existingUser)
	if result.Error == nil {
		return ErrUserAlreadyExists
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	if id == "" {
		return nil, ErrInvalidUserID
	}

	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// GetByEmail retrieves a user by email and tenant ID
func (r *userRepository) GetByEmail(ctx context.Context, email, tenantID string) (*models.User, error) {
	if email == "" || tenantID == "" {
		return nil, errors.New("email and tenant ID are required")
	}

	var user models.User
	err := r.db.WithContext(ctx).Where("email = ? AND tenant_id = ?", email, tenantID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// GetByEmailAndType retrieves a user by email, type, and tenant ID
func (r *userRepository) GetByEmailAndType(ctx context.Context, email, userType, tenantID string) (*models.User, error) {
	if email == "" || userType == "" || tenantID == "" {
		return nil, errors.New("email, user type, and tenant ID are required")
	}

	var user models.User
	err := r.db.WithContext(ctx).Where("email = ? AND user_type = ? AND tenant_id = ?", email, userType, tenantID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	if user == nil || user.ID == uuid.Nil {
		return errors.New("user and user ID are required")
	}

	result := r.db.WithContext(ctx).Model(user).Updates(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidUserID
	}

	result := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// List retrieves users with cursor-based pagination
func (r *userRepository) List(ctx context.Context, tenantID string, limit int, cursor string) ([]*models.User, error) {
	if tenantID == "" {
		return nil, errors.New("tenant ID is required")
	}

	var users []*models.User
	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	// Apply cursor condition if provided
	if cursor != "" {
		query = query.Where("id > ?", cursor)
	}

	// Order by created_at DESC, then by ID DESC for consistent pagination
	query = query.Order("created_at DESC, id DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&users).Error
	return users, err
}

// Count counts users in a tenant
func (r *userRepository) Count(ctx context.Context, tenantID string) (int64, error) {
	if tenantID == "" {
		return 0, errors.New("tenant ID is required")
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

// UpdateLastLogin updates the last login timestamp
func (r *userRepository) UpdateLastLogin(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidUserID
	}

	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("last_login_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdatePassword updates the user's password
func (r *userRepository) UpdatePassword(ctx context.Context, id, hashedPassword string) error {
	if id == "" || hashedPassword == "" {
		return errors.New("user ID and password are required")
	}

	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("password", hashedPassword)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// Activate activates a user account
func (r *userRepository) Activate(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidUserID
	}

	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("is_active", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// Deactivate deactivates a user account
func (r *userRepository) Deactivate(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidUserID
	}

	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}