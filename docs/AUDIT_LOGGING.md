# Audit Logging Documentation

## Overview

The auth-service implements comprehensive audit logging for all critical authentication and authorization operations. Audit logs provide a tamper-evident trail for compliance, security monitoring, and forensic analysis.

## Architecture

### Components

1. **Database Table** (`audit_logs`): Persistent storage for audit events
2. **AuditService**: Business logic for creating and querying audit logs
3. **AuditLogRepository**: Data access layer for audit_logs table
4. **Handler Integration**: Audit calls in HTTP handlers for key operations
5. **Structured Logger Integration**: Real-time audit events in application logs

### Dual Logging Strategy

Audit events are logged to **two destinations**:

1. **Database** (`audit_logs` table): Persistent, queryable records for compliance
2. **Structured Logs** (zerolog): Real-time visibility for monitoring and alerting

## Database Schema

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id UUID,
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_id VARCHAR(100),
    details JSONB,
    success BOOLEAN NOT NULL DEFAULT true,
    error TEXT,
    service VARCHAR(50) NOT NULL DEFAULT 'auth-service',
    method VARCHAR(200),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Indexed Fields

- `user_id`: Find all actions by a specific user
- `organization_id`: Find all actions within an organization
- `action`: Find all instances of a specific action type
- `resource`: Find all operations on a resource type
- `timestamp`, `created_at`: Time-based queries
- `request_id`: Correlate all audit events from a single HTTP request
- Composite: `(user_id, action, timestamp)` and `(organization_id, action, timestamp)`

## Audited Actions

### Authentication Events

| Action | Description | Resource | Details |
|--------|-------------|----------|---------|
| `login` | Successful user login | `auth` | email, organizations_count |
| `login_failed` | Failed login attempt | `auth` | email |
| `logout` | User logout | `auth` | session_id |
| `register` | New user registration | `auth` | email, first_name, last_name |
| `token_refresh` | JWT token refresh | `auth` | - |
| `password_change` | User changed password | `auth` | - |
| `password_reset` | Password reset via email | `auth` | email |
| `email_verification` | Email verified | `auth` | email |

### Authorization Events

| Action | Description | Resource | Details |
|--------|-------------|----------|---------|
| `role_create` | New role created | `role` | role_name, role_description |
| `role_update` | Role modified | `role` | changed_fields |
| `role_delete` | Role deleted | `role` | role_name |
| `role_assign` | Role assigned to user | `role` | target_user_id, role_id |
| `role_revoke` | Role removed from user | `role` | target_user_id, role_id |
| `permission_grant` | Permissions assigned to role | `permission` | role_id, permissions[] |
| `permission_revoke` | Permissions removed from role | `permission` | role_id, permissions[] |

### Organization Events

| Action | Description | Resource | Details |
|--------|-------------|----------|---------|
| `org_create` | Organization created | `organization` | org_name, org_slug |
| `org_update` | Organization modified | `organization` | changed_fields |
| `org_delete` | Organization deleted | `organization` | org_name |
| `member_invite` | User invited to organization | `member` | target_email, role |
| `member_remove` | User removed from organization | `member` | target_user_id |
| `member_update` | Member role/status changed | `member` | target_user_id, new_role |

### Session Events

| Action | Description | Resource | Details |
|--------|-------------|----------|---------|
| `session_create` | New session created | `session` | device_info |
| `session_revoke` | Session revoked | `session` | session_id |
| `session_revoke_all` | All user sessions revoked | `session` | reason |

### OAuth2 Events

| Action | Description | Resource | Details |
|--------|-------------|----------|---------|
| `oauth_authorize` | Authorization code granted | `oauth_client` | client_id, scopes |
| `oauth_token_grant` | Access token issued | `oauth_client` | client_id, grant_type |
| `oauth_token_revoke` | Token revoked | `token` | token_type |
| `client_create` | OAuth2 client created | `oauth_client` | client_name |
| `client_secret_rotate` | Client secret rotated | `oauth_client` | client_id |

## Usage Examples

### In Handlers

```go
// Example: Login with audit logging
func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    c.ShouldBindJSON(&req)
    
    response, err := h.authService.Login(c.Request.Context(), &req)
    
    // Audit log
    var userID *uuid.UUID
    if response != nil && response.User != nil {
        id, _ := uuid.Parse(response.User.ID)
        userID = &id
    }
    
    action := models.ActionLogin
    if err != nil {
        action = models.ActionLoginFailed
    }
    
    h.auditService.LogAuth(c.Request.Context(), action, userID, err == nil, map[string]interface{}{
        "email": req.Email,
    }, err)
    
    // ... return response
}
```

### Querying Audit Logs

```go
// Get all login attempts for a user
logs, err := auditService.GetUserAuditLogs(ctx, userID, 100, 0)

// Get all role changes in an organization
logs, err := auditService.GetOrganizationAuditLogs(ctx, orgID, 100, 0)

// Get all failed login attempts
logs, err := auditService.GetAuditLogsByAction(ctx, models.ActionLoginFailed, 100, 0)

// Get all audit events for a specific request
logs, err := auditService.GetAuditLogsByRequestID(ctx, requestID)
```

### Retention Policy

```go
// Clean up logs older than 90 days
deletedCount, err := auditService.CleanupOldLogs(ctx, 90)
```

## Context Propagation

Audit logs automatically capture context from the HTTP request:

- **request_id**: X-Request-ID header (generated if not present)
- **ip_address**: Client IP address
- **user_agent**: Browser/client user agent
- **user_id**: Authenticated user (from JWT)
- **organization_id**: Current organization context (from JWT)

These are extracted from the context by `StructuredLoggingMiddleware` and `AddUserContextMiddleware`.

## Structured Log Format

Each audit event also appears in structured logs with the `audit_` prefix:

```json
{
  "level": "info",
  "timestamp": "2025-11-18T11:45:23Z",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "ip_address": "192.168.1.100",
  "audit_action": "login",
  "audit_resource": "auth",
  "audit_success": true,
  "audit_details": {
    "email": "user@example.com",
    "organizations_count": 3
  },
  "message": "Audit event"
}
```

## Compliance Features

### Tamper Evidence

- **Immutable Logs**: Audit logs are insert-only (no updates or deletes except retention policy)
- **Foreign Key Protection**: `ON DELETE SET NULL` preserves audit trail even if user/org deleted
- **Timestamp Accuracy**: Uses PostgreSQL `NOW()` for server-side timestamps

### Data Privacy

- **Sensitive Data**: Passwords are NEVER logged
- **PII Handling**: Only necessary identifiers (email, names) are logged for accountability
- **Encryption**: Database connection uses SSL/TLS in production

### Retention

- **Default Retention**: Configurable via `CleanupOldLogs()`
- **Recommended**: 90 days for general logs, 1-2 years for compliance-critical events
- **Legal Hold**: Implement application-level flags to prevent deletion of specific records

## Monitoring & Alerting

### Key Metrics to Track

1. **Failed Login Rate**: Spike indicates brute force attack
2. **Permission Changes**: Monitor for unauthorized privilege escalation
3. **After-Hours Activity**: Unusual access patterns
4. **Multiple Failed Logins**: Same user or same IP
5. **Session Revocations**: Mass revocations may indicate compromise

### Query Examples

```sql
-- Failed logins in last hour
SELECT * FROM audit_logs 
WHERE action = 'login_failed' 
  AND timestamp > NOW() - INTERVAL '1 hour'
ORDER BY timestamp DESC;

-- Privilege escalations (permission grants)
SELECT * FROM audit_logs 
WHERE action IN ('permission_grant', 'role_assign')
  AND timestamp > NOW() - INTERVAL '24 hours'
ORDER BY timestamp DESC;

-- Activity by a specific user
SELECT action, resource, timestamp, success, details
FROM audit_logs
WHERE user_id = '123e4567-e89b-12d3-a456-426614174000'
ORDER BY timestamp DESC
LIMIT 100;

-- All events in a request trace
SELECT * FROM audit_logs
WHERE request_id = '550e8400-e29b-41d4-a716-446655440000'
ORDER BY timestamp;
```

## Best Practices

### DO

✅ Log every authentication attempt (success and failure)  
✅ Log all authorization changes (roles, permissions)  
✅ Include contextual details (who, what, when, where, why)  
✅ Use structured details (JSONB) for complex data  
✅ Capture request_id for request correlation  
✅ Log both to database and structured logs  
✅ Implement retention policies for compliance  

### DON'T

❌ Log passwords, tokens, or sensitive credentials  
❌ Log excessive PII (only what's necessary)  
❌ Block HTTP requests on audit log failures (use go routines)  
❌ Delete audit logs outside of retention policies  
❌ Modify existing audit log entries  

## Performance Considerations

### Async Logging

Audit log database inserts are **asynchronous** (via goroutines) to prevent blocking HTTP requests:

```go
go func() {
    if createErr := s.repo.Create(context.Background(), auditLog); createErr != nil {
        logger.Error(ctx).Err(createErr).Msg("Failed to persist audit log to database")
    }
}()
```

### Indexing

All frequently queried fields are indexed:
- Single column indexes: user_id, organization_id, action, resource, timestamp
- Composite indexes: (user_id, action, timestamp), (organization_id, action, timestamp)

### Partitioning (Future)

For high-volume environments, consider table partitioning by timestamp:

```sql
CREATE TABLE audit_logs_2025_11 PARTITION OF audit_logs
FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
```

## Troubleshooting

### Audit Logs Not Appearing

1. Check database connection
2. Verify migration 012 has run
3. Check application logs for "Failed to persist audit log"
4. Confirm user_id/org_id are valid UUIDs

### Performance Issues

1. Monitor index usage: `EXPLAIN ANALYZE SELECT ...`
2. Check table size: `SELECT pg_size_pretty(pg_total_relation_size('audit_logs'));`
3. Implement retention policy if table is too large
4. Consider read replicas for audit queries

### Missing Context Fields

1. Ensure `StructuredLoggingMiddleware()` is applied globally
2. Ensure `AddUserContextMiddleware()` is applied after `AuthRequired()`
3. Check context propagation in request chain

## Integration with SIEM/Log Aggregation

Audit logs in structured logger format are compatible with:

- **Elasticsearch/Kibana**: Forward via Filebeat/Fluentd
- **Splunk**: Forward via Universal Forwarder
- **DataDog**: Forward via DataDog Agent
- **CloudWatch**: Forward via AWS CloudWatch Agent
- **Grafana Loki**: Forward via Promtail

Example Filebeat config:

```yaml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/auth-service/*.log
  json.keys_under_root: true
  json.add_error_key: true
  fields:
    service: auth-service
    audit: true
  fields_under_root: true

processors:
  - drop_event:
      when:
        not:
          has_fields: ['audit_action']

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "audit-logs-%{+yyyy.MM.dd}"
```

## See Also

- [STRUCTURED_LOGGING.md](./STRUCTURED_LOGGING.md) - Structured logging overview
- [SECURITY.md](./SECURITY.md) - Security architecture
- [API_FLOWS.md](./API_FLOWS.md) - API authentication flows
