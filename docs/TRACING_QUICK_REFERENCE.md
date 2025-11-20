# Distributed Tracing Quick Reference

## Overview

The auth-service uses OpenTelemetry for distributed tracing with automatic instrumentation for:
- ✅ HTTP requests (Gin)
- ✅ Database queries (GORM)
- ✅ Redis commands
- ✅ Manual business logic spans

---

## Configuration

### Environment Variables

```bash
# Enable/disable tracing
TRACING_ENABLED=true

# Service name (appears in traces)
TRACING_SERVICE_NAME=auth-service

# Exporter type: "otlp" or "stdout"
TRACING_EXPORTER=otlp

# OTLP endpoint (for "otlp" exporter)
TRACING_OTLP_ENDPOINT=localhost:4317

# TLS setting (MUST be false in production)
TRACING_OTLP_INSECURE=false

# Sampling rate: 0.0 to 1.0 (1.0 = 100%, 0.1 = 10%)
TRACING_SAMPLING_RATE=0.1
```

---

## Automatic Instrumentation

### HTTP Requests (via Gin middleware)

All HTTP requests are automatically traced with:
- Request method, path, status code
- X-Trace-ID response header
- Error recording (5xx only, not 4xx)
- **Excluded paths**: `/health`, `/health/live`, `/health/ready`, `/metrics`

### Database Queries (via GORM plugin)

All database operations are automatically traced with:
- SQL query text
- Table name
- Execution time
- Row count

### Redis Commands (via Redis hook)

All Redis operations are automatically traced with:
- Command name
- Key(s)
- Execution time

---

## Manual Instrumentation

### Starting Custom Spans

```go
import "auth-service/pkg/tracing"

// Database operation
ctx, span := tracing.StartDBSpan(ctx, "select", "users")
defer span.End()

// Redis operation
ctx, span := tracing.StartRedisSpan(ctx, "get", "user:123")
defer span.End()

// Service operation
ctx, span := tracing.StartServiceSpan(ctx, "email", "send_verification")
defer span.End()
```

### Recording Success/Errors

```go
// Record success
tracing.RecordSuccess(span)

// Record error
if err != nil {
    tracing.RecordError(span, err)
    return err
}
```

### Adding Custom Attributes

```go
import "go.opentelemetry.io/otel/attribute"

tracing.AddAttributes(span,
    attribute.String("user_id", userID),
    attribute.String("org_id", orgID),
    attribute.Int("count", 42),
)
```

---

## Middleware Order (CRITICAL)

```go
router.Use(gin.Recovery())                       // 1. Panic recovery (MUST be first)
router.Use(middleware.TracingMiddleware(...))    // 2. Tracing (wraps everything)
router.Use(/* HTTP status/error recording */)    // 3. Span enrichment
router.Use(middleware.SecurityHeadersMiddleware()) // 4. Security headers
router.Use(middleware.StructuredLoggingMiddleware())
router.Use(middleware.MetricsMiddleware())
router.Use(middleware.CORSMiddleware(...))
router.Use(revocationMiddleware)
router.Use(middleware.CSRFMiddleware(...))
```

**Do NOT**:
- ❌ Add multiple recovery middlewares
- ❌ Put tracing after other middlewares
- ❌ Trace health check endpoints
- ❌ Use insecure OTLP in production

---

## Trace Context Propagation

### Outgoing HTTP Requests

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Wrap HTTP client with tracing
client := &http.Client{
    Transport: otelhttp.NewTransport(http.DefaultTransport),
}

// Context is automatically propagated
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
```

### Getting Trace ID

```go
import "go.opentelemetry.io/otel/trace"

span := trace.SpanFromContext(ctx)
traceID := span.SpanContext().TraceID().String()
```

---

## Sampling Strategies

### Production (10% sampling)
```bash
TRACING_SAMPLING_RATE=0.1
```
- Lower cost
- Still captures errors
- Good for high-traffic services

### Staging (50% sampling)
```bash
TRACING_SAMPLING_RATE=0.5
```
- Balanced visibility
- Moderate cost

### Development (100% sampling)
```bash
TRACING_SAMPLING_RATE=1.0
```
- Full visibility
- All requests traced

### Testing (disabled)
```bash
TRACING_ENABLED=false
```
- No overhead
- Faster tests

---

## Common Trace Queries

### Find slow requests (Jaeger/Tempo)
```
service.name=auth-service AND duration > 1000ms
```

### Find errors
```
service.name=auth-service AND error=true
```

### Find specific endpoint
```
service.name=auth-service AND http.route=/api/v1/auth/login
```

### Find database queries
```
service.name=auth-service AND db.system=postgresql
```

### Find Redis operations
```
service.name=auth-service AND db.system=redis
```

---

## Security Notes

### Production Requirements

✅ **MUST**:
- Set `TRACING_OTLP_INSECURE=false`
- Use TLS for OTLP endpoint
- Sample appropriately (0.1 recommended)
- Exclude health/metrics endpoints

❌ **NEVER**:
- Use `TRACING_OTLP_INSECURE=true` in production (will FATAL)
- Trace health checks (causes massive noise)
- Log sensitive data in span attributes
- Sample at 100% in high-traffic production

---

## Troubleshooting

### No traces appearing

1. Check `TRACING_ENABLED=true`
2. Verify OTLP endpoint is reachable
3. Check sampling rate > 0
4. Verify collector is running

### Missing trace IDs in logs

1. Check middleware order (tracing must be early)
2. Verify X-Trace-ID header is present
3. Check span is recording: `span.IsRecording()`

### Too many traces

1. Reduce sampling rate
2. Verify health endpoints are excluded
3. Check for trace loops

### Build fails with "insecure OTLP" error

This is correct behavior in production. Set:
```bash
TRACING_OTLP_INSECURE=false
```

Or use stdout exporter for local dev:
```bash
TRACING_EXPORTER=stdout
```

---

## Performance Impact

| Component          | Overhead | Notes                          |
| ------------------ | -------- | ------------------------------ |
| HTTP tracing       | ~1-2ms   | Per request                    |
| DB tracing         | <0.5ms   | Per query                      |
| Redis tracing      | <0.1ms   | Per command                    |
| Sampling @ 10%     | Minimal  | 90% of requests skip exporter |
| Health checks      | 0ms      | Excluded from tracing          |

---

## Integration with Monitoring

### Trace → Metrics

Use trace data to generate RED metrics:
- **R**ate: requests per second
- **E**rrors: error rate
- **D**uration: latency percentiles

### Trace → Logs

Correlation via trace ID:
```json
{
  "level": "error",
  "msg": "database query failed",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7"
}
```

### Trace → Alerts

Alert on trace-derived metrics:
- P99 latency > 1s
- Error rate > 5%
- Endpoint availability < 99%

---

## Links

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Gin OTel Integration](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/github.com/gin-gonic/gin/otelgin)
- [GORM OTel Plugin](https://github.com/go-gorm/opentelemetry)
- [Critical Fixes Applied](./TRACING_CRITICAL_FIXES.md)

---

**Last Updated**: November 18, 2025  
**Status**: Production Ready ✅
