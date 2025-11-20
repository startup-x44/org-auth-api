package middleware

import (
	"net/http"
	"strings"

	"auth-service/internal/service"
	"auth-service/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// RevocationMiddleware checks if a token has been revoked
func RevocationMiddleware(jwtService jwt.JWTService, revocationSvc service.RevocationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]

		// Check if token is revoked
		revoked, err := revocationSvc.IsTokenRevoked(c.Request.Context(), tokenString)
		if err != nil {
			// Log error but allow request to continue
			// The JWT validation will catch invalid tokens
			c.Next()
			return
		}

		if revoked {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "token has been revoked",
			})
			c.Abort()
			return
		}

		// Additional revocation checks can be added here
		// For now, we only check token-level revocation
		// User and org-level revocation can be checked in specific handlers if needed

		c.Next()
	}
}
