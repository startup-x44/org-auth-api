# Structured Logging Guide

## Overview

The auth-service uses **zerolog** for structured JSON logging with automatic context propagation across the request lifecycle.

## Features

- **JSON Format**: All logs are structured JSON for easy parsing and aggregation
- **Context Fields**: Automatic propagation of `request_id`, `user_id`, `org_id`, `session_id`, `trace_id`
- **Log Levels**: Debug, Info, Warn, Error, Fatal, Panic
- **Request Tracking**: Every HTTP request gets a unique `request_id` (X-Request-ID header)
- **Performance**: Automatic latency tracking for all HTTP requests
- **User Context**: Authenticated requests include user, org, and session information

## Configuration

Set environment variables to control logging behavior:

```bash
LOG_LEVEL=info        # Options: debug, info, warn, error
LOG_FORMAT=json       # Options: json, console
ENVIRONMENT=production
```

## Automatic Context Fields

The middleware automatically adds these fields to all logs:

| Field | Source | Description |
|-------|--------|-------------|
| `request_id` | X-Request-ID header or generated UUID | Unique identifier for each request |
| `trace_id` | X-Trace-ID header | Distributed tracing ID |
| `user_id` | Auth middleware | Authenticated user's ID |
| `org_id` | Auth middleware | Current organization ID |
| `session_id` | Auth middleware | Active session ID |
| `client_id` | OAuth2 middleware | OAuth2 client application ID |
| `ip_address` | Client IP | Request origin IP |
| `user_agent` | User-Agent header | Client user agent string |

## Usage in Code

### Basic Logging

```go
import "auth-service/pkg/logger"

func MyHandler(c *gin.Context) {
    ctx := c.Request.Context()
    
    // Info log
    logger.Info(ctx).Msg("Processing request")
    
    // With additional fields
    logger.Info(ctx).
        Str("action", "user_created").
        Str("email", email).
        Msg("User created successfully")
    
    // Error log
    if err != nil {
        logger.Error(ctx).
            Err(err).
            Str("operation", "database_query").
            Msg("Failed to query database")
    }
}
```

### Adding Custom Context

```go
// Add custom context fields
ctx = logger.WithContext(ctx, logger.TraceIDKey, traceID)
ctx = logger.WithContext(ctx, "custom_key", "custom_value")
```

### Log Levels

```go
logger.Debug(ctx).Msg("Detailed debug information")
logger.Info(ctx).Msg("Informational message")
logger.Warn(ctx).Msg("Warning message")
logger.Error(ctx).Err(err).Msg("Error occurred")
logger.Fatal(ctx).Err(err).Msg("Fatal error - will exit")
```

### Global Logging (Without Context)

```go
logger.InfoMsg("Server started", map[string]interface{}{
    "port": 8080,
    "env": "production",
})

logger.ErrorMsg("Database connection failed", err, map[string]interface{}{
    "host": "localhost",
    "port": 5432,
})
```

## Log Output Examples

### HTTP Request Log
```json
{
  "level": "info",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "org_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "method": "POST",
  "path": "/api/v1/users",
  "status": 201,
  "latency": 45,
  "client_ip": "192.168.1.100",
  "time": "2025-11-18T10:30:45Z",
  "message": "HTTP request"
}
```

### Application Log
```json
{
  "level": "error",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "error": "connection refused",
  "operation": "send_email",
  "time": "2025-11-18T10:30:46Z",
  "message": "Failed to send verification email"
}
```

## Middleware Stack

The logging middleware is applied in this order:

1. **StructuredLoggingMiddleware()** - Generates request_id, logs all requests
2. **AddUserContextMiddleware()** - Adds user/org/session context (after auth)

## Best Practices

### ✅ Do

- Use context-aware logging: `logger.Info(ctx).Msg(...)`
- Add meaningful fields: `.Str("operation", "user_creation")`
- Log errors with `.Err(err)`
- Use appropriate log levels
- Include request context in all service/handler logs

### ❌ Don't

- Log sensitive data (passwords, tokens, PII)
- Use `fmt.Println()` or standard `log` package
- Log at DEBUG level in production
- Create new loggers per request (use context-based logging)

## Querying Logs

### With jq (local development)

```bash
# Filter by user_id
cat logs.json | jq 'select(.user_id == "123e4567")'

# Filter errors
cat logs.json | jq 'select(.level == "error")'

# Find slow requests (latency > 1000ms)
cat logs.json | jq 'select(.latency > 1000)'

# Trace a request
cat logs.json | jq 'select(.request_id == "550e8400-e29b-41d4")'
```

### With Log Aggregation Tools

- **CloudWatch Logs Insights**
- **Elasticsearch + Kibana**
- **Datadog**
- **Grafana Loki**

Example CloudWatch query:
```
fields @timestamp, message, request_id, user_id, level
| filter level = "error"
| filter org_id = "7c9e6679-7425-40de-944b-e07fc1f90ae7"
| sort @timestamp desc
| limit 100
```

## Performance

Zerolog is designed for high performance:
- Zero allocation for disabled log levels
- Fast JSON encoding
- Minimal overhead on hot paths
- Lazy evaluation of log fields

## Testing

Logs can be captured in tests for verification:

```go
func TestHandler(t *testing.T) {
    // Create test logger
    buf := &bytes.Buffer{}
    testLogger := logger.Config{
        Level: "debug",
        Format: "json",
        Output: buf,
    }
    logger.Initialize(&testLogger)
    
    // Run your test
    // ...
    
    // Verify logs
    logs := buf.String()
    assert.Contains(t, logs, "expected message")
}
```

## Troubleshooting

### No logs appearing

1. Check `LOG_LEVEL` environment variable
2. Ensure logger is initialized in `main.go`
3. Verify you're using context-aware logging

### Missing context fields

1. Ensure middleware is applied in correct order
2. Check that `AddUserContextMiddleware()` comes after auth
3. Verify context is being passed correctly

### Too many logs

1. Increase `LOG_LEVEL` to `warn` or `error` in production
2. Remove debug logs from hot paths
3. Use sampling for high-traffic endpoints
