package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"auth-service/internal/repository"
	"auth-service/internal/service"
)

// ResourcePolicy defines access control policies for different resource types
type ResourcePolicy struct {
	// Administrative resources that superadmin can bypass (admin-only functions)
	AdminResources []string
	// Role-specific resources that require exact role match (no superadmin bypass)
	RoleSpecificResources []string
	// User-facing resources that follow role hierarchy
	UserResources []string
}

// DefaultResourcePolicy returns the default resource access policy
func DefaultResourcePolicy() *ResourcePolicy {
	return &ResourcePolicy{
		AdminResources: []string{
			"admin:", "system:", "rbac:", "client-apps:", "audit:",
			"users:create", "users:update", "users:delete", "users:activate", "users:deactivate",
			"organizations:create", "organizations:update", "organizations:delete",
			"roles:create", "roles:update", "roles:delete", "roles:assign",
			"permissions:create", "permissions:update", "permissions:delete",
		},
		RoleSpecificResources: []string{
			"role:user", "role:member", "role:admin", "role:superadmin",
			"dashboard:user", "dashboard:member", "dashboard:admin",
			"access:user-routes", "access:member-routes",
		},
		UserResources: []string{
			"profile:", "settings:", "notifications:", "user:read",
			"member:view", "organization:view",
		},
	}
}

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	authService service.AuthService
	repo        repository.Repository
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService service.AuthService, repo repository.Repository) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		repo:        repo,
	}
}

// canAccessResource implements the unified policy logic for resource access control
func (m *AuthMiddleware) canAccessResource(c *gin.Context, resource string, policy *ResourcePolicy) bool {
	isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)

	// Get user permissions and roles from token
	token := m.extractToken(c)
	if token == "" {
		return false
	}

	claims, err := m.authService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		return false
	}

	// 1. Check if it's a role-specific resource (strict matching required)
	for _, roleResource := range policy.RoleSpecificResources {
		if strings.HasPrefix(resource, roleResource) {
			// For role-specific resources, check exact permission match
			// SuperAdmin CANNOT bypass role-specific permissions
			return m.hasExactPermission(claims.Permissions, resource)
		}
	}

	// 2. Check if it's an administrative resource (superadmin bypass allowed)
	for _, adminResource := range policy.AdminResources {
		if strings.HasPrefix(resource, adminResource) {
			// SuperAdmin can bypass admin resources
			if isSuperadmin {
				return true
			}
			// Non-superadmin must have the specific permission
			return m.hasExactPermission(claims.Permissions, resource)
		}
	}

	// 3. Check if it's a user resource (role hierarchy applies)
	for _, userResource := range policy.UserResources {
		if strings.HasPrefix(resource, userResource) {
			// Check role hierarchy: superadmin > admin > member > user
			return m.hasRoleHierarchyAccess(claims.Roles, resource) ||
				m.hasExactPermission(claims.Permissions, resource)
		}
	}

	// 4. Default: check exact permission match (no superadmin bypass)
	return m.hasExactPermission(claims.Permissions, resource)
}

// hasExactPermission checks if user has the specific permission
func (m *AuthMiddleware) hasExactPermission(permissions []string, required string) bool {
	for _, p := range permissions {
		if p == required {
			return true
		}
	}
	return false
}

// hasRole checks if user has a specific role
func (m *AuthMiddleware) hasRole(c *gin.Context, requiredRole string) bool {
	token := m.extractToken(c)
	if token == "" {
		return false
	}

	claims, err := m.authService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		return false
	}

	for _, role := range claims.Roles {
		if role == requiredRole {
			return true
		}
	}
	return false
}

// hasRoleHierarchyAccess checks role hierarchy for user-facing resources
func (m *AuthMiddleware) hasRoleHierarchyAccess(userRoles []string, resource string) bool {
	// Define role hierarchy: superadmin > admin > member > user
	roleHierarchy := map[string]int{
		"superadmin": 4,
		"admin":      3,
		"member":     2,
		"user":       1,
	}

	// Get the highest role level for the user
	userLevel := 0
	for _, role := range userRoles {
		if level, exists := roleHierarchy[role]; exists && level > userLevel {
			userLevel = level
		}
	}

	// Determine required level based on resource
	requiredLevel := 1 // default to user level
	if strings.Contains(resource, "admin") {
		requiredLevel = 3
	} else if strings.Contains(resource, "member") {
		requiredLevel = 2
	}

	return userLevel >= requiredLevel
}

// canAccessResource implements the unified policy logic for resource access
func (m *AuthMiddleware) canAccessResource(c *gin.Context, resource string, policy *ResourcePolicy) bool {
	isSuperadmin, _ := c.Request.Context().Value("is_superadmin").(bool)

	// Extract user info from context
	token := m.extractToken(c)
	if token == "" {
		return false
	}

	claims, err := m.authService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		return false
	}

	// 1. Check if it's a role-specific resource (strict matching required - NO superadmin bypass)
	for _, roleResource := range policy.RoleSpecificResources {
		if strings.HasPrefix(resource, roleResource) {
			// For role-specific resources, check exact permission match only
			return m.hasExactPermission(claims.Permissions, resource)
		}
	}

	// 2. Check if it's an administrative resource (superadmin bypass allowed)
	for _, adminResource := range policy.AdminResources {
		if strings.HasPrefix(resource, adminResource) {
			// SuperAdmin can bypass admin resources
			if isSuperadmin {
				return true
			}
			// Non-superadmin must have the specific permission
			return m.hasExactPermission(claims.Permissions, resource)
		}
	}

	// 3. Check if it's a user resource (role hierarchy applies)
	for _, userResource := range policy.UserResources {
		if strings.HasPrefix(resource, userResource) {
			// Check role hierarchy or exact permission
			return m.hasRoleHierarchyAccess(claims, resource) ||
				m.hasExactPermission(claims.Permissions, resource)
		}
	}

	// 4. Default: check exact permission match (no superadmin bypass)
	return m.hasExactPermission(claims.Permissions, resource)
}

// hasExactPermission checks if user has the specific permission
func (m *AuthMiddleware) hasExactPermission(permissions []string, required string) bool {
	for _, p := range permissions {
		if p == required {
			return true
		}
	}
	return false
}

// hasRole checks if user has a specific role
func (m *AuthMiddleware) hasRole(c *gin.Context, role string) bool {
	token := m.extractToken(c)
	if token == "" {
		return false
	}

	claims, err := m.authService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		return false
	}

	// Check if user has the specific role
	for _, userRole := range claims.Roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// hasRoleHierarchyAccess checks role hierarchy for user resources
// Hierarchy: superadmin > admin > member > user
func (m *AuthMiddleware) hasRoleHierarchyAccess(claims interface{}, resource string) bool {
	// This is a placeholder for role hierarchy logic
	// You can implement specific hierarchy rules here based on your needs
	// For now, we'll rely on exact permission matching
	return false
}

// AuthRequired middleware requires valid JWT token
// NO superadmin bypass here - just validates token and sets context
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
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

		// Set comprehensive user context for policy checks
		ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "organization_id", claims.OrganizationID)
		ctx = context.WithValue(ctx, "organization_role", claims.OrganizationRole)
		ctx = context.WithValue(ctx, "is_superadmin", claims.IsSuperadmin)
		ctx = context.WithValue(ctx, "permissions", claims.Permissions)
		ctx = context.WithValue(ctx, "roles", claims.Roles)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// LoadUser middleware loads the full user object and sets it in Gin context
// This should be used after AuthRequired middleware
func (m *AuthMiddleware) LoadUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Request.Context().Value("user_id").(string)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User ID not found in context",
			})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid user ID format",
			})
			c.Abort()
			return
		}

		// Get user from repository
		user, err := m.repo.User().GetByID(c.Request.Context(), userID.String())
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User not found",
			})
			c.Abort()
			return
		}

		// Set user in Gin context for handlers to access
		c.Set("user", user)
		c.Next()
	}
}

// AdminRequired middleware requires superadmin privileges
func (m *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		isSuperadmin, exists := c.Request.Context().Value("is_superadmin").(bool)
		if !exists || !isSuperadmin {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Superadmin access required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequirePermission middleware requires a specific permission
// Superadmins bypass all permission checks
func (m *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Superadmin bypass
		if isSuperadmin, exists := c.Request.Context().Value("is_superadmin").(bool); exists && isSuperadmin {
			c.Next()
			return
		}

		// Extract permissions from token claims
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

		// Check if user has the required permission
		hasPermission := false
		for _, p := range claims.Permissions {
			if p == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"success":  false,
				"message":  "Insufficient permissions",
				"required": permission,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission middleware requires at least one of the specified permissions
func (m *AuthMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Superadmin bypass
		if isSuperadmin, exists := c.Request.Context().Value("is_superadmin").(bool); exists && isSuperadmin {
			c.Next()
			return
		}

		// Extract permissions from token claims
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

		// Check if user has any of the required permissions
		hasPermission := false
		for _, requiredPerm := range permissions {
			for _, userPerm := range claims.Permissions {
				if userPerm == requiredPerm {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"success":  false,
				"message":  "Insufficient permissions",
				"required": permissions,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllPermissions middleware requires all of the specified permissions
func (m *AuthMiddleware) RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Superadmin bypass
		if isSuperadmin, exists := c.Request.Context().Value("is_superadmin").(bool); exists && isSuperadmin {
			c.Next()
			return
		}

		// Extract permissions from token claims
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

		// Check if user has all required permissions
		userPermMap := make(map[string]bool)
		for _, p := range claims.Permissions {
			userPermMap[p] = true
		}

		missingPerms := []string{}
		for _, requiredPerm := range permissions {
			if !userPermMap[requiredPerm] {
				missingPerms = append(missingPerms, requiredPerm)
			}
		}

		if len(missingPerms) > 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Insufficient permissions",
				"missing": missingPerms,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OrganizationRequired middleware ensures organization context is set
func (m *AuthMiddleware) OrganizationRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		organizationID := c.GetHeader("X-Organization-ID")
		if organizationID == "" {
			// Try to get from JWT claims
			if claimsOrgID, exists := c.Request.Context().Value("organization_id").(string); exists {
				organizationID = claimsOrgID
			}
		}

		if organizationID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Organization ID required",
			})
			c.Abort()
			return
		}

		// Set organization context
		ctx := context.WithValue(c.Request.Context(), "organization_id", organizationID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// extractToken extracts JWT token from Authorization header
func (m *AuthMiddleware) extractToken(c *gin.Context) string {
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

// CORSMiddleware handles CORS headers with tenant subdomain support
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// If no origin header, skip CORS processing
		if origin == "" {
			c.Next()
			return
		}

		// Check if the origin is allowed
		if !isOriginAllowed(origin, allowedOrigins) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Origin not allowed",
			})
			c.Abort()
			return
		}

		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Tenant-ID, X-Organization-ID")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Header("Access-Control-Expose-Headers", "X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isOriginAllowed checks if an origin is allowed based on the configuration
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}

		// Check for wildcard patterns like "*.sprout.com"
		if strings.HasPrefix(allowed, "*.") {
			baseDomain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, baseDomain) {
				// Ensure it's a subdomain (not the base domain itself unless explicitly allowed)
				originWithoutProtocol := strings.TrimPrefix(origin, "http://")
				originWithoutProtocol = strings.TrimPrefix(originWithoutProtocol, "https://")
				if strings.Contains(originWithoutProtocol, ".") && strings.HasSuffix(originWithoutProtocol, baseDomain) {
					return true
				}
			}
		} else if allowed == origin {
			// Exact match
			return true
		}
	}
	return false
}

// RateLimitMiddleware provides basic rate limiting (placeholder)
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement proper rate limiting with Redis
		// For now, just pass through
		c.Next()
	}
}

// LoggingMiddleware logs requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement structured logging
		c.Next()
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
