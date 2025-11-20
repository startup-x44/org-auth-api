# Prometheus Metrics

This document describes the Prometheus metrics implementation for the auth-service, providing comprehensive observability into authentication, authorization, token management, and system performance.

## Overview

The auth-service exposes detailed metrics via the `/metrics` endpoint in Prometheus format. These metrics enable monitoring of:

- HTTP request patterns and performance
- Authentication success/failure rates
- Token lifecycle and operations
- OAuth2 flows
- RBAC permission checks
- Database and Redis performance
- Rate limiting effectiveness
- Audit logging operations
- System resource usage

## Metrics Endpoint

**Endpoint:** `GET /metrics`

**Access:** Public (consider restricting in production via network policies or authentication)

**Format:** Prometheus text-based exposition format

**Example:**
```bash
curl http://localhost:8080/metrics
```

## Metrics Categories

### 1. HTTP Request Metrics

#### `http_requests_total`
**Type:** Counter  
**Labels:** `method`, `path`, `status`  
**Description:** Total number of HTTP requests processed

**Example:**
```
http_requests_total{method="POST",path="/api/v1/auth/login",status="200"} 1523
http_requests_total{method="POST",path="/api/v1/auth/login",status="401"} 47
```

#### `http_request_duration_seconds`
**Type:** Histogram  
**Labels:** `method`, `path`, `status`  
**Buckets:** 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10  
**Description:** HTTP request latency distribution

**Usage:** Calculate p95, p99 latencies:
```promql
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path))
```

#### `http_request_size_bytes`
**Type:** Histogram  
**Labels:** `method`, `path`  
**Description:** HTTP request body size distribution

#### `http_response_size_bytes`
**Type:** Histogram  
**Labels:** `method`, `path`  
**Description:** HTTP response body size distribution

### 2. Authentication Metrics

#### `auth_attempts_total`
**Type:** Counter  
**Labels:** `type`, `organization`  
**Description:** Total authentication attempts  
**Types:** `login`, `oauth2`, `api_key`, `refresh`

#### `auth_success_total`
**Type:** Counter  
**Labels:** `type`, `organization`  
**Description:** Successful authentications

**Alert Example:**
```yaml
# Low authentication success rate
- alert: LowAuthSuccessRate
  expr: |
    rate(auth_success_total[5m]) / rate(auth_attempts_total[5m]) < 0.5
  for: 5m
  annotations:
    summary: "Auth success rate below 50%"
```

#### `auth_failures_total`
**Type:** Counter  
**Labels:** `type`, `reason`, `organization`  
**Description:** Failed authentication attempts  
**Reasons:** `invalid_credentials`, `account_disabled`, `email_not_verified`, `rate_limited`

#### `registrations_total`
**Type:** Counter  
**Labels:** `status`  
**Description:** User registration attempts  
**Status:** `success`, `error`

#### `password_resets_total`
**Type:** Counter  
**Labels:** `status`  
**Description:** Password reset requests

#### `email_verifications_total`
**Type:** Counter  
**Labels:** `status`  
**Description:** Email verification attempts

### 3. Token Metrics

#### `tokens_issued_total`
**Type:** Counter  
**Labels:** `type`, `organization`  
**Description:** Tokens issued  
**Types:** `access`, `refresh`, `authorization_code`

#### `token_refreshes_total`
**Type:** Counter  
**Labels:** `status`, `organization`  
**Description:** Token refresh operations  
**Status:** `success`, `error`, `revoked`, `expired`

#### `token_revocations_total`
**Type:** Counter  
**Labels:** `type`, `scope`  
**Description:** Token revocations  
**Types:** `access`, `refresh`  
**Scopes:** `user`, `organization`, `single`

#### `token_validations_total`
**Type:** Counter  
**Labels:** `status`, `reason`  
**Description:** Token validation checks  
**Status:** `valid`, `invalid`  
**Reasons:** `expired`, `revoked`, `malformed`, `signature_invalid`

#### `access_tokens_active`
**Type:** Gauge  
**Description:** Current number of active access tokens (estimated)

#### `refresh_tokens_active`
**Type:** Gauge  
**Description:** Current number of active refresh tokens

### 4. OAuth2 Metrics

#### `oauth2_authorizations_total`
**Type:** Counter  
**Labels:** `client_id`, `status`  
**Description:** OAuth2 authorization requests

#### `oauth2_token_grants_total`
**Type:** Counter  
**Labels:** `grant_type`, `client_id`, `status`  
**Description:** OAuth2 token grants  
**Grant Types:** `authorization_code`, `refresh_token`, `client_credentials`

#### `oauth2_token_refreshes_total`
**Type:** Counter  
**Labels:** `client_id`, `status`  
**Description:** OAuth2 token refresh operations

#### `oauth2_flow_duration_seconds`
**Type:** Histogram  
**Labels:** `flow_type`, `status`  
**Buckets:** 0.1, 0.25, 0.5, 1, 2.5, 5, 10  
**Description:** OAuth2 flow completion time

### 5. RBAC Metrics

#### `permission_checks_total`
**Type:** Counter  
**Labels:** `permission`, `organization`, `result`  
**Description:** Permission authorization checks  
**Results:** `granted`, `denied`

**Example Query:**
```promql
# Permission denial rate by permission
sum(rate(permission_checks_total{result="denied"}[5m])) by (permission)
```

#### `role_assignments_total`
**Type:** Counter  
**Labels:** `role`, `organization`, `action`  
**Description:** Role assignment operations  
**Actions:** `assign`, `revoke`

#### `permission_grants_total`
**Type:** Counter  
**Labels:** `permission`, `organization`, `action`  
**Description:** Permission grant/revoke operations

#### `authorization_denials_total`
**Type:** Counter  
**Labels:** `permission`, `organization`, `reason`  
**Description:** Authorization denials with reasons

### 6. Session Metrics

#### `sessions_active`
**Type:** Gauge  
**Description:** Current number of active user sessions

#### `session_creations_total`
**Type:** Counter  
**Labels:** `organization`  
**Description:** Sessions created

#### `session_destroyed_total`
**Type:** Counter  
**Labels:** `reason`, `organization`  
**Description:** Sessions destroyed  
**Reasons:** `logout`, `timeout`, `revocation`, `admin_action`

#### `concurrent_sessions`
**Type:** Histogram  
**Labels:** `organization`  
**Buckets:** 1, 2, 3, 5, 10, 20, 50  
**Description:** Distribution of concurrent sessions per user

### 7. Rate Limiting Metrics

#### `rate_limit_hits_total`
**Type:** Counter  
**Labels:** `scope`, `result`  
**Description:** Rate limit check results  
**Scopes:** `login`, `registration`, `password_reset`, `token_refresh`, `oauth2_token`, `api_calls`  
**Results:** `allowed`, `blocked`

#### `rate_limit_blocks_total`
**Type:** Counter  
**Labels:** `scope`, `identifier`  
**Description:** Requests blocked by rate limiting

**Alert Example:**
```yaml
- alert: HighRateLimitBlocks
  expr: rate(rate_limit_blocks_total[5m]) > 100
  for: 5m
  annotations:
    summary: "Unusually high rate limit blocks"
```

### 8. Database Metrics

#### `db_queries_total`
**Type:** Counter  
**Labels:** `operation`, `table`, `status`  
**Description:** Database queries executed  
**Operations:** `SELECT`, `INSERT`, `UPDATE`, `DELETE`

#### `db_query_duration_seconds`
**Type:** Histogram  
**Labels:** `operation`, `table`  
**Buckets:** 0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1  
**Description:** Database query execution time

**Slow Query Detection:**
```promql
histogram_quantile(0.99, sum(rate(db_query_duration_seconds_bucket[5m])) by (le, table)) > 0.1
```

#### `db_connections_active`
**Type:** Gauge  
**Description:** Active database connections

#### `db_connections_idle`
**Type:** Gauge  
**Description:** Idle database connections in pool

#### `db_connections_waiting`
**Type:** Gauge  
**Description:** Connections waiting for available connection

**Alert Example:**
```yaml
- alert: DatabaseConnectionPoolExhaustion
  expr: db_connections_waiting > 10
  for: 2m
  annotations:
    summary: "DB connection pool under pressure"
```

### 9. Redis Metrics

#### `redis_operations_total`
**Type:** Counter  
**Labels:** `operation`, `status`  
**Description:** Redis operations executed  
**Operations:** `GET`, `SET`, `DEL`, `EXPIRE`, `INCR`, `ZADD`

#### `redis_operation_duration_seconds`
**Type:** Histogram  
**Labels:** `operation`  
**Buckets:** 0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1  
**Description:** Redis operation execution time

#### `redis_connections_active`
**Type:** Gauge  
**Description:** Active Redis connections

### 10. Audit Logging Metrics

#### `audit_logs_total`
**Type:** Counter  
**Labels:** `action`, `resource`, `status`  
**Description:** Audit log entries created  
**Actions:** See `internal/models/audit_log.go` for action constants  
**Resources:** `user`, `role`, `permission`, `organization`, `session`, `oauth`, `api_key`

#### `audit_log_write_duration_seconds`
**Type:** Histogram  
**Labels:** `destination`  
**Buckets:** 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5  
**Description:** Audit log write latency  
**Destinations:** `database`, `structured_log`

### 11. Organization Metrics

#### `organizations_total`
**Type:** Gauge  
**Description:** Total number of organizations

#### `org_members_total`
**Type:** Gauge  
**Description:** Total number of organization members

#### `org_invitations_total`
**Type:** Counter  
**Labels:** `status`, `organization`  
**Description:** Organization invitation operations  
**Status:** `sent`, `accepted`, `cancelled`, `expired`

### 12. API Key Metrics

#### `api_keys_active`
**Type:** Gauge  
**Description:** Current number of active API keys

#### `api_key_validations_total`
**Type:** Counter  
**Labels:** `status`  
**Description:** API key validation attempts  
**Status:** `valid`, `invalid`, `expired`, `revoked`

### 13. Error Metrics

#### `errors_total`
**Type:** Counter  
**Labels:** `type`, `component`  
**Description:** Errors by type and component  
**Types:** `database`, `redis`, `validation`, `authentication`, `authorization`, `internal`  
**Components:** `auth_handler`, `role_handler`, `oauth_handler`, etc.

#### `panics_recovered_total`
**Type:** Counter  
**Description:** Total number of panics recovered by middleware

## Metrics Collection

### Automatic Collection

The `MetricsMiddleware` automatically collects HTTP request metrics for all endpoints:

```go
router.Use(middleware.MetricsMiddleware())
```

### Periodic Collection

The `MetricsCollector` runs every 30 seconds to update gauge metrics:

- Database connection pool stats
- Redis connection stats
- Active token counts
- Organization/member counts
- Active session counts
- API key counts

```go
metricsCollector := metrics.NewCollector(db, redisClient, 30*time.Second)
go metricsCollector.Start(context.Background())
```

### Manual Instrumentation

Components can record custom metrics using the metrics package:

```go
import "auth-service/pkg/metrics"

m := metrics.GetMetrics()

// Record authentication attempt
m.RecordAuthAttempt("login", orgID, success, reason)

// Record token issuance
m.RecordTokenIssuance("access", orgID)

// Record permission check
m.RecordPermissionCheck("role:create", orgID, granted)

// Record database query
m.RecordDBQuery("SELECT", "users", duration, err)
```

## Sample Queries

### Authentication Performance

```promql
# Authentication success rate
sum(rate(auth_success_total[5m])) / sum(rate(auth_attempts_total[5m]))

# Login p95 latency
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{path="/api/v1/auth/login"}[5m])) by (le))

# Failed login attempts by reason
sum(rate(auth_failures_total[5m])) by (reason)
```

### Token Operations

```promql
# Token refresh rate
sum(rate(token_refreshes_total[5m])) by (status)

# Active sessions trend
sessions_active

# Token revocation rate by scope
sum(rate(token_revocations_total[5m])) by (scope)
```

### System Health

```promql
# Request rate by endpoint
sum(rate(http_requests_total[5m])) by (path)

# Error rate
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))

# Database connection pool utilization
db_connections_active / (db_connections_active + db_connections_idle)
```

### RBAC Analysis

```promql
# Top denied permissions
topk(10, sum(rate(authorization_denials_total[1h])) by (permission))

# Permission check success rate by organization
sum(rate(permission_checks_total{result="granted"}[5m])) by (organization) 
  / sum(rate(permission_checks_total[5m])) by (organization)
```

## Alerting Rules

### Critical Alerts

```yaml
groups:
  - name: auth_service_critical
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m])) 
            / sum(rate(http_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Error rate above 5%"

      - alert: DatabaseConnectionPoolExhaustion
        expr: db_connections_waiting > 10
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Database connection pool saturated"

      - alert: HighAuthFailureRate
        expr: |
          sum(rate(auth_failures_total[5m])) 
            / sum(rate(auth_attempts_total[5m])) > 0.3
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Auth failure rate above 30%"
```

### Performance Alerts

```yaml
  - name: auth_service_performance
    rules:
      - alert: HighLatency
        expr: |
          histogram_quantile(0.95, 
            sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path)
          ) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "p95 latency above 1 second"

      - alert: SlowDatabaseQueries
        expr: |
          histogram_quantile(0.99, 
            sum(rate(db_query_duration_seconds_bucket[5m])) by (le, table)
          ) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Slow database queries detected"
```

### Security Alerts

```yaml
  - name: auth_service_security
    rules:
      - alert: SuspiciousLoginFailures
        expr: |
          sum(rate(auth_failures_total{reason="invalid_credentials"}[5m])) > 50
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Unusually high login failure rate"

      - alert: HighRateLimitBlocks
        expr: sum(rate(rate_limit_blocks_total[5m])) > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High rate of rate-limited requests"

      - alert: UnusualTokenRevocations
        expr: sum(rate(token_revocations_total[5m])) > 10
        for: 5m
        labels:
          severity: info
        annotations:
          summary: "Elevated token revocation rate"
```

## Grafana Dashboards

### Sample Dashboard Panels

**1. Overview Panel**
- Request rate (requests/sec)
- Error rate (%)
- p95 latency (ms)
- Active sessions

**2. Authentication Panel**
- Login success rate
- Registration rate
- Password reset requests
- Email verification rate

**3. Token Management Panel**
- Token issuance rate by type
- Token refresh success rate
- Active tokens (access/refresh)
- Token revocations

**4. System Resources Panel**
- Database connections (active/idle/waiting)
- Redis connections
- Database query latency (p50, p95, p99)
- Redis operation latency

**5. Security Panel**
- Failed authentication attempts by reason
- Rate limit blocks by scope
- Authorization denials by permission
- API key validation failures

## Integration with Monitoring Stack

### Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'auth-service'
    scrape_interval: 15s
    static_configs:
      - targets: ['auth-service:8080']
    metrics_path: /metrics
```

### Kubernetes Service Monitor

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: auth-service
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: auth-service
  endpoints:
    - port: http
      path: /metrics
      interval: 15s
```

## Security Considerations

### Metrics Endpoint Protection

In production, consider:

1. **Network-level restriction:** Use Kubernetes NetworkPolicy or firewall rules
2. **Authentication:** Add basic auth or mTLS
3. **Label cardinality:** Avoid high-cardinality labels (e.g., user IDs, email addresses)
4. **Sensitive data:** Never include PII in metric labels or values

### Example NetworkPolicy

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-prometheus-to-auth-service
spec:
  podSelector:
    matchLabels:
      app: auth-service
  ingress:
    - from:
      - namespaceSelector:
          matchLabels:
            name: monitoring
      ports:
      - protocol: TCP
        port: 8080
```

## Performance Impact

### Metrics Collection Overhead

- **HTTP middleware:** ~0.1-0.5ms per request
- **Histogram observations:** ~10-50μs
- **Counter increments:** ~5-10μs
- **Gauge updates:** ~5-10μs
- **Periodic collector:** Minimal (runs every 30s)

### Memory Usage

- Estimated: 2-5 MB for metric storage (depends on label cardinality)
- Histogram buckets: Pre-allocated, fixed memory

### Best Practices

1. Keep label cardinality low (< 10 values per label)
2. Use appropriate metric types (Counter vs Gauge vs Histogram)
3. Avoid creating metrics dynamically at runtime
4. Use histogram buckets aligned with SLOs
5. Monitor scrape duration: should be < 1 second

## Next Steps

After implementing Prometheus metrics, consider:

1. **Task #13:** Health check endpoints (`/health/live`, `/health/ready`)
2. **Task #14:** Distributed tracing with OpenTelemetry
3. **Task #31:** Configure alerting rules in Prometheus/Alertmanager
4. Set up Grafana dashboards for visualization
5. Configure Prometheus federation for multi-cluster deployments

## References

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)
- [Metric Types](https://prometheus.io/docs/concepts/metric_types/)
- [PromQL Queries](https://prometheus.io/docs/prometheus/latest/querying/basics/)
