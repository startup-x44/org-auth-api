# Auth Service - Project Overview

## 1. Project Type & Purpose

### What This Is
**Enterprise Multi-Tenant Authentication & Authorization Microservice** with OAuth2 support, designed for SaaS platforms requiring sophisticated access control and organization management.

### Problem It Solves
- **Multi-tenant authentication** - Single auth service managing multiple isolated organizations
- **Fine-grained RBAC** - Custom roles and permissions per organization
- **OAuth2 SSO** - Third-party application integration via OAuth2/OIDC
- **Secure session management** - Device tracking, geographic location, session revocation
- **Enterprise-grade security** - CSRF protection, rate limiting, audit logging

### Domain
**Enterprise SaaS Infrastructure** - Applicable to:
- Educational platforms (credentialing systems, LMS)
- B2B SaaS platforms
- Healthcare portals
- Financial services platforms
- Any multi-tenant enterprise application

### Expected Workflows

#### User Journey
1. **Global Registration** → User creates account (email verification)
2. **Organization Selection** → User selects/creates organization workspace
3. **Role Assignment** → Organization admin assigns roles with permissions
4. **Access Control** → API validates permissions per request
5. **Multi-org Access** → User can switch between organizations

#### Admin Journey
1. **Superadmin Access** → Platform-level administration
2. **System Roles** → Manages global roles (Owner, Admin, Member)
3. **User Management** → Activate/deactivate users globally
4. **Organization Oversight** → Monitor all organizations
5. **OAuth2 Apps** → Register third-party client applications

#### Developer Journey
1. **OAuth2 Registration** → Register client application
2. **Authorization Flow** → Implement PKCE-secured OAuth2
3. **API Integration** → Use access tokens for API calls
4. **API Key Management** → Create API keys for server-to-server

---

## 2. Architecture Type

### Pattern: **Clean Architecture + Multi-Tenant Organization Isolation**

```
┌─────────────────────────────────────────────────────┐
│                   API Layer (Gin)                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │ Handlers │──│Middleware│──│  Routes  │          │
│  └──────────┘  └──────────┘  └──────────┘          │
└─────────────────────────────────────────────────────┘
                         │
┌─────────────────────────────────────────────────────┐
│               Service Layer (Business Logic)         │
│  ┌────────────┐ ┌───────────┐ ┌─────────────┐      │
│  │ Auth       │ │  Role     │ │Organization │      │
│  │ Service    │ │  Service  │ │  Service    │      │
│  └────────────┘ └───────────┘ └─────────────┘      │
│  ┌────────────┐ ┌───────────┐ ┌─────────────┐      │
│  │  OAuth2    │ │ Session   │ │   User      │      │
│  │  Service   │ │ Service   │ │  Service    │      │
│  └────────────┘ └───────────┘ └─────────────┘      │
└─────────────────────────────────────────────────────┘
                         │
┌─────────────────────────────────────────────────────┐
│            Repository Layer (Data Access)            │
│  ┌────────────┐ ┌───────────┐ ┌─────────────┐      │
│  │   User     │ │   Role    │ │Organization │      │
│  │   Repo     │ │   Repo    │ │    Repo     │      │
│  └────────────┘ └───────────┘ └─────────────┘      │
└─────────────────────────────────────────────────────┘
                         │
┌─────────────────────────────────────────────────────┐
│              Infrastructure Layer                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │PostgreSQL│  │  Redis   │  │   SMTP   │          │
│  └──────────┘  └──────────┘  └──────────┘          │
└─────────────────────────────────────────────────────┘
```

### Key Architectural Decisions

1. **Repository Pattern** - Abstraction over database operations
2. **Service Layer** - Business logic isolation from HTTP handlers
3. **Organization Isolation** - Every resource scoped to organization
4. **System vs Custom Resources** - Global system resources + org-specific custom resources
5. **JWT-based Authentication** - Stateless tokens with Redis session tracking
6. **RBAC Model** - Roles → Permissions → Resources

---

## 3. Multi-Tenancy Model

### Slack-Style Organization Isolation

```
User (Global)
  ├── Organization A
  │     ├── Role: Owner (system role)
  │     ├── Permissions: Full access
  │     └── Sessions: Org-scoped sessions
  │
  ├── Organization B
  │     ├── Role: Member (system role)
  │     ├── Permissions: Limited access
  │     └── Sessions: Separate org sessions
  │
  └── Superadmin (platform level)
        └── Access: All organizations
```

### Isolation Strategy

1. **Data Isolation** - All resources have `organization_id`
2. **Session Isolation** - Sessions scoped per organization
3. **Token Isolation** - JWT tokens contain organization context
4. **Permission Isolation** - Permissions checked per organization
5. **Role Isolation** - Custom roles per organization + global system roles

### System vs Custom Resources

| Resource Type | System | Custom |
|--------------|--------|---------|
| **Roles** | Owner, Admin, Member (global) | Org-specific roles |
| **Permissions** | 22 predefined permissions | Org-specific permissions |
| **Scope** | `organization_id = NULL` | `organization_id = <uuid>` |
| **Management** | Superadmin only | Org admin |
| **Isolation** | Shared across all orgs | Isolated per org |

---

## 4. Security Model

### Defense Layers

1. **Authentication** - JWT tokens + session tracking
2. **Authorization** - RBAC with permission checking
3. **CSRF Protection** - Token-based CSRF middleware
4. **Rate Limiting** - IP-based and user-based limits
5. **Session Security** - Device fingerprinting, geo-location
6. **Audit Logging** - All security events logged
7. **Password Security** - Argon2 hashing
8. **Email Verification** - Required for account activation

### Token Security

```
Access Token (1 hour)
  ├── Payload: user_id, organization_id, permissions
  ├── Signature: HS256 with secret
  └── Storage: httpOnly cookie (frontend), localStorage (dev)

Refresh Token (30 days)
  ├── Hash: Argon2 in database
  ├── Scope: Organization-specific
  ├── Revocation: On logout, password change
  └── Session Link: Tied to user session
```

### RBAC Permission Check Flow

```
HTTP Request
    ↓
Middleware: Extract JWT → Validate
    ↓
Get User Permissions for Organization
    ↓
Check: HasPermission(user, org, permission)
    ↓
    ├─ YES → Allow Request
    └─ NO  → 403 Forbidden
```

---

## 5. Technology Decisions & Rationale

### Why Go + Gin?
- **Performance** - Compiled, concurrent, low latency
- **Scalability** - Goroutines handle thousands of connections
- **Security** - Strong typing prevents many vulnerabilities
- **Deployment** - Single binary, minimal dependencies

### Why PostgreSQL?
- **ACID Compliance** - Critical for auth/finance data
- **JSONB** - Flexible for organization settings
- **Foreign Keys** - Data integrity enforcement
- **UUID Support** - Native UUID for primary keys

### Why Redis?
- **Session Store** - Fast in-memory session lookups
- **Rate Limiting** - Efficient counters and TTLs
- **Token Blacklist** - Quick revocation checks
- **Cache Layer** - Permission caching (future)

### Why React + TypeScript?
- **Type Safety** - Catch errors at compile time
- **Developer Experience** - Better IDE support
- **Maintainability** - Self-documenting code
- **Ecosystem** - Rich UI library (Radix UI)

---

## 6. Current State Assessment

### Maturity Level: **Production-Ready with Areas for Improvement**

**Strengths ✅**
- Solid authentication foundation
- Comprehensive RBAC implementation
- OAuth2/OIDC support
- Multi-tenant architecture
- Audit logging framework
- Docker containerization
- Kubernetes manifests

**Production Gaps ⚠️**
- Missing database migration version control
- No load testing benchmarks
- Incomplete error monitoring integration
- Limited observability (no tracing)
- Missing backup/restore procedures
- No disaster recovery plan

---

## 7. Recommended Next Steps

### Immediate (1-2 weeks)
1. Add database migration versioning (Flyway/Liquibase style)
2. Implement structured logging (JSON format)
3. Add health check endpoints with dependency checks
4. Create runbook for common operations

### Short-term (1 month)
1. Implement distributed tracing (OpenTelemetry)
2. Add Prometheus metrics
3. Create load testing suite (k6/Gatling)
4. Implement circuit breakers for external services

### Long-term (3 months)
1. Multi-region deployment strategy
2. Read replica support
3. Event sourcing for audit trail
4. GraphQL API layer
5. WebSocket support for real-time features

---

**Last Updated**: November 18, 2025  
**Version**: 1.0.0  
**Status**: Production-Ready with Active Development
