package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"auth-service/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// RateLimitScope defines the type of rate limiting scope
type RateLimitScope string

const (
	ScopeLogin         RateLimitScope = "login"
	ScopeRegistration  RateLimitScope = "registration"
	ScopePasswordReset RateLimitScope = "password_reset"
	ScopeTokenRefresh  RateLimitScope = "token_refresh"
	ScopeOAuth2Token   RateLimitScope = "oauth2_token"
	ScopeAPICalls      RateLimitScope = "api_calls"
)

// RateLimiter provides Redis-backed rate limiting
type RateLimiter struct {
	redis  *redis.Client
	config *config.RateLimitConfig
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(redisClient *redis.Client, cfg *config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		config: cfg,
	}
}

// RateLimitConfig holds configuration for a specific rate limit scope
type RateLimitScopeConfig struct {
	MaxAttempts int
	Window      time.Duration
}

// GetScopeConfig returns the configuration for a specific scope
func (rl *RateLimiter) GetScopeConfig(scope RateLimitScope) RateLimitScopeConfig {
	switch scope {
	case ScopeLogin:
		return RateLimitScopeConfig{
			MaxAttempts: rl.config.LoginAttempts,
			Window:      time.Duration(rl.config.LoginWindow) * time.Second,
		}
	case ScopeRegistration:
		return RateLimitScopeConfig{
			MaxAttempts: rl.config.Registration,
			Window:      time.Duration(rl.config.RegistrationWindow) * time.Second,
		}
	case ScopePasswordReset:
		return RateLimitScopeConfig{
			MaxAttempts: rl.config.PasswordReset,
			Window:      time.Duration(rl.config.PasswordResetWindow) * time.Second,
		}
	case ScopeTokenRefresh:
		return RateLimitScopeConfig{
			MaxAttempts: rl.config.TokenRefresh,
			Window:      time.Duration(rl.config.TokenRefreshWindow) * time.Second,
		}
	case ScopeOAuth2Token:
		return RateLimitScopeConfig{
			MaxAttempts: rl.config.OAuth2Token,
			Window:      time.Duration(rl.config.OAuth2TokenWindow) * time.Second,
		}
	case ScopeAPICalls:
		return RateLimitScopeConfig{
			MaxAttempts: rl.config.APICalls,
			Window:      time.Duration(rl.config.APICallsWindow) * time.Second,
		}
	default:
		// Default to API calls config
		return RateLimitScopeConfig{
			MaxAttempts: rl.config.APICalls,
			Window:      time.Duration(rl.config.APICallsWindow) * time.Second,
		}
	}
}

// Check checks if the identifier has exceeded the rate limit for the given scope
// Returns (allowed bool, remaining int, resetTime time.Time, error)
func (rl *RateLimiter) Check(ctx context.Context, scope RateLimitScope, identifier string) (bool, int, time.Time, error) {
	if !rl.config.Enabled {
		// Rate limiting disabled - always allow
		return true, 0, time.Now(), nil
	}

	if rl.redis == nil {
		// Redis not available - fail open (allow request but log warning)
		// In production, you might want to fail closed instead
		return true, 0, time.Now(), fmt.Errorf("redis client not available")
	}

	scopeConfig := rl.GetScopeConfig(scope)
	key := fmt.Sprintf("ratelimit:%s:%s", scope, identifier)

	// Use Redis INCR with expiry for sliding window
	pipe := rl.redis.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, scopeConfig.Window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		// Redis error - fail open (allow request)
		return true, 0, time.Now(), fmt.Errorf("redis error: %w", err)
	}

	count := int(incrCmd.Val())
	remaining := scopeConfig.MaxAttempts - count
	if remaining < 0 {
		remaining = 0
	}

	// Get TTL for reset time
	ttl, err := rl.redis.TTL(ctx, key).Result()
	if err != nil {
		ttl = scopeConfig.Window
	}
	resetTime := time.Now().Add(ttl)

	allowed := count <= scopeConfig.MaxAttempts

	return allowed, remaining, resetTime, nil
}

// Reset resets the rate limit counter for a specific identifier (useful for testing or manual reset)
func (rl *RateLimiter) Reset(ctx context.Context, scope RateLimitScope, identifier string) error {
	if rl.redis == nil {
		return fmt.Errorf("redis client not available")
	}

	key := fmt.Sprintf("ratelimit:%s:%s", scope, identifier)
	return rl.redis.Del(ctx, key).Err()
}

// Middleware creates a Gin middleware for rate limiting
// Usage: router.Use(rateLimiter.Middleware(ScopeLogin, func(c *gin.Context) string { return c.ClientIP() }))
func (rl *RateLimiter) Middleware(scope RateLimitScope, identifierFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.config.Enabled {
			c.Next()
			return
		}

		identifier := identifierFunc(c)
		if identifier == "" {
			// No identifier - skip rate limiting
			c.Next()
			return
		}

		allowed, remaining, resetTime, err := rl.Check(c.Request.Context(), scope, identifier)

		// Set rate limit headers
		scopeConfig := rl.GetScopeConfig(scope)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", scopeConfig.MaxAttempts))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if err != nil {
			// Log error but don't block request (fail open)
			// In production, you might want to handle this differently
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     fmt.Sprintf("Too many requests. Please try again in %s", time.Until(resetTime).Round(time.Second)),
				"retry_after": int(time.Until(resetTime).Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ByIP creates a middleware that rate limits by IP address
func (rl *RateLimiter) ByIP(scope RateLimitScope) gin.HandlerFunc {
	return rl.Middleware(scope, func(c *gin.Context) string {
		return c.ClientIP()
	})
}

// ByEmail creates a middleware that rate limits by email from request body
func (rl *RateLimiter) ByEmail(scope RateLimitScope, emailField string) gin.HandlerFunc {
	return rl.Middleware(scope, func(c *gin.Context) string {
		// Try to get email from cached body first
		if body, exists := c.Get("rate_limit_body"); exists {
			if bodyMap, ok := body.(map[string]interface{}); ok {
				if email, ok := bodyMap[emailField].(string); ok {
					return email
				}
			}
		}

		// Read and cache body for downstream handlers
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			// Cannot read body - fallback to IP
			return c.ClientIP()
		}

		// Cache for reuse
		c.Set("rate_limit_body", body)

		if email, ok := body[emailField].(string); ok {
			return email
		}

		// Fallback to IP if no email found
		return c.ClientIP()
	})
}

// ByUserID creates a middleware that rate limits by authenticated user ID
func (rl *RateLimiter) ByUserID(scope RateLimitScope) gin.HandlerFunc {
	return rl.Middleware(scope, func(c *gin.Context) string {
		// Get user ID from context (set by auth middleware)
		if userID, exists := c.Get("user_id"); exists {
			return fmt.Sprintf("%v", userID)
		}
		// Fallback to IP if not authenticated
		return c.ClientIP()
	})
}

// Combined creates a middleware that applies multiple rate limiting strategies
// All strategies must pass for the request to be allowed
func (rl *RateLimiter) Combined(checks []struct {
	Scope      RateLimitScope
	Identifier func(*gin.Context) string
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.config.Enabled {
			c.Next()
			return
		}

		for _, check := range checks {
			identifier := check.Identifier(c)
			if identifier == "" {
				continue
			}

			allowed, remaining, resetTime, err := rl.Check(c.Request.Context(), check.Scope, identifier)

			// Set headers for first check only
			if check.Scope == checks[0].Scope {
				scopeConfig := rl.GetScopeConfig(check.Scope)
				c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", scopeConfig.MaxAttempts))
				c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
				c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
			}

			if err != nil {
				// Log error but continue (fail open)
				continue
			}

			if !allowed {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "Rate limit exceeded",
					"message":     fmt.Sprintf("Too many requests. Please try again in %s", time.Until(resetTime).Round(time.Second)),
					"retry_after": int(time.Until(resetTime).Seconds()),
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
