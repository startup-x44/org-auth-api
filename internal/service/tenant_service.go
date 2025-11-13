package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/logger"
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
	repo        repository.Repository
	auditLogger *logger.AuditLogger
}

// NewTenantService creates a new tenant service
func NewTenantService(repo repository.Repository) TenantService {
	return &tenantService{
		repo:        repo,
		auditLogger: logger.NewAuditLogger(),
	}
}

// CreateTenant creates a new tenant
func (s *tenantService) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*TenantResponse, error) {
	// Validate request
	if err := validation.ValidateTenantName(req.Name); err != nil {
		s.auditLogger.LogTenantAction("system", "create_tenant", "", getClientIP(ctx), getUserAgent(ctx), false, fmt.Errorf("invalid tenant name: %w", err), fmt.Sprintf("name=%s, domain=%s", req.Name, req.Domain))
		return nil, fmt.Errorf("invalid tenant name: %w", err)
	}
	if err := validation.ValidateTenantDomain(req.Domain); err != nil {
		s.auditLogger.LogTenantAction("system", "create_tenant", "", getClientIP(ctx), getUserAgent(ctx), false, fmt.Errorf("invalid domain: %w", err), fmt.Sprintf("name=%s, domain=%s", req.Name, req.Domain))
		return nil, fmt.Errorf("invalid domain: %w", err)
	}

	// Create tenant
	tenant := &models.Tenant{
		Name:   req.Name,
		Domain: req.Domain,
	}

	if err := s.repo.Tenant().Create(ctx, tenant); err != nil {
		s.auditLogger.LogTenantAction("system", "create_tenant", "", getClientIP(ctx), getUserAgent(ctx), false, err, fmt.Sprintf("name=%s, domain=%s", req.Name, req.Domain))
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	s.auditLogger.LogTenantAction("system", "create_tenant", tenant.ID.String(), getClientIP(ctx), getUserAgent(ctx), true, nil, fmt.Sprintf("name=%s, domain=%s", req.Name, req.Domain))
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
		s.auditLogger.LogTenantAction("system", "update_tenant", tenantID, getClientIP(ctx), getUserAgent(ctx), false, err, "Tenant not found")
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Validate and update fields
	if req.Name != "" {
		if err := validation.ValidateTenantName(req.Name); err != nil {
			s.auditLogger.LogTenantAction("system", "update_tenant", tenantID, getClientIP(ctx), getUserAgent(ctx), false, err, fmt.Sprintf("Invalid name: %s", req.Name))
			return nil, fmt.Errorf("invalid tenant name: %w", err)
		}
		tenant.Name = req.Name
	}

	if req.Domain != "" {
		if err := validation.ValidateTenantDomain(req.Domain); err != nil {
			s.auditLogger.LogTenantAction("system", "update_tenant", tenantID, getClientIP(ctx), getUserAgent(ctx), false, err, fmt.Sprintf("Invalid domain: %s", req.Domain))
			return nil, fmt.Errorf("invalid domain: %w", err)
		}
		tenant.Domain = req.Domain
	}

	// Save changes
	if err := s.repo.Tenant().Update(ctx, tenant); err != nil {
		s.auditLogger.LogTenantAction("system", "update_tenant", tenantID, getClientIP(ctx), getUserAgent(ctx), false, err, "Failed to save changes")
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	s.auditLogger.LogTenantAction("system", "update_tenant", tenantID, getClientIP(ctx), getUserAgent(ctx), true, nil, "Tenant updated successfully")
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
		s.auditLogger.LogTenantAction("system", "delete_tenant", tenantID, getClientIP(ctx), getUserAgent(ctx), false, err, "Failed to check tenant users")
		return fmt.Errorf("failed to check tenant users: %w", err)
	}

	if userCount > 0 {
		err := errors.New("cannot delete tenant with existing users")
		s.auditLogger.LogTenantAction("system", "delete_tenant", tenantID, getClientIP(ctx), getUserAgent(ctx), false, err, fmt.Sprintf("Tenant has %d users", userCount))
		return err
	}

	if err := s.repo.Tenant().Delete(ctx, tenantID); err != nil {
		s.auditLogger.LogTenantAction("system", "delete_tenant", tenantID, getClientIP(ctx), getUserAgent(ctx), false, err, "Failed to delete tenant")
		return err
	}

	s.auditLogger.LogTenantAction("system", "delete_tenant", tenantID, getClientIP(ctx), getUserAgent(ctx), true, nil, "Tenant deleted successfully")
	return nil
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
		ID:        tenant.ID.String(),
		Name:      tenant.Name,
		Domain:    tenant.Domain,
		CreatedAt: tenant.CreatedAt,
		UpdatedAt: tenant.UpdatedAt,
	}
}

// Helper functions

// getClientIP extracts client IP from context (simplified)
func getClientIP(ctx context.Context) string {
	// In a real implementation, extract from gin.Context or HTTP headers
	// For now, return a placeholder
	return "127.0.0.1"
}

// getUserAgent extracts user agent from context (simplified)
func getUserAgent(ctx context.Context) string {
	// In a real implementation, extract from gin.Context
	// For now, return a placeholder
	return "unknown"
}