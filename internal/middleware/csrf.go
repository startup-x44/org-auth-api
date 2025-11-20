package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// CSRFConfig holds CSRF middleware configuration
type CSRFConfig struct {
	Secret     string
	Secure     bool
	TokenName  string
	HeaderName string
	SkipPaths  []string
}

// DefaultCSRFConfig returns default CSRF configuration
func DefaultCSRFConfig(secret string, secure bool) CSRFConfig {
	return CSRFConfig{
		Secret:     secret,
		Secure:     secure,
		TokenName:  "_csrf",
		HeaderName: "X-CSRF-Token",
	}
}

// CSRFMiddleware provides CSRF protection for state-changing operations
func CSRFMiddleware(config CSRFConfig) gin.HandlerFunc {
	// In production, you might want to use a more sophisticated token store
	tokenStore := &sync.Map{}

	return func(c *gin.Context) {
		// Skip CSRF check for configured paths
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// Skip CSRF check for safe methods
		if isSafeMethod(c.Request.Method) {
			// For GET requests, provide CSRF token in header
			token := generateCSRFToken()
			c.Header(config.HeaderName, token)
			tokenStore.Store(c.ClientIP(), token)
			c.Next()
			return
		}

		// For unsafe methods, validate CSRF token
		clientToken := c.GetHeader(config.HeaderName)
		if clientToken == "" {
			// Try to get from form data
			clientToken = c.PostForm(config.TokenName)
		}

		if clientToken == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "CSRF token missing",
			})
			c.Abort()
			return
		}

		// Check if token exists in store
		storedToken, exists := tokenStore.Load(c.ClientIP())
		if !exists || storedToken.(string) != clientToken {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "CSRF token validation failed",
			})
			c.Abort()
			return
		}

		// Token is valid - in development, keep it for reuse
		// In production, you'd want to implement token rotation

		c.Next()
	}
}

// isSafeMethod checks if HTTP method is safe (read-only)
func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

// generateCSRFToken generates a random CSRF token
func generateCSRFToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback if crypto/rand fails
		return "fallback-token"
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
