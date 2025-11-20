package unit_test

import (
	"context"
	"testing"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/middleware"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRateLimiter_BasicFunctionality tests core rate limiting logic
func TestRateLimiter_BasicFunctionality(t *testing.T) {
	// Setup mock Redis server
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	cfg := &config.RateLimitConfig{
		Enabled:             true,
		LoginAttempts:       5,
		LoginWindow:         60,
		Registration:        3,
		RegistrationWindow:  300,
		PasswordReset:       3,
		PasswordResetWindow: 3600,
		TokenRefresh:        10,
		TokenRefreshWindow:  60,
		OAuth2Token:         10,
		OAuth2TokenWindow:   60,
		APICalls:            100,
		APICallsWindow:      60,
	}

	limiter := middleware.NewRateLimiter(redisClient, cfg)
	ctx := context.Background()

	t.Run("Allow requests within limit", func(t *testing.T) {
		identifier := "test-user-1@example.com"

		// First 5 login attempts should be allowed
		for i := 0; i < 5; i++ {
			allowed, remaining, _, err := limiter.Check(ctx, middleware.ScopeLogin, identifier)
			require.NoError(t, err)
			assert.True(t, allowed, "Request %d should be allowed", i+1)
			assert.Equal(t, 5-i-1, remaining, "Remaining requests should decrease")
		}
	})

	t.Run("Block requests exceeding limit", func(t *testing.T) {
		identifier := "test-user-2@example.com"

		// Use up all 5 login attempts
		for i := 0; i < 5; i++ {
			allowed, _, _, err := limiter.Check(ctx, middleware.ScopeLogin, identifier)
			require.NoError(t, err)
			assert.True(t, allowed)
		}

		// 6th attempt should be blocked
		allowed, remaining, resetTime, err := limiter.Check(ctx, middleware.ScopeLogin, identifier)
		require.NoError(t, err)
		assert.False(t, allowed, "Request exceeding limit should be blocked")
		assert.Equal(t, 0, remaining)
		assert.True(t, resetTime.After(time.Now()), "Reset time should be in the future")
	})

	t.Run("Different scopes have independent limits", func(t *testing.T) {
		identifier := "test-user-3@example.com"

		// Use login scope
		for i := 0; i < 5; i++ {
			allowed, _, _, err := limiter.Check(ctx, middleware.ScopeLogin, identifier)
			require.NoError(t, err)
			assert.True(t, allowed)
		}

		// Login scope should be exhausted
		allowed, _, _, err := limiter.Check(ctx, middleware.ScopeLogin, identifier)
		require.NoError(t, err)
		assert.False(t, allowed)

		// But registration scope should still be available
		allowed, _, _, err = limiter.Check(ctx, middleware.ScopeRegistration, identifier)
		require.NoError(t, err)
		assert.True(t, allowed, "Registration scope should be independent from login scope")
	})

	t.Run("Different identifiers have independent limits", func(t *testing.T) {
		user1 := "user1@example.com"
		user2 := "user2@example.com"

		// Exhaust user1's limit
		for i := 0; i < 5; i++ {
			allowed, _, _, err := limiter.Check(ctx, middleware.ScopeLogin, user1)
			require.NoError(t, err)
			assert.True(t, allowed)
		}

		// User1 should be blocked
		allowed, _, _, err := limiter.Check(ctx, middleware.ScopeLogin, user1)
		require.NoError(t, err)
		assert.False(t, allowed)

		// But user2 should still be allowed
		allowed, _, _, err = limiter.Check(ctx, middleware.ScopeLogin, user2)
		require.NoError(t, err)
		assert.True(t, allowed, "User2 should have independent rate limit from user1")
	})
}

func TestRateLimiter_ScopeConfigurations(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	cfg := &config.RateLimitConfig{
		Enabled:             true,
		LoginAttempts:       5,
		LoginWindow:         60,
		Registration:        3,
		RegistrationWindow:  300,
		PasswordReset:       3,
		PasswordResetWindow: 3600,
		TokenRefresh:        10,
		TokenRefreshWindow:  60,
		OAuth2Token:         10,
		OAuth2TokenWindow:   60,
		APICalls:            100,
		APICallsWindow:      60,
	}

	limiter := middleware.NewRateLimiter(redisClient, cfg)
	ctx := context.Background()

	testCases := []struct {
		scope       middleware.RateLimitScope
		maxAttempts int
		name        string
	}{
		{middleware.ScopeLogin, 5, "Login scope"},
		{middleware.ScopeRegistration, 3, "Registration scope"},
		{middleware.ScopePasswordReset, 3, "Password reset scope"},
		{middleware.ScopeTokenRefresh, 10, "Token refresh scope"},
		{middleware.ScopeOAuth2Token, 10, "OAuth2 token scope"},
		{middleware.ScopeAPICalls, 100, "API calls scope"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			identifier := "test-" + string(tc.scope) + "@example.com"

			// Test that the correct number of attempts are allowed
			for i := 0; i < tc.maxAttempts; i++ {
				allowed, remaining, _, err := limiter.Check(ctx, tc.scope, identifier)
				require.NoError(t, err)
				assert.True(t, allowed, "Attempt %d/%d should be allowed", i+1, tc.maxAttempts)
				assert.Equal(t, tc.maxAttempts-i-1, remaining)
			}

			// Next attempt should be blocked
			allowed, _, _, err := limiter.Check(ctx, tc.scope, identifier)
			require.NoError(t, err)
			assert.False(t, allowed, "Attempt %d should be blocked", tc.maxAttempts+1)
		})
	}
}

func TestRateLimiter_ResetWindow(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	// Use very short window for testing
	cfg := &config.RateLimitConfig{
		Enabled:       true,
		LoginAttempts: 2,
		LoginWindow:   2, // 2 seconds
	}

	limiter := middleware.NewRateLimiter(redisClient, cfg)
	ctx := context.Background()
	identifier := "reset-test@example.com"

	t.Run("Limit resets after window expires", func(t *testing.T) {
		// Use up the limit
		for i := 0; i < 2; i++ {
			allowed, _, _, err := limiter.Check(ctx, middleware.ScopeLogin, identifier)
			require.NoError(t, err)
			assert.True(t, allowed)
		}

		// Should be blocked
		allowed, _, _, err := limiter.Check(ctx, middleware.ScopeLogin, identifier)
		require.NoError(t, err)
		assert.False(t, allowed, "Should be blocked after limit exceeded")

		// Fast-forward time in miniredis
		mr.FastForward(3 * time.Second)

		// Should be allowed again after window reset
		allowed, remaining, _, err := limiter.Check(ctx, middleware.ScopeLogin, identifier)
		require.NoError(t, err)
		assert.True(t, allowed, "Should be allowed after window reset")
		assert.Equal(t, 1, remaining, "Should have full limit minus one")
	})
}

func TestRateLimiter_DisabledConfig(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	cfg := &config.RateLimitConfig{
		Enabled: false, // Rate limiting disabled
	}

	limiter := middleware.NewRateLimiter(redisClient, cfg)
	ctx := context.Background()
	identifier := "test@example.com"

	t.Run("All requests allowed when disabled", func(t *testing.T) {
		// Should allow unlimited requests when disabled
		for i := 0; i < 100; i++ {
			allowed, _, _, err := limiter.Check(ctx, middleware.ScopeLogin, identifier)
			require.NoError(t, err)
			assert.True(t, allowed, "All requests should be allowed when rate limiting is disabled")
		}
	})
}

func TestRateLimiter_RedisFailure(t *testing.T) {
	// Create Redis client pointing to non-existent server
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Non-existent Redis server
	})
	defer redisClient.Close()

	cfg := &config.RateLimitConfig{
		Enabled:       true,
		LoginAttempts: 5,
		LoginWindow:   60,
	}

	limiter := middleware.NewRateLimiter(redisClient, cfg)
	ctx := context.Background()
	identifier := "test@example.com"

	t.Run("Fail open when Redis unavailable", func(t *testing.T) {
		// Should fail open (allow request) when Redis is unavailable
		allowed, _, _, err := limiter.Check(ctx, middleware.ScopeLogin, identifier)

		// Error should be returned but request should be allowed
		assert.Error(t, err, "Should return error when Redis unavailable")
		assert.True(t, allowed, "Should fail open (allow) when Redis unavailable")
	})
}

func TestRateLimiter_ConcurrentRequests(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	cfg := &config.RateLimitConfig{
		Enabled:       true,
		LoginAttempts: 10,
		LoginWindow:   60,
	}

	limiter := middleware.NewRateLimiter(redisClient, cfg)
	ctx := context.Background()
	identifier := "concurrent-test@example.com"

	t.Run("Concurrent requests respect limit", func(t *testing.T) {
		concurrency := 20
		successChan := make(chan bool, concurrency)

		// Launch concurrent requests
		for i := 0; i < concurrency; i++ {
			go func() {
				allowed, _, _, _ := limiter.Check(ctx, middleware.ScopeLogin, identifier)
				successChan <- allowed
			}()
		}

		// Collect results
		successCount := 0
		blockedCount := 0
		for i := 0; i < concurrency; i++ {
			if <-successChan {
				successCount++
			} else {
				blockedCount++
			}
		}

		// Should allow exactly the limit (10) and block the rest
		assert.LessOrEqual(t, successCount, 10, "Should not allow more than the limit")
		assert.GreaterOrEqual(t, successCount, 8, "Should allow close to the limit (accounting for race conditions)")
		assert.GreaterOrEqual(t, blockedCount, 10, "Should block requests exceeding the limit")

		t.Logf("Concurrent test: %d allowed, %d blocked (limit: 10)", successCount, blockedCount)
	})
}

func TestRateLimiter_GetScopeConfig(t *testing.T) {
	cfg := &config.RateLimitConfig{
		Enabled:             true,
		LoginAttempts:       5,
		LoginWindow:         60,
		Registration:        3,
		RegistrationWindow:  300,
		PasswordReset:       3,
		PasswordResetWindow: 3600,
		TokenRefresh:        10,
		TokenRefreshWindow:  60,
		OAuth2Token:         15,
		OAuth2TokenWindow:   30,
		APICalls:            100,
		APICallsWindow:      60,
	}

	limiter := middleware.NewRateLimiter(nil, cfg)

	testCases := []struct {
		scope          middleware.RateLimitScope
		expectedMax    int
		expectedWindow time.Duration
		name           string
	}{
		{middleware.ScopeLogin, 5, 60 * time.Second, "Login"},
		{middleware.ScopeRegistration, 3, 300 * time.Second, "Registration"},
		{middleware.ScopePasswordReset, 3, 3600 * time.Second, "Password Reset"},
		{middleware.ScopeTokenRefresh, 10, 60 * time.Second, "Token Refresh"},
		{middleware.ScopeOAuth2Token, 15, 30 * time.Second, "OAuth2 Token"},
		{middleware.ScopeAPICalls, 100, 60 * time.Second, "API Calls"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scopeCfg := limiter.GetScopeConfig(tc.scope)
			assert.Equal(t, tc.expectedMax, scopeCfg.MaxAttempts, "Max attempts mismatch")
			assert.Equal(t, tc.expectedWindow, scopeCfg.Window, "Window mismatch")
		})
	}
}
