package middleware

import (
	"os"
	"strings"
	"time"

	"auth-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// StructuredLoggingMiddleware adds request_id and logs all requests with structured fields
func StructuredLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Create context with request ID
		ctx := c.Request.Context()
		ctx = logger.WithContext(ctx, logger.RequestIDKey, requestID)

		// Add IP address to context
		clientIP := c.ClientIP()
		ctx = logger.WithContext(ctx, logger.IPAddressKey, clientIP)

		// Add user agent to context
		userAgent := c.Request.UserAgent()
		ctx = logger.WithContext(ctx, logger.UserAgentKey, userAgent)

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Determine if this is a high-traffic endpoint
		path := c.Request.URL.Path
		isAuthEndpoint := strings.HasPrefix(path, "/api/v1/auth/") ||
			strings.Contains(path, "/login") ||
			strings.Contains(path, "/refresh") ||
			strings.Contains(path, "/oauth/token")

		// Apply sampling based on environment and endpoint
		isProduction := os.Getenv("ENVIRONMENT") == "production"
		shouldLog := true

		if isProduction {
			if isAuthEndpoint {
				// Sample auth endpoints at 1:20 in production
				sampler := &zerolog.BasicSampler{N: 20}
				shouldLog = sampler.Sample(0)
			} else {
				// Sample all other endpoints at 1:10 in production
				sampler := &zerolog.BasicSampler{N: 10}
				shouldLog = sampler.Sample(0)
			}
		}

		// Log request with structured fields
		if shouldLog {
			logger.Info(ctx).
				Str("method", c.Request.Method).
				Str("path", path).
				Str("query", c.Request.URL.RawQuery).
				Int("status", c.Writer.Status()).
				Dur("latency", latency).
				Int("body_size", c.Writer.Size()).
				Str("client_ip", clientIP).
				Str("user_agent", userAgent).
				Msg("HTTP request")
		}

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				metaStr, _ := err.Meta.(string)
				logger.Error(ctx).
					Err(err.Err).
					Str("error_meta", metaStr).
					Msg("Request error")
			}
		}
	}
}

// AddUserContextMiddleware adds user_id, org_id, session_id to logging context
// Should be applied after authentication middleware
func AddUserContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Extract user_id from context (set by auth middleware)
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(string); ok && uid != "" {
				ctx = logger.WithContext(ctx, logger.UserIDKey, uid)
			}
		}

		// Extract org_id from context
		if orgID, exists := c.Get("org_id"); exists {
			if oid, ok := orgID.(string); ok && oid != "" {
				ctx = logger.WithContext(ctx, logger.OrgIDKey, oid)
			}
		}

		// Extract session_id from context
		if sessionID, exists := c.Get("session_id"); exists {
			if sid, ok := sessionID.(string); ok && sid != "" {
				ctx = logger.WithContext(ctx, logger.SessionIDKey, sid)
			}
		}

		// Extract client_id for OAuth2 requests
		if clientID, exists := c.Get("client_id"); exists {
			if cid, ok := clientID.(string); ok && cid != "" {
				ctx = logger.WithContext(ctx, logger.ClientIDKey, cid)
			}
		}

		// Update request context
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
