package repository

import (
	"context"
	"errors"
	"time"

	"auth-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthorizationCodeRepository defines methods for authorization code data access
type AuthorizationCodeRepository interface {
	Create(ctx context.Context, code *models.AuthorizationCode) error
	GetByCode(ctx context.Context, code string) (*models.AuthorizationCode, error)
	MarkAsUsed(ctx context.Context, code string) error
	DeleteExpired(ctx context.Context) error
}

type authorizationCodeRepository struct {
	db *gorm.DB
}

// NewAuthorizationCodeRepository creates a new AuthorizationCodeRepository
func NewAuthorizationCodeRepository(db *gorm.DB) AuthorizationCodeRepository {
	return &authorizationCodeRepository{db: db}
}

func (r *authorizationCodeRepository) Create(ctx context.Context, code *models.AuthorizationCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

func (r *authorizationCodeRepository) GetByCode(ctx context.Context, code string) (*models.AuthorizationCode, error) {
	var authCode models.AuthorizationCode
	err := r.db.WithContext(ctx).Where("code = ? AND used = false AND expires_at > ?", code, time.Now()).First(&authCode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("authorization code not found or expired")
		}
		return nil, err
	}
	return &authCode, nil
}

func (r *authorizationCodeRepository) MarkAsUsed(ctx context.Context, code string) error {
	return r.db.WithContext(ctx).Model(&models.AuthorizationCode{}).
		Where("code = ?", code).
		Update("used", true).Error
}

func (r *authorizationCodeRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ? OR used = true", time.Now().Add(-1*time.Hour)).
		Delete(&models.AuthorizationCode{}).Error
}

// OAuthRefreshTokenRepository defines methods for OAuth refresh token data access
type OAuthRefreshTokenRepository interface {
	Create(ctx context.Context, token *models.OAuthRefreshToken) error
	GetByToken(ctx context.Context, tokenHash string) (*models.OAuthRefreshToken, error)
	GetByUserAndClient(ctx context.Context, userID uuid.UUID, clientID string) ([]*models.OAuthRefreshToken, error)
	Revoke(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID, clientID string) error
	DeleteExpired(ctx context.Context) error
}

type oauthRefreshTokenRepository struct {
	db *gorm.DB
}

// NewOAuthRefreshTokenRepository creates a new OAuthRefreshTokenRepository
func NewOAuthRefreshTokenRepository(db *gorm.DB) OAuthRefreshTokenRepository {
	return &oauthRefreshTokenRepository{db: db}
}

func (r *oauthRefreshTokenRepository) Create(ctx context.Context, token *models.OAuthRefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *oauthRefreshTokenRepository) GetByToken(ctx context.Context, tokenHash string) (*models.OAuthRefreshToken, error) {
	var token models.OAuthRefreshToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND revoked = false AND expires_at > ?", tokenHash, time.Now()).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refresh token not found, revoked, or expired")
		}
		return nil, err
	}
	return &token, nil
}

func (r *oauthRefreshTokenRepository) GetByUserAndClient(ctx context.Context, userID uuid.UUID, clientID string) ([]*models.OAuthRefreshToken, error) {
	var tokens []*models.OAuthRefreshToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND client_id = ? AND revoked = false AND expires_at > ?", userID, clientID, time.Now()).
		Find(&tokens).Error
	return tokens, err
}

func (r *oauthRefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).Model(&models.OAuthRefreshToken{}).
		Where("token = ?", tokenHash).
		Update("revoked", true).Error
}

func (r *oauthRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID, clientID string) error {
	return r.db.WithContext(ctx).Model(&models.OAuthRefreshToken{}).
		Where("user_id = ? AND client_id = ?", userID, clientID).
		Update("revoked", true).Error
}

func (r *oauthRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ? OR revoked = true", time.Now().Add(-30*24*time.Hour)).
		Delete(&models.OAuthRefreshToken{}).Error
}
