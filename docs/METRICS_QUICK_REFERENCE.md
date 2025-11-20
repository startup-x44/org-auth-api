# Prometheus Metrics - Quick Reference

## Accessing Metrics

```bash
# Development
curl http://localhost:8080/metrics

# Production
curl https://auth-service.example.com/metrics
```

## Common Metrics to Monitor

### Service Health
```promql
# Overall request rate
sum(rate(http_requests_total[5m]))

# Error rate percentage
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100

# p95 response time
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path))
```

### Authentication
```promql
# Login success rate
sum(rate(auth_success_total{type="login"}[5m])) / sum(rate(auth_attempts_total{type="login"}[5m])) * 100

# Failed login reasons
sum(rate(auth_failures_total[5m])) by (reason)

# Registration rate
sum(rate(registrations_total{status="success"}[5m]))
```

### Tokens
```promql
# Active sessions
sessions_active

# Token refresh success rate
sum(rate(token_refreshes_total{status="success"}[5m])) / sum(rate(token_refreshes_total[5m])) * 100

# Token revocation rate
sum(rate(token_revocations_total[5m])) by (scope)
```

### Database
```promql
# Connection pool usage
db_connections_active / (db_connections_active + db_connections_idle) * 100

# Slow queries (p99 > 100ms)
histogram_quantile(0.99, sum(rate(db_query_duration_seconds_bucket[5m])) by (le, table)) > 0.1

# Query errors
sum(rate(db_queries_total{status="error"}[5m])) by (table)
```

### RBAC
```promql
# Permission denial rate
sum(rate(permission_checks_total{result="denied"}[5m])) / sum(rate(permission_checks_total[5m])) * 100

# Most denied permissions
topk(10, sum(rate(authorization_denials_total[1h])) by (permission))
```

## Recording Custom Metrics

### In Your Handler
```go
import "auth-service/pkg/metrics"

func (h *MyHandler) MyEndpoint(c *gin.Context) {
    m := metrics.GetMetrics()
    
    // Record authentication
    success := true
    reason := ""
    m.RecordAuthAttempt("login", orgID, success, reason)
    
    // Record permission check
    granted := true
    m.RecordPermissionCheck("role:create", orgID, granted)
    
    // Record token issuance
    m.RecordTokenIssuance("access", orgID)
    
    // Record error
    m.RecordError("validation", "my_handler")
}
```

### Database Instrumentation
```go
start := time.Now()
err := db.Find(&users).Error
m.RecordDBQuery("SELECT", "users", time.Since(start), err)
```

### Redis Instrumentation
```go
start := time.Now()
err := redisClient.Get(ctx, key).Err()
m.RecordRedisOperation("GET", time.Since(start), err)
```

## Critical Alerts

### High Error Rate
```yaml
alert: HighErrorRate
expr: sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > 0.05
for: 5m
```

### Database Connection Pool Exhaustion
```yaml
alert: DatabaseConnectionPoolExhaustion
expr: db_connections_waiting > 10
for: 2m
```

### High Authentication Failures
```yaml
alert: HighAuthFailureRate
expr: sum(rate(auth_failures_total[5m])) / sum(rate(auth_attempts_total[5m])) > 0.3
for: 10m
```

## Grafana Dashboard Queries

### Request Rate Panel
```promql
sum(rate(http_requests_total[5m])) by (path)
```

### Latency Panel (p50, p95, p99)
```promql
histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
```

### Active Sessions Gauge
```promql
sessions_active
```

### Database Connections
```promql
db_connections_active
db_connections_idle
db_connections_waiting
```

## Troubleshooting

### No metrics appearing?
1. Check `/metrics` endpoint is accessible
2. Verify Prometheus scrape config points to correct target
3. Check service logs for metrics initialization errors

### High cardinality warning?
1. Review label usage - avoid user IDs, emails in labels
2. Check for dynamic metric creation
3. Limit label value count to < 100 per label

### Metrics endpoint slow?
1. Check label cardinality
2. Review histogram bucket configuration
3. Consider sampling for high-volume metrics

## See Also
- Full documentation: `docs/PROMETHEUS_METRICS.md`
- Implementation summary: `docs/PROMETHEUS_METRICS_SUMMARY.md`
- Code: `pkg/metrics/metrics.go`, `pkg/metrics/collector.go`
