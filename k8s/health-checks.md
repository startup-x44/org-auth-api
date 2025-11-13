# Health Check Endpoints Configuration
# The application should implement these endpoints for Kubernetes probes

# Liveness Probe: /health/live
# - Should return 200 OK when the application is running
# - Should return 500 when the application is unhealthy
# - Should check basic application health (not database connectivity)

# Readiness Probe: /health/ready
# - Should return 200 OK when the application is ready to serve traffic
# - Should return 503 when the application is not ready
# - Should check dependencies (database, redis, external services)

# Metrics Endpoint: /metrics
# - Should expose Prometheus metrics
# - Should include custom business metrics
# - Should include standard Go metrics

# Example implementation in Gin:
#
# router.GET("/health/live", func(c *gin.Context) {
#     c.JSON(200, gin.H{"status": "ok"})
# })
#
# router.GET("/health/ready", func(c *gin.Context) {
#     // Check database connectivity
#     if dbHealth := checkDatabaseHealth(); !dbHealth {
#         c.JSON(503, gin.H{"status": "not ready", "reason": "database"})
#         return
#     }
#     // Check Redis connectivity
#     if redisHealth := checkRedisHealth(); !redisHealth {
#         c.JSON(503, gin.H{"status": "not ready", "reason": "redis"})
#         return
#     }
#     c.JSON(200, gin.H{"status": "ready"})
# })
#
# router.GET("/metrics", gin.WrapH(promhttp.Handler()))