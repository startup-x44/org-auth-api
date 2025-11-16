package repository

import (
	"context"
	"errors"

	"auth-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ClientAppRepository defines methods for client app data access
type ClientAppRepository interface {
	Create(ctx context.Context, clientApp *models.ClientApp) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.ClientApp, error)
	GetByClientID(ctx context.Context, clientID string) (*models.ClientApp, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.ClientApp, int64, error)
	Update(ctx context.Context, clientApp *models.ClientApp) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type clientAppRepository struct {
	db *gorm.DB
}

// NewClientAppRepository creates a new ClientAppRepository
func NewClientAppRepository(db *gorm.DB) ClientAppRepository {
	return &clientAppRepository{db: db}
}

func (r *clientAppRepository) Create(ctx context.Context, clientApp *models.ClientApp) error {
	return r.db.WithContext(ctx).Create(clientApp).Error
}

func (r *clientAppRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ClientApp, error) {
	var clientApp models.ClientApp
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&clientApp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("client app not found")
		}
		return nil, err
	}
	return &clientApp, nil
}

func (r *clientAppRepository) GetByClientID(ctx context.Context, clientID string) (*models.ClientApp, error) {
	var clientApp models.ClientApp
	err := r.db.WithContext(ctx).Where("client_id = ?", clientID).First(&clientApp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("client app not found")
		}
		return nil, err
	}
	return &clientApp, nil
}

func (r *clientAppRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.ClientApp, int64, error) {
	var clientApps []*models.ClientApp
	var total int64

	if err := r.db.WithContext(ctx).Model(&models.ClientApp{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := r.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&clientApps).Error
	return clientApps, total, err
}

func (r *clientAppRepository) Update(ctx context.Context, clientApp *models.ClientApp) error {
	return r.db.WithContext(ctx).Save(clientApp).Error
}

func (r *clientAppRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.ClientApp{}, "id = ?", id).Error
}
