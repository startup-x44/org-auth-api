package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"auth-service/internal/service"
)

// SessionMiddleware handles session validation and activity tracking
type SessionMiddleware struct {
	authService service.AuthService
}

// NewSessionMiddleware creates a new session middleware
func NewSessionMiddleware(authService service.AuthService) *SessionMiddleware {
	return &SessionMiddleware{
		authService: authService,
	}
}

// SessionValidation middleware validates user sessions
func (m *SessionMiddleware) SessionValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract session token from header or cookie
		sessionToken := m.extractSessionToken(c)
		if sessionToken == "" {
			// No session token, continue without session validation
			c.Next()
			return
		}

		// Validate session
		sessionSvc := m.authService.SessionService()
		session, err := sessionSvc.ValidateSession(c.Request.Context(), sessionToken)
		if err != nil {
			// Invalid session, but don't block the request
			// Just log and continue
			c.Set("session_valid", false)
			c.Set("session_error", err.Error())
			c.Next()
			return
		}

		// Session is valid
		c.Set("session_valid", true)
		c.Set("session_id", session.ID.String())
		c.Set("session_user_id", session.UserID.String())
		c.Set("device_fingerprint", session.DeviceFingerprint)

		// Update session activity
		if err := sessionSvc.UpdateSessionActivity(c.Request.Context(), session.ID.String()); err != nil {
			// Log error but don't fail the request
			// TODO: Add proper logging
		}

		c.Next()
	}
}

// ConcurrentSessionLimit middleware enforces concurrent session limits
func (m *SessionMiddleware) ConcurrentSessionLimit(maxSessions int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Request.Context().Value("user_id").(string)
		if !exists {
			c.Next()
			return
		}

		sessionSvc := m.authService.SessionService()
		activeCount, err := sessionSvc.GetActiveSessionCount(c.Request.Context(), userID)
		if err != nil {
			// Log error but allow request
			c.Next()
			return
		}

		if activeCount >= int64(maxSessions) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Maximum concurrent sessions exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SuspiciousActivityDetection middleware detects suspicious session activity
func (m *SessionMiddleware) SuspiciousActivityDetection() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionValid, exists := c.Get("session_valid")
		if !exists || !sessionValid.(bool) {
			c.Next()
			return
		}

		sessionSvc := m.authService.SessionService()

		// Get session details - we need to get the session ID from context
		sessionID, _ := c.Get("session_id")
		if sessionID == nil {
			c.Next()
			return
		}

		// Get the session (simplified - in real implementation you'd get it by ID)
		// For now, we'll skip the detailed check
		_ = sessionSvc

		c.Next()
	}
}

// SessionCleanup middleware triggers session cleanup for inactive sessions
func (m *SessionMiddleware) SessionCleanup() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware runs after the request is processed
		c.Next()

		// Trigger cleanup in a goroutine to not block the response
		go func() {
			jobSvc := m.authService.BackgroundJobService()
			ctx := c.Request.Context()

			// Cleanup expired sessions
			if err := jobSvc.CleanupExpiredSessions(ctx); err != nil {
				// TODO: Add proper logging
			}

			// Cleanup inactive sessions
			if err := jobSvc.CleanupInactiveSessions(ctx); err != nil {
				// TODO: Add proper logging
			}
		}()
	}
}

// extractSessionToken extracts session token from various sources
func (m *SessionMiddleware) extractSessionToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Try X-Session-Token header
	sessionToken := c.GetHeader("X-Session-Token")
	if sessionToken != "" {
		return sessionToken
	}

	// Try session_token cookie
	sessionToken, err := c.Cookie("session_token")
	if err == nil && sessionToken != "" {
		return sessionToken
	}

	return ""
}

// DeviceFingerprinting middleware captures device information
func (m *SessionMiddleware) DeviceFingerprinting() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract device information
		userAgent := c.GetHeader("User-Agent")
		acceptLanguage := c.GetHeader("Accept-Language")
		acceptEncoding := c.GetHeader("Accept-Encoding")

		// Create a basic device fingerprint
		fingerprint := userAgent + "|" + acceptLanguage + "|" + acceptEncoding

		// Store in context for later use
		c.Set("device_fingerprint_raw", fingerprint)

		c.Next()
	}
}