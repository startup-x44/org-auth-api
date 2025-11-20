package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db          *sql.DB
	redisClient *redis.Client
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:          db,
		redisClient: redisClient,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status       string            `json:"status"`
	Timestamp    string            `json:"timestamp"`
	Version      string            `json:"version,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	Details      map[string]any    `json:"details,omitempty"`
}

// LivenessProbe handles Kubernetes liveness probe
// Returns 200 if the service is alive (not deadlocked/crashed)
// Kubernetes will restart the pod if this fails
//
// GET /health/live
func (h *HealthHandler) LivenessProbe(c *gin.Context) {
	// Liveness probe only checks if the application is running
	// It should NOT check dependencies (DB, Redis, etc.)
	// Those are checked in readiness probe
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "alive",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadinessProbe handles Kubernetes readiness probe
// Returns 200 if the service is ready to accept traffic
// Kubernetes will remove pod from service if this fails
//
// GET /health/ready
func (h *HealthHandler) ReadinessProbe(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	status := "ready"
	statusCode := http.StatusOK
	dependencies := make(map[string]string)
	details := make(map[string]any)

	// Check database
	dbStatus, dbDetails := h.checkDatabase(ctx)
	dependencies["database"] = dbStatus
	if dbStatus != "healthy" {
		status = "not_ready"
		statusCode = http.StatusServiceUnavailable
		details["database"] = dbDetails
	}

	// Check Redis
	redisStatus, redisDetails := h.checkRedis(ctx)
	dependencies["redis"] = redisStatus
	if redisStatus != "healthy" {
		status = "not_ready"
		statusCode = http.StatusServiceUnavailable
		details["redis"] = redisDetails
	}

	response := HealthResponse{
		Status:       status,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Dependencies: dependencies,
	}

	if len(details) > 0 {
		response.Details = details
	}

	c.JSON(statusCode, response)
}

// HealthCheck handles comprehensive health check (legacy endpoint)
// Provides detailed health information for all dependencies
//
// GET /health
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	status := "healthy"
	statusCode := http.StatusOK
	dependencies := make(map[string]string)
	details := make(map[string]any)

	// Check database
	dbStatus, dbDetails := h.checkDatabase(ctx)
	dependencies["database"] = dbStatus
	details["database"] = dbDetails
	if dbStatus != "healthy" {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	// Check Redis
	redisStatus, redisDetails := h.checkRedis(ctx)
	dependencies["redis"] = redisStatus
	details["redis"] = redisDetails
	if redisStatus != "healthy" {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthResponse{
		Status:       status,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Version:      getVersion(),
		Dependencies: dependencies,
		Details:      details,
	})
}

// checkDatabase performs database health check
func (h *HealthHandler) checkDatabase(ctx context.Context) (string, map[string]any) {
	details := make(map[string]any)
	start := time.Now()

	// Check connection
	if err := h.db.PingContext(ctx); err != nil {
		details["error"] = err.Error()
		details["latency_ms"] = time.Since(start).Milliseconds()
		return "unhealthy", details
	}

	// Get database stats
	stats := h.db.Stats()
	details["open_connections"] = stats.OpenConnections
	details["in_use"] = stats.InUse
	details["idle"] = stats.Idle
	details["wait_count"] = stats.WaitCount
	details["wait_duration_ms"] = stats.WaitDuration.Milliseconds()
	details["max_idle_closed"] = stats.MaxIdleClosed
	details["max_lifetime_closed"] = stats.MaxLifetimeClosed
	details["latency_ms"] = time.Since(start).Milliseconds()

	// Check if connections are available
	if stats.OpenConnections >= stats.MaxOpenConnections && stats.MaxOpenConnections > 0 {
		details["warning"] = "connection pool exhausted"
		return "degraded", details
	}

	return "healthy", details
}

// checkRedis performs Redis health check
func (h *HealthHandler) checkRedis(ctx context.Context) (string, map[string]any) {
	details := make(map[string]any)
	start := time.Now()

	// Check connection
	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		details["error"] = err.Error()
		details["latency_ms"] = time.Since(start).Milliseconds()
		return "unhealthy", details
	}

	// Get Redis pool stats
	poolStats := h.redisClient.PoolStats()
	details["hits"] = poolStats.Hits
	details["misses"] = poolStats.Misses
	details["timeouts"] = poolStats.Timeouts
	details["total_conns"] = poolStats.TotalConns
	details["idle_conns"] = poolStats.IdleConns
	details["stale_conns"] = poolStats.StaleConns
	details["latency_ms"] = time.Since(start).Milliseconds()

	// Check for timeout issues
	if poolStats.Timeouts > 100 {
		details["warning"] = "high timeout count"
		return "degraded", details
	}

	return "healthy", details
}

// serviceVersion can be overridden via ldflags:
// go build -ldflags "-X 'auth-service/internal/handler.serviceVersion=1.2.3'"
var serviceVersion = "1.0.0-dev"

func getVersion() string {
	return serviceVersion
}
