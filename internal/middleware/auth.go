package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"auth-service/internal/service"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	authService service.AuthService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// AuthRequired middleware requires valid JWT token
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

		// Set user context
		ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "is_superadmin", claims.IsSuperadmin)
		c.Request = c.Request.WithContext(ctx)

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

// TenantRequired middleware ensures tenant context is set
func (m *AuthMiddleware) TenantRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			// Try to get from JWT claims
			if claimsTenantID, exists := c.Request.Context().Value("tenant_id").(string); exists {
				tenantID = claimsTenantID
			}
		}

		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Tenant ID required",
			})
			c.Abort()
			return
		}

		// Set tenant context
		ctx := context.WithValue(c.Request.Context(), "tenant_id", tenantID)
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
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Tenant-ID")
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
