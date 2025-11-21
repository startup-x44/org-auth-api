package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CSRFConfig holds CSRF middleware configuration
type CSRFConfig struct {
	Secret       string
	Secure       bool
	TokenName    string
	HeaderName   string
	SkipPaths    []string
	TokenExpiry  time.Duration // Token expiration time
	RotateTokens bool          // Enable token rotation
}

// CSRFToken represents a CSRF token with metadata
type CSRFToken struct {
	Value     string
	ExpiresAt time.Time
	SessionID string
}

// DefaultCSRFConfig returns default CSRF configuration
func DefaultCSRFConfig(secret string, secure bool) CSRFConfig {
	return CSRFConfig{
		Secret:       secret,
		Secure:       secure,
		TokenName:    "_csrf",
		HeaderName:   "X-CSRF-Token",
		TokenExpiry:  time.Hour * 2, // 2 hours
		RotateTokens: true,
	}
}

// CSRFMiddleware provides CSRF protection with token rotation for state-changing operations
func CSRFMiddleware(config CSRFConfig) gin.HandlerFunc {
	tokenStore := &sync.Map{}

	// Background cleanup goroutine for expired tokens
	go func() {
		ticker := time.NewTicker(time.Minute * 10) // Cleanup every 10 minutes
		defer ticker.Stop()

		for range ticker.C {
			cleanupExpiredTokens(tokenStore)
		}
	}()

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
			// For GET requests, provide new CSRF token in header
			token := generateEnhancedCSRFToken(config.Secret, config.TokenExpiry)
			c.Header(config.HeaderName, token.Value)

			// Store token with session identifier
			sessionKey := getSessionKey(c)
			tokenStore.Store(sessionKey, token)

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
				"success":    false,
				"error_code": "CSRF_TOKEN_MISSING",
				"message":    "CSRF token is required",
			})
			c.Abort()
			return
		}

		// Validate token
		sessionKey := getSessionKey(c)
		storedTokenInterface, exists := tokenStore.Load(sessionKey)
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "CSRF_TOKEN_INVALID",
				"message":    "CSRF token validation failed",
			})
			c.Abort()
			return
		}

		storedToken, ok := storedTokenInterface.(*CSRFToken)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "CSRF_TOKEN_INVALID",
				"message":    "CSRF token validation failed",
			})
			c.Abort()
			return
		}

		// Check token expiration
		if time.Now().After(storedToken.ExpiresAt) {
			tokenStore.Delete(sessionKey)
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "CSRF_TOKEN_EXPIRED",
				"message":    "CSRF token has expired",
			})
			c.Abort()
			return
		}

		// Validate token value
		if !validateCSRFToken(clientToken, storedToken.Value, config.Secret) {
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "CSRF_TOKEN_INVALID",
				"message":    "CSRF token validation failed",
			})
			c.Abort()
			return
		}

		// Token is valid - rotate if enabled
		if config.RotateTokens {
			newToken := generateEnhancedCSRFToken(config.Secret, config.TokenExpiry)
			c.Header(config.HeaderName, newToken.Value)
			tokenStore.Store(sessionKey, newToken)
		}

		c.Next()
	}
}

// isSafeMethod checks if HTTP method is safe (read-only)
func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

// generateCSRFToken generates a random CSRF token (legacy)
func generateCSRFToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback if crypto/rand fails
		return "fallback-token"
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// generateEnhancedCSRFToken generates a CSRF token with expiration and HMAC
func generateEnhancedCSRFToken(secret string, expiry time.Duration) *CSRFToken {
	// Generate random bytes for the token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		// Fallback if crypto/rand fails
		tokenBytes = []byte("fallback-token-bytes-32-chars")
	}

	// Create timestamp for expiration
	expiresAt := time.Now().Add(expiry)
	expiryStr := strconv.FormatInt(expiresAt.Unix(), 10)

	// Create token payload: base64(tokenBytes):timestamp
	tokenValue := base64.URLEncoding.EncodeToString(tokenBytes)
	payload := tokenValue + ":" + expiryStr

	// Create HMAC signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	// Final token: payload:signature
	finalToken := payload + ":" + signature

	return &CSRFToken{
		Value:     finalToken,
		ExpiresAt: expiresAt,
		SessionID: tokenValue, // Use token value as session identifier
	}
}

// validateCSRFToken validates a CSRF token with HMAC verification
func validateCSRFToken(clientToken, storedToken, secret string) bool {
	// For enhanced tokens, validate HMAC
	if strings.Contains(clientToken, ":") {
		parts := strings.Split(clientToken, ":")
		if len(parts) != 3 {
			return false
		}

		payload := parts[0] + ":" + parts[1]
		signature := parts[2]

		// Verify HMAC
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(payload))
		expectedSignature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

		// Check timestamp
		timestamp, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return false
		}

		if time.Now().Unix() > timestamp {
			return false
		}

		return hmac.Equal([]byte(signature), []byte(expectedSignature))
	}

	// Fallback to simple string comparison for legacy tokens
	return clientToken == storedToken
}

// getSessionKey generates a session key for token storage
func getSessionKey(c *gin.Context) string {
	// Try to get session ID from context or cookie
	if sessionID, exists := c.Get("session_id"); exists {
		if sid, ok := sessionID.(string); ok {
			return "csrf:" + sid
		}
	}

	// Fallback to IP + User-Agent hash for session identification
	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	hash := sha256.Sum256([]byte(clientIP + ":" + userAgent))
	return "csrf:" + base64.URLEncoding.EncodeToString(hash[:])
}

// cleanupExpiredTokens removes expired tokens from the store
func cleanupExpiredTokens(tokenStore *sync.Map) {
	now := time.Now()
	keysToDelete := make([]interface{}, 0)

	tokenStore.Range(func(key, value interface{}) bool {
		if token, ok := value.(*CSRFToken); ok {
			if now.After(token.ExpiresAt) {
				keysToDelete = append(keysToDelete, key)
			}
		}
		return true
	})

	// Delete expired tokens
	for _, key := range keysToDelete {
		tokenStore.Delete(key)
	}
}
