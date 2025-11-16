package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/password"

	"github.com/google/uuid"
)

// ClientAppService defines client application management operations
type ClientAppService interface {
	CreateClientApp(ctx context.Context, req *CreateClientAppRequest, createdBy *models.User) (*ClientAppResponse, string, error)
	GetClientApp(ctx context.Context, id uuid.UUID, requestedBy *models.User) (*ClientAppResponse, error)
	GetClientAppByClientID(ctx context.Context, clientID string) (*models.ClientApp, error)
	ListClientApps(ctx context.Context, limit, offset int, requestedBy *models.User) ([]*ClientAppResponse, int64, error)
	UpdateClientApp(ctx context.Context, id uuid.UUID, req *UpdateClientAppRequest, updatedBy *models.User) (*ClientAppResponse, error)
	DeleteClientApp(ctx context.Context, id uuid.UUID, deletedBy *models.User) error
	RotateClientSecret(ctx context.Context, id uuid.UUID, rotatedBy *models.User) (string, error)
	ValidateClientCredentials(ctx context.Context, clientID, clientSecret string) (*models.ClientApp, error)
	ValidateRedirectURI(ctx context.Context, clientID, redirectURI string) error
}

type clientAppService struct {
	repo        repository.Repository
	passwordSvc password.PasswordService
}

// NewClientAppService creates a new client app service
func NewClientAppService(repo repository.Repository) ClientAppService {
	return &clientAppService{
		repo:        repo,
		passwordSvc: password.NewService(),
	}
}

// CreateClientAppRequest represents request to create a client app
type CreateClientAppRequest struct {
	Name           string   `json:"name" validate:"required,min=3,max=255"`
	RedirectURIs   []string `json:"redirect_uris" validate:"required,min=1,dive,url"`
	AllowedOrigins []string `json:"allowed_origins" validate:"omitempty,dive,url"`
	AllowedScopes  []string `json:"allowed_scopes" validate:"omitempty"`
	IsConfidential bool     `json:"is_confidential"` // default true
}

// UpdateClientAppRequest represents request to update a client app
type UpdateClientAppRequest struct {
	Name           string   `json:"name" validate:"omitempty,min=3,max=255"`
	RedirectURIs   []string `json:"redirect_uris" validate:"omitempty,min=1,dive,url"`
	AllowedOrigins []string `json:"allowed_origins" validate:"omitempty,dive,url"`
	AllowedScopes  []string `json:"allowed_scopes" validate:"omitempty"`
}

// ClientAppResponse represents a client app response (without secret)
type ClientAppResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	ClientID       string    `json:"client_id"`
	RedirectURIs   []string  `json:"redirect_uris"`
	AllowedOrigins []string  `json:"allowed_origins"`
	AllowedScopes  []string  `json:"allowed_scopes"`
	IsConfidential bool      `json:"is_confidential"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
}

func (s *clientAppService) CreateClientApp(ctx context.Context, req *CreateClientAppRequest, createdBy *models.User) (*ClientAppResponse, string, error) {
	// Only superadmin can create client apps
	if !createdBy.IsSuperadmin {
		return nil, "", errors.New("only superadmin can create client applications")
	}

	// Validate redirect URIs
	for _, uri := range req.RedirectURIs {
		if err := validateURI(uri); err != nil {
			return nil, "", fmt.Errorf("invalid redirect URI %s: %w", uri, err)
		}
	}

	// Generate client ID
	clientID, err := generateClientID()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate client ID: %w", err)
	}

	// Generate client secret
	clientSecret, err := generateClientSecret()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate client secret: %w", err)
	}

	// Hash client secret
	hashedSecret, err := s.passwordSvc.Hash(clientSecret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash client secret: %w", err)
	}

	// Set default scopes if not provided
	allowedScopes := req.AllowedScopes
	if len(allowedScopes) == 0 {
		allowedScopes = []string{"email", "profile", "org.read", "org.write"}
	}

	clientApp := &models.ClientApp{
		Name:           req.Name,
		ClientID:       clientID,
		ClientSecret:   hashedSecret,
		RedirectURIs:   req.RedirectURIs,
		AllowedOrigins: req.AllowedOrigins,
		AllowedScopes:  allowedScopes,
		IsConfidential: req.IsConfidential,
	}

	if err := s.repo.ClientApp().Create(ctx, clientApp); err != nil {
		return nil, "", fmt.Errorf("failed to create client app: %w", err)
	}

	response := toClientAppResponse(clientApp)
	return response, clientSecret, nil
}

func (s *clientAppService) GetClientApp(ctx context.Context, id uuid.UUID, requestedBy *models.User) (*ClientAppResponse, error) {
	if !requestedBy.IsSuperadmin {
		return nil, errors.New("only superadmin can view client applications")
	}

	clientApp, err := s.repo.ClientApp().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toClientAppResponse(clientApp), nil
}

func (s *clientAppService) GetClientAppByClientID(ctx context.Context, clientID string) (*models.ClientApp, error) {
	return s.repo.ClientApp().GetByClientID(ctx, clientID)
}

func (s *clientAppService) ListClientApps(ctx context.Context, limit, offset int, requestedBy *models.User) ([]*ClientAppResponse, int64, error) {
	if !requestedBy.IsSuperadmin {
		return nil, 0, errors.New("only superadmin can list client applications")
	}

	clientApps, total, err := s.repo.ClientApp().GetAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*ClientAppResponse, len(clientApps))
	for i, app := range clientApps {
		responses[i] = toClientAppResponse(app)
	}

	return responses, total, nil
}

func (s *clientAppService) UpdateClientApp(ctx context.Context, id uuid.UUID, req *UpdateClientAppRequest, updatedBy *models.User) (*ClientAppResponse, error) {
	if !updatedBy.IsSuperadmin {
		return nil, errors.New("only superadmin can update client applications")
	}

	clientApp, err := s.repo.ClientApp().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		clientApp.Name = req.Name
	}
	if len(req.RedirectURIs) > 0 {
		for _, uri := range req.RedirectURIs {
			if err := validateURI(uri); err != nil {
				return nil, fmt.Errorf("invalid redirect URI %s: %w", uri, err)
			}
		}
		clientApp.RedirectURIs = req.RedirectURIs
	}
	if req.AllowedOrigins != nil {
		clientApp.AllowedOrigins = req.AllowedOrigins
	}
	if req.AllowedScopes != nil {
		clientApp.AllowedScopes = req.AllowedScopes
	}

	if err := s.repo.ClientApp().Update(ctx, clientApp); err != nil {
		return nil, fmt.Errorf("failed to update client app: %w", err)
	}

	return toClientAppResponse(clientApp), nil
}

func (s *clientAppService) DeleteClientApp(ctx context.Context, id uuid.UUID, deletedBy *models.User) error {
	if !deletedBy.IsSuperadmin {
		return errors.New("only superadmin can delete client applications")
	}

	return s.repo.ClientApp().Delete(ctx, id)
}

func (s *clientAppService) RotateClientSecret(ctx context.Context, id uuid.UUID, rotatedBy *models.User) (string, error) {
	if !rotatedBy.IsSuperadmin {
		return "", errors.New("only superadmin can rotate client secrets")
	}

	clientApp, err := s.repo.ClientApp().GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	// Generate new client secret
	clientSecret, err := generateClientSecret()
	if err != nil {
		return "", fmt.Errorf("failed to generate client secret: %w", err)
	}

	// Hash new secret
	hashedSecret, err := s.passwordSvc.Hash(clientSecret)
	if err != nil {
		return "", fmt.Errorf("failed to hash client secret: %w", err)
	}

	clientApp.ClientSecret = hashedSecret
	if err := s.repo.ClientApp().Update(ctx, clientApp); err != nil {
		return "", fmt.Errorf("failed to rotate client secret: %w", err)
	}

	return clientSecret, nil
}

func (s *clientAppService) ValidateClientCredentials(ctx context.Context, clientID, clientSecret string) (*models.ClientApp, error) {
	clientApp, err := s.repo.ClientApp().GetByClientID(ctx, clientID)
	if err != nil {
		return nil, errors.New("invalid client credentials")
	}

	// Public clients don't require secret validation
	if !clientApp.IsConfidential {
		return clientApp, nil
	}

	// Verify client secret for confidential clients
	valid, err := s.passwordSvc.Verify(clientSecret, clientApp.ClientSecret)
	if err != nil || !valid {
		return nil, errors.New("invalid client credentials")
	}

	return clientApp, nil
}

func (s *clientAppService) ValidateRedirectURI(ctx context.Context, clientID, redirectURI string) error {
	clientApp, err := s.repo.ClientApp().GetByClientID(ctx, clientID)
	if err != nil {
		return errors.New("invalid client")
	}

	// Check if redirect URI matches exactly
	for _, allowedURI := range clientApp.RedirectURIs {
		if redirectURI == allowedURI {
			return nil
		}
	}

	return errors.New("redirect URI not allowed for this client")
}

// Helper functions

func generateClientID() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func generateClientSecret() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func validateURI(uri string) error {
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("invalid URI format: %w", err)
	}

	if parsedURI.Scheme == "" {
		return errors.New("URI must have a scheme (http, https, etc.)")
	}

	// No wildcard allowed
	if strings.Contains(uri, "*") {
		return errors.New("wildcard URIs are not allowed")
	}

	return nil
}

func toClientAppResponse(app *models.ClientApp) *ClientAppResponse {
	return &ClientAppResponse{
		ID:             app.ID,
		Name:           app.Name,
		ClientID:       app.ClientID,
		RedirectURIs:   app.RedirectURIs,
		AllowedOrigins: app.AllowedOrigins,
		AllowedScopes:  app.AllowedScopes,
		IsConfidential: app.IsConfidential,
		CreatedAt:      app.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      app.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
