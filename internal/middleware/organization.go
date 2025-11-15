package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"auth-service/internal/service"
)

// OrganizationMiddleware handles organization-scoped authentication
type OrganizationMiddleware struct {
	authService service.AuthService
}

// NewOrganizationMiddleware creates a new organization middleware
func NewOrganizationMiddleware(authService service.AuthService) *OrganizationMiddleware {
	return &OrganizationMiddleware{
		authService: authService,
	}
}

// AuthRequired middleware requires valid JWT token and loads user
func (m *OrganizationMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization token required",
			})
			c.Abort()
			return
		}

		claims, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user context
		ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "is_superadmin", claims.IsSuperadmin)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// OrgResolve middleware resolves current organization from header or JWT claim
func (m *OrganizationMiddleware) OrgResolve() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.GetHeader("X-Organization-Id")
		if orgID == "" {
			// Try to get from JWT claim (for backward compatibility)
			if currentOrgID, exists := c.Request.Context().Value("current_org_id").(string); exists {
				orgID = currentOrgID
			}
		}

		if orgID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Organization ID required (X-Organization-Id header)",
			})
			c.Abort()
			return
		}

		// Validate organization exists and is active
		org, err := m.authService.OrganizationService().GetOrganization(c.Request.Context(), orgID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Organization not found",
			})
			c.Abort()
			return
		}

		if org.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Organization is not active",
			})
			c.Abort()
			return
		}

		// Set organization context
		ctx := context.WithValue(c.Request.Context(), "organization_id", orgID)
		ctx = context.WithValue(ctx, "organization", org)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// MembershipRequired middleware checks user membership and role in current organization
func (m *OrganizationMiddleware) MembershipRequired(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Request.Context().Value("user_id").(string)
		orgID, _ := c.Request.Context().Value("organization_id").(string)
		isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)

		// Superadmin bypasses membership checks
		if isSuperadmin {
			c.Next()
			return
		}

		if userID == "" || orgID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		// Check membership exists and is active
		membership, err := m.authService.OrganizationService().GetMembership(c.Request.Context(), orgID, userID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Not a member of this organization",
			})
			c.Abort()
			return
		}

		if membership.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Membership is not active",
			})
			c.Abort()
			return
		}

		// Get role name from preloaded relation
		var roleName string
		if membership.Role != nil {
			roleName = membership.Role.Name
		}

		// Check role permissions if required
		if requiredRole != "" && !m.hasRequiredRole(roleName, requiredRole) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		// Set membership context
		ctx := context.WithValue(c.Request.Context(), "membership", membership)
		ctx = context.WithValue(ctx, "user_role", roleName)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// OrgAdminRequired middleware requires organization admin role
func (m *OrganizationMiddleware) OrgAdminRequired() gin.HandlerFunc {
	return m.MembershipRequired("admin")
}

// extractToken extracts JWT token from Authorization header
func (m *OrganizationMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check if it starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}

	// Extract token part
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return strings.TrimSpace(token)
}

// hasRequiredRole checks if user role meets the required role level
func (m *OrganizationMiddleware) hasRequiredRole(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		"student": 1,
		"rto":     2,
		"issuer":  3,
		"admin":   4,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}
