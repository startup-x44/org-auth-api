package metrics

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"auth-service/pkg/logger"
)

// Collector periodically collects system metrics
type Collector struct {
	db          *gorm.DB
	redisClient *redis.Client
	metrics     *Metrics
	interval    time.Duration
	stopChan    chan struct{}
}

// NewCollector creates a new metrics collector
func NewCollector(db *gorm.DB, redisClient *redis.Client, interval time.Duration) *Collector {
	return &Collector{
		db:          db,
		redisClient: redisClient,
		metrics:     GetMetrics(),
		interval:    interval,
		stopChan:    make(chan struct{}),
	}
}

// Start begins periodic collection of metrics
func (c *Collector) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Collect initial metrics
	c.collectMetrics(ctx)

	for {
		select {
		case <-ticker.C:
			c.collectMetrics(ctx)
		case <-c.stopChan:
			logger.InfoMsg("Metrics collector stopped")
			return
		case <-ctx.Done():
			logger.InfoMsg("Metrics collector stopped due to context cancellation")
			return
		}
	}
}

// Stop stops the metrics collector
func (c *Collector) Stop() {
	close(c.stopChan)
}

// collectMetrics collects all periodic metrics
func (c *Collector) collectMetrics(ctx context.Context) {
	c.collectDatabaseMetrics()
	c.collectRedisMetrics(ctx)
	c.collectTokenMetrics(ctx)
	c.collectOrganizationMetrics(ctx)
	c.collectSessionMetrics(ctx)
	c.collectAPIKeyMetrics(ctx)
}

// collectDatabaseMetrics collects database connection pool metrics
func (c *Collector) collectDatabaseMetrics() {
	sqlDB, err := c.db.DB()
	if err != nil {
		logger.Error(context.Background()).Err(err).Msg("Failed to get SQL database instance for metrics")
		return
	}

	stats := sqlDB.Stats()
	c.metrics.DBConnectionsActive.Set(float64(stats.InUse))
	c.metrics.DBConnectionsIdle.Set(float64(stats.Idle))
	c.metrics.DBConnectionsWaiting.Set(float64(stats.WaitCount))
}

// collectRedisMetrics collects Redis connection metrics
func (c *Collector) collectRedisMetrics(ctx context.Context) {
	stats := c.redisClient.PoolStats()
	// Redis v8 only reliably exposes IdleConns
	// TotalConns - IdleConns is NOT accurate for active connections in v8
	// Just track idle connections as a health indicator
	c.metrics.RedisConnectionsIdle.Set(float64(stats.IdleConns))
}

// collectTokenMetrics collects active token counts
func (c *Collector) collectTokenMetrics(ctx context.Context) {
	// Count active refresh tokens (not revoked and not expired)
	var refreshTokenCount int64
	if err := c.db.WithContext(ctx).
		Table("refresh_tokens").
		Where("revoked_at IS NULL AND expires_at > ?", time.Now()).
		Count(&refreshTokenCount).Error; err != nil {
		logger.Error(ctx).Err(err).Msg("Failed to count active refresh tokens")
	} else {
		c.metrics.RefreshTokensActive.Set(float64(refreshTokenCount))
	}

	// Note: Access tokens are JWTs and not stored in DB
	// Active count can be approximated by recent session activity or estimated from refresh tokens
}

// collectOrganizationMetrics collects organization-related metrics
func (c *Collector) collectOrganizationMetrics(ctx context.Context) {
	// Count total organizations
	var orgCount int64
	if err := c.db.WithContext(ctx).
		Table("organizations").
		Count(&orgCount).Error; err != nil {
		logger.Error(ctx).Err(err).Msg("Failed to count organizations")
	} else {
		c.metrics.OrganizationsTotal.Set(float64(orgCount))
	}

	// Count total organization members
	var memberCount int64
	if err := c.db.WithContext(ctx).
		Table("organization_memberships").
		Count(&memberCount).Error; err != nil {
		logger.Error(ctx).Err(err).Msg("Failed to count organization members")
	} else {
		c.metrics.OrgMembersTotal.Set(float64(memberCount))
	}
}

// collectSessionMetrics collects session-related metrics
func (c *Collector) collectSessionMetrics(ctx context.Context) {
	// Count active sessions (refresh tokens that are not revoked and not expired)
	var activeSessionCount int64
	if err := c.db.WithContext(ctx).
		Table("refresh_tokens").
		Where("revoked_at IS NULL AND expires_at > ?", time.Now()).
		Count(&activeSessionCount).Error; err != nil {
		logger.Error(ctx).Err(err).Msg("Failed to count active sessions")
	} else {
		c.metrics.ActiveSessions.Set(float64(activeSessionCount))
	}
}

// collectAPIKeyMetrics collects API key metrics
func (c *Collector) collectAPIKeyMetrics(ctx context.Context) {
	// Count active API keys (not revoked and not expired)
	var apiKeyCount int64
	if err := c.db.WithContext(ctx).
		Table("api_keys").
		Where("(revoked = false OR revoked IS NULL) AND (expires_at IS NULL OR expires_at > ?)", time.Now()).
		Count(&apiKeyCount).Error; err != nil {
		logger.Error(ctx).Err(err).Msg("Failed to count active API keys")
	} else {
		c.metrics.APIKeysActive.Set(float64(apiKeyCount))
	}
}
