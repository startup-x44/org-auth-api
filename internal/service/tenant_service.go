package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/validation"
)

// TenantService defines the interface for tenant business logic
type TenantService interface {
	CreateTenant(ctx context.Context, req *CreateTenantRequest) (*TenantResponse, error)
	GetTenant(ctx context.Context, tenantID string) (*TenantResponse, error)
	GetTenantByDomain(ctx context.Context, domain string) (*TenantResponse, error)
	UpdateTenant(ctx context.Context, tenantID string, req *UpdateTenantRequest) (*TenantResponse, error)
	DeleteTenant(ctx context.Context, tenantID string) error
	ListTenants(ctx context.Context, limit, offset int) (*TenantListResponse, error)
}

// CreateTenantRequest represents tenant creation request
type CreateTenantRequest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

// TenantResponse represents tenant response
type TenantResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateTenantRequest represents tenant update request
type UpdateTenantRequest struct {
	Name   string `json:"name,omitempty"`
	Domain string `json:"domain,omitempty"`
}

// TenantListResponse represents paginated tenant list response
type TenantListResponse struct {
	Tenants []*TenantResponse `json:"tenants"`
	Total   int64             `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
}

// tenantService implements TenantService interface
type tenantService struct {
	repo repository.Repository
}

// NewTenantService creates a new tenant service
func NewTenantService(repo repository.Repository) TenantService {
	return &tenantService{repo: repo}
}

// CreateTenant creates a new tenant
func (s *tenantService) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*TenantResponse, error) {
	// Validate request
	if err := validation.ValidateTenantName(req.Name); err != nil {
		return nil, fmt.Errorf("invalid tenant name: %w", err)
	}
	if err := validation.ValidateTenantDomain(req.Domain); err != nil {
		return nil, fmt.Errorf("invalid domain: %w", err)
	}

	// Create tenant
	tenant := &models.Tenant{
		Name:   req.Name,
		Domain: req.Domain,
	}

	if err := s.repo.Tenant().Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return s.convertToTenantResponse(tenant), nil
}

// GetTenant gets a tenant by ID
func (s *tenantService) GetTenant(ctx context.Context, tenantID string) (*TenantResponse, error) {
	if tenantID == "" {
		return nil, errors.New("tenant ID is required")
	}

	tenant, err := s.repo.Tenant().GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	return s.convertToTenantResponse(tenant), nil
}

// GetTenantByDomain gets a tenant by domain
func (s *tenantService) GetTenantByDomain(ctx context.Context, domain string) (*TenantResponse, error) {
	if domain == "" {
		return nil, errors.New("domain is required")
	}

	tenant, err := s.repo.Tenant().GetByDomain(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	return s.convertToTenantResponse(tenant), nil
}

// UpdateTenant updates a tenant
func (s *tenantService) UpdateTenant(ctx context.Context, tenantID string, req *UpdateTenantRequest) (*TenantResponse, error) {
	if tenantID == "" {
		return nil, errors.New("tenant ID is required")
	}

	// Get current tenant
	tenant, err := s.repo.Tenant().GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Validate and update fields
	if req.Name != "" {
		if err := validation.ValidateTenantName(req.Name); err != nil {
			return nil, fmt.Errorf("invalid tenant name: %w", err)
		}
		tenant.Name = req.Name
	}

	if req.Domain != "" {
		if err := validation.ValidateTenantDomain(req.Domain); err != nil {
			return nil, fmt.Errorf("invalid domain: %w", err)
		}
		tenant.Domain = req.Domain
	}

	// Save changes
	if err := s.repo.Tenant().Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return s.convertToTenantResponse(tenant), nil
}

// DeleteTenant deletes a tenant
func (s *tenantService) DeleteTenant(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return errors.New("tenant ID is required")
	}

	// Check if tenant has users
	userCount, err := s.repo.User().Count(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to check tenant users: %w", err)
	}

	if userCount > 0 {
		return errors.New("cannot delete tenant with existing users")
	}

	return s.repo.Tenant().Delete(ctx, tenantID)
}

// ListTenants lists tenants with pagination
func (s *tenantService) ListTenants(ctx context.Context, limit, offset int) (*TenantListResponse, error) {
	tenants, err := s.repo.Tenant().List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	total, err := s.repo.Tenant().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count tenants: %w", err)
	}

	tenantResponses := make([]*TenantResponse, len(tenants))
	for i, tenant := range tenants {
		tenantResponses[i] = s.convertToTenantResponse(tenant)
	}

	return &TenantListResponse{
		Tenants: tenantResponses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

// convertToTenantResponse converts model to response
func (s *tenantService) convertToTenantResponse(tenant *models.Tenant) *TenantResponse {
	return &TenantResponse{
		ID:        tenant.ID,
		Name:      tenant.Name,
		Domain:    tenant.Domain,
		CreatedAt: tenant.CreatedAt,
		UpdatedAt: tenant.UpdatedAt,
	}
}