# Multi-Tenant Authentication Service

A comprehensive SaaS authentication system built with Go and React, featuring domain-based multi-tenancy, JWT authentication, and role-based access control.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Technology Stack](#technology-stack)
- [Database Schema](#database-schema)
- [API Documentation](#api-documentation)
- [Authentication Flow](#authentication-flow)
- [Multi-Tenant Implementation](#multi-tenant-implementation)
- [Development Setup](#development-setup)
- [Deployment](#deployment)

## Architecture Overview

### System Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React SPA     │    │   Gin API       │    │   PostgreSQL    │
│   (Frontend)    │◄──►│   (Backend)     │◄──►│   (Database)    │
│                 │    │                 │    │                 │
│ • JWT Auth      │    │ • JWT Service   │    │ • Users         │
│ • Tenant-aware  │    │ • Multi-tenant  │    │ • Tenants       │
│ • localStorage  │    │ • Role-based    │    │ • Sessions      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │     Redis       │
                       │  (Sessions)     │
                       └─────────────────┘
```

### Key Features

- **Multi-Tenant Architecture**: Domain-based tenant isolation
- **JWT Authentication**: RSA-signed tokens with tenant claims
- **Session Management**: Redis-backed sessions with device tracking
- **Role-Based Access**: Global user types with admin privileges
- **Security Features**: CSRF protection, rate limiting, account lockout
- **Auto Tenant Creation**: Organizations created automatically from email domains

## Technology Stack

### Backend
- **Language**: Go 1.23
- **Framework**: Gin (HTTP router)
- **ORM**: GORM v2
- **Database**: PostgreSQL
- **Cache**: Redis
- **Authentication**: Custom JWT with RSA signing
- **Password Hashing**: Argon2id

### Frontend
- **Framework**: React 18.2.0
- **Build Tool**: Create React App
- **Styling**: Tailwind CSS 3.3.2
- **Routing**: React Router DOM v6
- **HTTP Client**: Axios with interceptors
- **State Management**: React Context

### DevOps
- **Containerization**: Docker & Docker Compose
- **Development**: Hot reload with CompileDaemon
- **Testing**: Isolated test databases
- **CI/CD**: GitHub Actions (planned)

## Database Schema

### Core Tables

#### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    email VARCHAR NOT NULL,
    email_verified_at TIMESTAMP,
    password_hash VARCHAR NOT NULL,
    firstname VARCHAR(100),
    lastname VARCHAR(100),
    address TEXT,
    phone VARCHAR(20),
    user_type VARCHAR NOT NULL, -- Admin, Student, RTO, Issuer, Validator, badger, Non-partner, Partner
    status VARCHAR DEFAULT 'active', -- active, suspended, deactivated
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

-- Indexes
CREATE INDEX idx_users_type ON users(user_type);
CREATE INDEX idx_users_status ON users(status);
CREATE UNIQUE INDEX idx_users_email_tenant ON users(email, tenant_id);
```

#### Tenants Table
```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR NOT NULL,
    domain VARCHAR UNIQUE NOT NULL,
    status VARCHAR DEFAULT 'active', -- active, suspended
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

#### Sessions Table
```sql
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    token_hash VARCHAR UNIQUE NOT NULL,
    ip_address INET,
    user_agent TEXT,
    device_fingerprint TEXT,
    location TEXT,
    is_active BOOLEAN DEFAULT true,
    last_activity TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    revoked_reason TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

-- Indexes
CREATE INDEX idx_sessions_user ON user_sessions(user_id);
CREATE INDEX idx_sessions_tenant ON user_sessions(tenant_id);
CREATE UNIQUE INDEX idx_sessions_token ON user_sessions(token_hash);
CREATE INDEX idx_sessions_expires ON user_sessions(expires_at);
CREATE INDEX idx_sessions_activity ON user_sessions(last_activity);
```

#### Additional Tables

**Refresh Tokens**
```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    token_hash VARCHAR UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
```

**Password Resets**
```sql
CREATE TABLE password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    token_hash VARCHAR UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
```

**Failed Login Attempts**
```sql
CREATE TABLE failed_login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    tenant_id UUID NOT NULL,
    email VARCHAR NOT NULL,
    ip_address INET NOT NULL,
    user_agent TEXT,
    attempted_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

-- Indexes
CREATE INDEX idx_failed_attempts_user ON failed_login_attempts(user_id);
CREATE INDEX idx_failed_attempts_email_ip ON failed_login_attempts(email, ip_address, tenant_id);
```

## API Documentation

### Authentication Endpoints

#### POST /api/v1/auth/register
Register a new user account.

**Request Body:**
```json
{
  "email": "user@company.com",
  "password": "SecurePass123!",
  "confirm_password": "SecurePass123!",
  "user_type": "Student",
  "tenant_id": "company.com",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@company.com",
      "user_type": "Student",
      "tenant_id": "uuid",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true
    },
    "token": {
      "access_token": "jwt_token",
      "refresh_token": "refresh_token",
      "expires_in": 3600,
      "token_type": "Bearer"
    }
  }
}
```

#### POST /api/v1/auth/login
Authenticate a user.

**Request Body:**
```json
{
  "email": "user@company.com",
  "password": "SecurePass123!",
  "tenant_id": "company.com"
}
```

#### POST /api/v1/auth/refresh
Refresh access token.

**Request Body:**
```json
{
  "refresh_token": "refresh_token_here"
}
```

### User Management Endpoints

#### GET /api/v1/user/profile
Get current user profile.

**Headers:**
```
Authorization: Bearer <access_token>
X-Tenant-ID: <tenant_id>
```

#### PUT /api/v1/user/profile
Update user profile.

**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Smith",
  "phone": "+1987654321"
}
```

#### POST /api/v1/user/change-password
Change user password.

**Request Body:**
```json
{
  "current_password": "OldPass123!",
  "new_password": "NewPass123!",
  "confirm_password": "NewPass123!"
}
```

### Admin Endpoints

#### GET /api/v1/admin/users
List users (paginated).

**Query Parameters:**
- `limit`: Number of users to return (default: 10, max: 100)
- `cursor`: Cursor for pagination

#### PUT /api/v1/admin/users/{userId}/activate
Activate a user account.

#### PUT /api/v1/admin/users/{userId}/deactivate
Deactivate a user account.

#### DELETE /api/v1/admin/users/{userId}
Delete a user account.

#### POST /api/v1/admin/tenants
Create a new tenant.

**Request Body:**
```json
{
  "name": "New Company",
  "domain": "newcompany.com"
}
```

#### GET /api/v1/admin/tenants
List tenants (paginated).

#### GET /api/v1/admin/tenants/{tenantId}
Get tenant details.

#### PUT /api/v1/admin/tenants/{tenantId}
Update tenant.

#### DELETE /api/v1/admin/tenants/{tenantId}
Delete tenant.

## Authentication Flow

### Registration Flow

1. **Client Request**: User submits registration form with email
2. **Tenant Resolution**: Frontend extracts domain from email (e.g., `user@company.com` → `company.com`)
3. **Backend Validation**: Server validates email format and password strength
4. **Tenant Lookup**: Check if tenant exists for domain
5. **Auto Tenant Creation**: If tenant doesn't exist, create it automatically
6. **User Creation**: Create user account linked to tenant
7. **Password Hashing**: Hash password using Argon2id
8. **Token Generation**: Generate JWT tokens with tenant claims
9. **Session Creation**: Create Redis-backed session
10. **Response**: Return user data and tokens

### Login Flow

1. **Client Request**: User submits email, password, and tenant_id
2. **Tenant Validation**: Verify tenant exists and is active
3. **Account Lockout Check**: Check for too many failed attempts
4. **User Lookup**: Find user by email and tenant_id
5. **Password Verification**: Compare Argon2id hash
6. **Session Management**: Create/update Redis session
7. **Token Generation**: Generate new JWT tokens
8. **Activity Logging**: Update last login timestamp
9. **Response**: Return user data and tokens

### Token Refresh Flow

1. **Client Request**: Send refresh token
2. **Token Validation**: Verify refresh token exists and not expired
3. **User Verification**: Ensure user still exists and is active
4. **New Token Generation**: Create fresh access token
5. **Token Update**: Update refresh token expiration
6. **Response**: Return new token pair

## Multi-Tenant Implementation

### Tenant Resolution Strategy

The system uses a hierarchical tenant resolution approach:

1. **Explicit Tenant ID**: Direct UUID or domain in request
2. **Subdomain Detection**: `subdomain.sprout.com` → `subdomain`
3. **Email Domain Extraction**: `user@company.com` → `company.com`
4. **Auto Tenant Creation**: New tenants created automatically from domains

### Tenant Isolation

- **Database Level**: All queries include `WHERE tenant_id = ?`
- **JWT Claims**: Tokens contain `tenant_id` for stateless verification
- **Middleware Enforcement**: `TenantRequired` middleware injects tenant context
- **Storage Isolation**: Frontend uses tenant-specific localStorage keys

### Security Boundaries

- **Data Isolation**: Users can only access data from their tenant
- **Session Scoping**: Sessions are tenant-specific
- **Token Validation**: JWT claims verified for tenant ownership
- **Admin Restrictions**: Admins can only manage users in their tenant

## Development Setup

### Prerequisites

- Go 1.23+
- Node.js 18+
- PostgreSQL 13+
- Redis 6+
- Docker & Docker Compose

### Local Development

1. **Clone the repository**
```bash
git clone <repository-url>
cd auth-service
```

2. **Environment Configuration**
```bash
cp .env.example .env
# Edit .env with your local configuration
```

3. **Start Development Environment**
```bash
# Start all services with hot reload
./dev.sh dev

# Or start in background
./dev.sh dev-d
```

4. **Database Setup**
The application automatically:
- Creates PostgreSQL database
- Runs GORM migrations
- Seeds initial data (default tenants)

5. **Frontend Development**
```bash
cd frontend
yarn install
yarn start
```

### Testing

```bash
# Run all tests with isolated databases
./dev.sh test

# Run specific test types
go test ./tests/unit/... -v     # Unit tests
go test ./tests/feature/... -v  # Integration tests
```

### Development Commands

```bash
./dev.sh dev      # Start with live reload
./dev.sh dev-d    # Start in background
./dev.sh stop     # Stop containers
./dev.sh logs     # View logs
./dev.sh shell    # Open container shell
./dev.sh clean    # Remove containers and volumes
```

## Deployment

### Production Configuration

1. **Environment Variables**
```env
ENVIRONMENT=production
GIN_MODE=release
DATABASE_URL=postgresql://user:pass@host:5432/db
REDIS_URL=redis://host:6379
JWT_SECRET=your-secret-key
```

2. **Security Considerations**
- Use strong, unique JWT secrets
- Configure CORS for production domains
- Enable CSRF protection
- Set up proper SSL/TLS
- Configure rate limiting
- Enable audit logging

3. **Database Migration**
```bash
# Migrations run automatically on startup
# For manual migration in production:
go run cmd/server/main.go --migrate-only
```

4. **Health Checks**
```bash
# Health endpoint
GET /health

# Response
{
  "success": true,
  "data": {
    "status": "healthy",
    "database": "connected",
    "redis": "connected",
    "uptime": "1h 30m"
  }
}
```

### Docker Deployment

```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  auth-service:
    image: auth-service:latest
    environment:
      - ENVIRONMENT=production
      - GIN_MODE=release
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:13
    environment:
      - POSTGRES_DB=auth_db
      - POSTGRES_USER=auth_user
      - POSTGRES_PASSWORD=secure_password

  redis:
    image: redis:6-alpine
```

### Monitoring & Observability

- **Health Checks**: `/health` endpoint for load balancer
- **Logging**: Structured logging with request IDs
- **Metrics**: Database connection pools, request latency
- **Alerts**: Failed login attempts, account lockouts

## Security Features

### Authentication Security
- **Argon2id Password Hashing**: Memory-hard function resistant to attacks
- **RSA-signed JWT Tokens**: Asymmetric signing for enhanced security
- **Token Expiration**: Short-lived access tokens (1 hour) with refresh tokens (7 days)
- **Secure Token Storage**: HTTP-only cookies recommended for production

### Multi-Tenant Security
- **Tenant Isolation**: Database-level row security
- **Context Injection**: Tenant ID injected into every request context
- **Ownership Validation**: All operations verify tenant ownership
- **Session Scoping**: Sessions isolated per tenant

### Protection Mechanisms
- **CSRF Protection**: Double-submit cookie pattern
- **Rate Limiting**: Request throttling by IP and endpoint
- **Account Lockout**: Progressive delays after failed attempts
- **Audit Logging**: Security events and admin actions logged
- **Input Validation**: Comprehensive validation of all user inputs

### Compliance Considerations
- **Data Encryption**: Passwords hashed, sensitive data encrypted at rest
- **Access Logging**: All authentication attempts logged
- **Session Management**: Automatic cleanup of expired sessions
- **GDPR Compliance**: User data export/deletion capabilities

## Future Enhancements

### Planned Features
- **OAuth Integration**: Social login providers (Google, GitHub)
- **MFA Support**: TOTP and SMS-based two-factor authentication
- **Advanced Permissions**: Granular role-based access control
- **Audit Trails**: Comprehensive user activity logging
- **API Rate Limiting**: Per-user and per-endpoint limits
- **Email Templates**: Customizable email notifications

### Architecture Improvements
- **Microservices Split**: Separate auth service from user management
- **Event Sourcing**: Event-driven architecture for audit trails
- **CQRS Pattern**: Separate read/write models for performance
- **API Gateway**: Centralized request routing and authentication
- **Service Mesh**: Istio for inter-service communication

---

This documentation provides a comprehensive overview of the Multi-Tenant Authentication Service. For specific implementation details, refer to the source code and inline documentation.