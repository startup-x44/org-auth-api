# Security Model & Analysis

## Overview

This authentication service implements a multi-layered security approach combining industry best practices for password storage, token management, session security, multi-tenant isolation, and attack prevention.

---

## 1. Password Security

### Hashing Algorithm
```go
// Argon2id configuration
Config:
  Memory:      64 MB (65536 KB)
  Time:        3 iterations
  Threads:     4 parallel threads
  Salt:        16 bytes (crypto/rand)
  Key Length:  32 bytes
```

**Why Argon2id?**
- ✅ Winner of Password Hashing Competition (2015)
- ✅ Resistant to GPU/ASIC attacks
- ✅ Memory-hard function (defeats rainbow tables)
- ✅ Side-channel attack resistant
- ✅ Configurable work factor

**Storage Format:**
```
$argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
```

### Password Requirements
```
Minimum:    8 characters
Maximum:    72 characters (bcrypt limitation for compatibility)
Validation: Custom rules enforced at application layer
```

**Implemented Validations:**
- ✅ Minimum length check
- ❌ Complexity requirements (uppercase, numbers, symbols) - NOT enforced
- ❌ Password history - NOT implemented
- ❌ Common password blacklist - NOT implemented

**Recommendations:**
- 🔴 Implement password complexity requirements
- 🔴 Add common password blacklist (e.g., top 10k compromised passwords)
- 🟡 Consider password history (prevent reuse of last 5 passwords)
- 🟡 Add password strength meter on frontend

---

## 2. JWT Token Security

### Access Token Configuration
```go
Algorithm:    HS256 (HMAC-SHA256)
Expiration:   1 hour
Secret:       32+ byte random key from environment
Issuer:       "auth-service"
Audience:     "api"
```

**Claims:**
```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "organization_id": "org-uuid",
  "role_id": "role-uuid",
  "role_name": "admin",
  "permissions": ["user:read", "user:write", ...],
  "iat": 1700000000,
  "exp": 1700003600,
  "iss": "auth-service",
  "aud": "api"
}
```

### Refresh Token Configuration
```
Algorithm:    Opaque token (not JWT)
Storage:      Argon2id hashed in database
Expiration:   30 days
Rotation:     On each use (optional)
Revocation:   On logout, password change
```

**Security Features:**
- ✅ Short-lived access tokens (1 hour)
- ✅ Refresh tokens hashed in database
- ✅ Refresh tokens scoped to organization
- ✅ Tokens revoked on logout
- ✅ All tokens revoked on password change
- ✅ Token linked to session (device tracking)

**Vulnerabilities & Mitigations:**

| Vulnerability | Current State | Mitigation |
|--------------|---------------|------------|
| **Token theft via XSS** | 🔴 Vulnerable | Tokens in localStorage/sessionStorage |
| **Token theft via man-in-the-middle** | 🟢 Protected | HTTPS enforced |
| **Token replay attacks** | 🟡 Partial | Short expiration, but no jti/nonce |
| **Refresh token reuse** | 🟢 Protected | Hashed storage, revocation on use |
| **Token leakage in logs** | 🟢 Protected | Tokens not logged |

**Recommendations:**
- 🔴 **CRITICAL**: Move tokens to `httpOnly` cookies (prevents XSS)
- 🔴 Implement `jti` (JWT ID) claim + blacklist for revocation
- 🟡 Add token binding (link to IP/User-Agent)
- 🟡 Implement anomaly detection (location/device changes)
- 🟢 Consider shorter access token TTL (15-30 min)

---

## 3. Session Security

### Session Management
```go
Storage:       PostgreSQL sessions table
Tracking:      IP address, User-Agent, device fingerprint
Expiration:    30 days (sliding window)
Cleanup:       Expired sessions deleted on cron
```

**Session Fields:**
```sql
id                UUID PRIMARY KEY
user_id           UUID NOT NULL
organization_id   UUID NOT NULL
refresh_token     TEXT (hashed)
ip_address        VARCHAR(45)
user_agent        TEXT
device_name       VARCHAR(255)
last_activity_at  TIMESTAMP
expires_at        TIMESTAMP
created_at        TIMESTAMP
```

**Features:**
- ✅ Multi-device support
- ✅ Session listing (user can see all active sessions)
- ✅ Remote logout (user can revoke any session)
- ✅ Device fingerprinting
- ✅ IP tracking
- ✅ Last activity tracking
- ✅ Session expiration

**Security Concerns:**
- 🟡 No session fixation prevention (regenerate ID after login)
- 🟡 No concurrent session limits
- 🟡 No anomaly detection (e.g., same user from different countries)

**Recommendations:**
- 🟡 Regenerate session ID after login (prevent fixation)
- 🟡 Limit concurrent sessions per user (e.g., max 5 devices)
- 🟡 Alert on suspicious activity (new device/location)
- 🟢 Implement "trusted devices" feature

---

## 4. CSRF Protection

### Implementation
```go
Middleware:    custom CSRF middleware
Token Storage: Redis (in-memory)
Token Format:  Secure random 32-byte hex string
Header:        X-CSRF-Token
Cookie:        csrf_token (httpOnly, sameSite: Strict)
```

**Protection Flow:**
```
1. GET /api/v1/csrf-token
   -> Server generates token
   -> Stores in Redis (key: user_id, value: token)
   -> Returns token in response
   -> Sets httpOnly cookie with same token

2. POST/PUT/DELETE request
   -> Client includes X-CSRF-Token header
   -> Server validates header matches cookie
   -> Server validates token exists in Redis
   -> Request proceeds
```

**Exempt Routes:**
- Public auth endpoints (login, register)
- GET requests (CSRF only affects state-changing operations)

**Security Level:** 🟢 Strong
- ✅ Double-submit cookie pattern
- ✅ Server-side validation
- ✅ SameSite=Strict cookie
- ✅ Secure flag on cookies

**Recommendations:**
- 🟢 Current implementation is solid
- 🟡 Consider origin/referer header validation (defense in depth)

---

## 5. Rate Limiting

### Configuration
```go
Store:          Redis (distributed rate limiting)
Window:         Sliding window algorithm
Key Format:     "ratelimit:{endpoint}:{identifier}"
```

### Limits

| Endpoint | Rate Limit | Window | Identifier |
|----------|-----------|--------|------------|
| `/auth/register` | 10 requests | 1 hour | IP address |
| `/auth/login` | 20 requests | 15 minutes | IP address |
| `/auth/forgot-password` | 3 requests | 1 hour | Email |
| `/auth/refresh` | 10 requests | 1 minute | User ID |
| API endpoints | 1000 requests | 1 hour | User ID |

**Response:**
```
HTTP 429 Too Many Requests
X-RateLimit-Limit: 20
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1700003600

{ "error": "Rate limit exceeded. Try again in 15 minutes." }
```

**Security Level:** 🟢 Good
- ✅ Prevents brute force attacks
- ✅ Prevents credential stuffing
- ✅ Distributed (works across multiple servers)
- ✅ Per-endpoint granularity

**Recommendations:**
- 🟡 Implement progressive delays (exponential backoff)
- 🟡 Add CAPTCHA after 3 failed login attempts
- 🟡 Add IP reputation scoring (block known malicious IPs)

---

## 6. Multi-Tenant Isolation

### Isolation Strategy
```
Model: Database-level isolation (shared database, logical separation)
```

**Isolation Layers:**

1. **Application Layer:**
   ```go
   // Every query includes organization_id filter
   db.Where("organization_id = ?", ctx.OrganizationID)
   ```

2. **Database Layer:**
   ```sql
   -- Row-Level Security (RLS) policies
   CREATE POLICY org_isolation ON users
     USING (organization_id = current_setting('app.current_org')::uuid);
   ```

3. **API Layer:**
   ```go
   // Middleware extracts org_id from JWT
   // Stores in request context
   // All handlers use context org_id
   ```

**Security Matrix:**

| Threat | Protection | Status |
|--------|-----------|--------|
| **Cross-org data access** | WHERE clause filtering | 🟢 Protected |
| **Privilege escalation** | RBAC permission checks | 🟢 Protected |
| **JWT tampering** | HMAC signature validation | 🟢 Protected |
| **Org ID injection** | Org ID from JWT (not request) | 🟢 Protected |
| **SQL injection** | GORM parameterized queries | 🟢 Protected |
| **Missing isolation** | Code review required | 🟡 Risk exists |

**Current State:**
- ✅ Organization ID in JWT (server-controlled)
- ✅ All queries filtered by organization_id
- ✅ RBAC enforced at organization level
- ❌ No database-level RLS policies implemented
- ❌ No automated tests for isolation

**Vulnerabilities:**
- 🔴 **Developer error risk**: Forgetting to add `organization_id` filter
- 🟡 **Global resources**: Some queries need system-level access (system roles)

**Recommendations:**
- 🔴 **CRITICAL**: Implement Row-Level Security (RLS) in PostgreSQL
- 🔴 Add automated tests for multi-tenant isolation
- 🟡 Add linter/static analysis to detect missing org_id filters
- 🟡 Create read-only database views with built-in filtering

---

## 7. Audit Logging

### Implementation
```go
Table:         audit_logs
Storage:       PostgreSQL
Async:         Background goroutine (non-blocking)
Retention:     90 days (configurable)
```

**Audit Schema:**
```sql
id                UUID PRIMARY KEY
user_id           UUID
organization_id   UUID
action            VARCHAR(100)
resource_type     VARCHAR(50)
resource_id       UUID
ip_address        VARCHAR(45)
user_agent        TEXT
request_id        UUID
metadata          JSONB
created_at        TIMESTAMP
```

**Logged Events:**

| Category | Events |
|----------|--------|
| **Authentication** | login, logout, login_failed, password_changed, email_verified |
| **Authorization** | role_assigned, permission_granted, permission_denied |
| **User Management** | user_created, user_updated, user_deleted, user_invited |
| **Organization** | org_created, org_updated, member_added, member_removed |
| **RBAC** | role_created, role_updated, role_deleted, permission_assigned |
| **OAuth2** | oauth_authorization, oauth_token_issued, oauth_token_revoked |
| **Session** | session_created, session_revoked |

**Metadata Examples:**
```json
// Login attempt
{
  "email": "user@example.com",
  "success": false,
  "reason": "invalid_password",
  "device": "Chrome on macOS"
}

// Permission denied
{
  "required_permission": "user:delete",
  "user_permissions": ["user:read", "user:write"],
  "resource_type": "user",
  "resource_id": "uuid"
}
```

**Security Level:** 🟢 Good
- ✅ Comprehensive event coverage
- ✅ Tamper-resistant (append-only)
- ✅ Includes context (IP, User-Agent, request ID)
- ✅ JSONB metadata for flexibility

**Recommendations:**
- 🟡 Add log export API for SIEM integration
- 🟡 Implement log integrity verification (hash chain)
- 🟡 Add real-time alerting for suspicious patterns
- 🟢 Consider dedicated audit database (read-only replica)

---

## 8. Security Headers

### HTTP Headers
```go
// Set by middleware
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

**CORS Configuration:**
```go
AllowOrigins:     [configurable whitelist]
AllowMethods:     GET, POST, PUT, DELETE, PATCH
AllowHeaders:     Authorization, Content-Type, X-CSRF-Token
AllowCredentials: true
MaxAge:           12 hours
```

**Security Level:** 🟢 Good
- ✅ All major security headers implemented
- ✅ CORS configured (not wide open)
- ✅ HTTPS enforced

**Recommendations:**
- 🟡 Tighten CSP policy (current is permissive)
- 🟢 Add Subresource Integrity (SRI) for CDN resources

---

## 9. Input Validation

### Validation Layers

1. **Schema Validation (go-playground/validator)**
   ```go
   type RegisterRequest struct {
       Email    string `json:"email" binding:"required,email"`
       Password string `json:"password" binding:"required,min=8"`
   }
   ```

2. **Custom Validation**
   ```go
   // Email format validation
   // Password complexity
   // UUID format validation
   // Enum validation (role names, permission names)
   ```

3. **Database Constraints**
   ```sql
   UNIQUE (email)
   NOT NULL constraints
   FOREIGN KEY constraints
   CHECK constraints
   ```

**Protected Against:**
- ✅ SQL Injection (GORM parameterized queries)
- ✅ XSS (React auto-escaping, CSP headers)
- ✅ Command Injection (no exec/system calls)
- ✅ Path Traversal (no file operations from user input)
- ✅ Email validation (RFC 5322 format)
- ✅ UUID validation (prevents ID enumeration)

**Recommendations:**
- 🟡 Add request size limits (prevent DoS)
- 🟡 Implement input sanitization library
- 🟢 Add JSON schema validation for complex objects

---

## 10. Dependency Security

### Go Dependencies (go.mod)
```
Total:     40+ dependencies
Direct:    25 dependencies
Indirect:  15+ dependencies
```

**Key Security-Sensitive Dependencies:**
- `github.com/gin-gonic/gin` v1.9.1 (web framework)
- `github.com/golang-jwt/jwt` v4.5.2 (JWT library)
- `gorm.io/gorm` v1.25.2 (ORM)
- `golang.org/x/crypto` v0.31.0 (Argon2, bcrypt)

**Vulnerability Scanning:**
- ❌ No automated dependency scanning
- ❌ No CI/CD security checks
- ❌ No SBOM (Software Bill of Materials)

**Recommendations:**
- 🔴 **CRITICAL**: Add `govulncheck` to CI/CD pipeline
- 🟡 Implement Dependabot/Renovate for automated updates
- 🟡 Generate SBOM for compliance
- 🟢 Add SCA (Software Composition Analysis) tool

---

## 11. Attack Surface Analysis

### Attack Vectors

| Vector | Risk | Current Protection | Recommendation |
|--------|------|-------------------|----------------|
| **Brute Force Login** | 🟡 Medium | Rate limiting | Add CAPTCHA, progressive delays |
| **Credential Stuffing** | 🟡 Medium | Rate limiting, Argon2 | Add device fingerprinting, breach detection |
| **Session Hijacking** | 🟡 Medium | HTTPS, short TTL | Move tokens to httpOnly cookies |
| **CSRF** | 🟢 Low | CSRF middleware | Current protection sufficient |
| **XSS** | 🟡 Medium | React escaping, CSP | Tighten CSP, add sanitization |
| **SQL Injection** | 🟢 Low | GORM parameterized | Current protection sufficient |
| **JWT Tampering** | 🟢 Low | HMAC signature | Consider RS256 for distributed systems |
| **Multi-Tenant Leakage** | 🔴 High | App-level filtering | **CRITICAL**: Add RLS policies |
| **Privilege Escalation** | 🟡 Medium | RBAC checks | Add automated tests |
| **DoS** | 🟡 Medium | Rate limiting | Add request size limits, connection limits |
| **Dependency Vulnerabilities** | 🟡 Medium | None | **CRITICAL**: Add vulnerability scanning |

---

## 12. Compliance Considerations

### GDPR Compliance
- ✅ User data deletion (soft delete implemented)
- ✅ Audit logging (data access tracking)
- ✅ Password hashing (data protection at rest)
- ❌ Data export functionality (user data portability)
- ❌ Consent management
- ❌ Data retention policies

### OWASP Top 10 (2021)

| Risk | Status | Notes |
|------|--------|-------|
| **A01: Broken Access Control** | 🟡 Partial | RBAC implemented, needs RLS |
| **A02: Cryptographic Failures** | 🟢 Protected | Argon2, HTTPS, hashed tokens |
| **A03: Injection** | 🟢 Protected | GORM parameterized queries |
| **A04: Insecure Design** | 🟡 Partial | Missing threat modeling |
| **A05: Security Misconfiguration** | 🟡 Partial | Headers good, CSP permissive |
| **A06: Vulnerable Components** | 🔴 At Risk | No dependency scanning |
| **A07: Identification/Auth Failures** | 🟢 Protected | Strong auth, MFA missing |
| **A08: Software/Data Integrity** | 🟡 Partial | No code signing, SRI missing |
| **A09: Logging Failures** | 🟢 Protected | Comprehensive audit logs |
| **A10: SSRF** | 🟢 Protected | No outbound requests from user input |

---

## 13. Security Roadmap

### Critical (Fix Immediately)
- 🔴 Implement PostgreSQL Row-Level Security (RLS)
- 🔴 Add dependency vulnerability scanning
- 🔴 Move tokens to httpOnly cookies
- 🔴 Add automated multi-tenant isolation tests

### High Priority (Next Sprint)
- 🟡 Implement password complexity requirements
- 🟡 Add CAPTCHA to login/register
- 🟡 Add common password blacklist
- 🟡 Implement JWT blacklist (jti claims)

### Medium Priority (Next Quarter)
- 🟡 Add MFA/2FA support
- 🟡 Implement anomaly detection
- 🟡 Add SIEM integration
- 🟡 Implement data export API (GDPR)

### Nice to Have (Future)
- 🟢 Add passwordless authentication
- 🟢 Implement trusted devices
- 🟢 Add biometric authentication support
- 🟢 Implement zero-trust architecture

---

**Last Updated**: November 18, 2025  
**Security Contact**: security@example.com  
**Vulnerability Reporting**: security@example.com (PGP key available)
