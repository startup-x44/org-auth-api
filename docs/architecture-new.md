# Architecture Guide - Multi-Tenant Authentication Service

## System Overview

The Multi-Tenant Authentication Service is a production-ready SaaS authentication microservice built with Go and React, designed to provide secure, scalable user authentication and authorization with complete tenant isolation.

## Architecture Principles

### 1. Multi-Tenant Isolation
- **Database Level**: Every query includes tenant filtering
- **Application Level**: Tenant context injected into every request
- **Security Level**: Users can only access their tenant's data

### 2. Clean Architecture
- **Dependency Inversion**: Interfaces define contracts
- **Single Responsibility**: Each layer has one purpose
- **Dependency Injection**: Services injected via interfaces

### 3. Security First
- **Defense in Depth**: Multiple security layers
- **Zero Trust**: Every request validated
- **Audit Trail**: All actions logged

## System Components

### Backend Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Layer (Gin)                         │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Middleware Stack                           │ │
│  │  • CORS • Rate Limiting • Logging • Recovery • CSRF     │ │
│  │  • AuthRequired • TenantRequired • AdminRequired       │ │
│  └─────────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Handler Layer                              │ │
│  │  • AuthHandler • AdminHandler                           │ │
│  └─────────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Service Layer                              │ │
│  │  • UserService • TenantService • AuthService           │ │
│  └─────────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Repository Layer                           │ │
│  │  • UserRepository • TenantRepository • SessionRepository│ │
│  └─────────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Data Layer                                 │ │
│  │  • PostgreSQL • Redis • GORM                            │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Frontend Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    React Application                        │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Component Layer                            │ │
│  │  • Auth Components • Admin Components • Layout          │ │
│  └─────────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Context Layer                              │ │
│  │  • AuthContext • Tenant Resolution                      │ │
│  └─────────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Service Layer                              │ │
│  │  • API Client • Axios Interceptors • Token Management  │ │
│  └─────────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Utility Layer                              │ │
│  │  • Tenant Utils • Validation • Storage Helpers         │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Data Flow Architecture

### Authentication Flow

```
1. User Request
       ↓
2. Frontend Resolution
   • Extract tenant from email
   • Validate form data
       ↓
3. API Request
   • Axios interceptor adds tenant header
   • JWT token attached if available
       ↓
4. Middleware Stack
   • CORS validation
   • Rate limiting
   • CSRF protection
   • JWT validation
   • Tenant context injection
       ↓
5. Handler Layer
   • Route to appropriate handler
   • Bind request data
       ↓
6. Service Layer
   • Business logic validation
   • Tenant ownership checks
   • Password hashing/verification
       ↓
7. Repository Layer
   • Database queries with tenant filtering
   • Transaction management
       ↓
8. Database Layer
   • PostgreSQL with GORM
   • Redis for sessions
       ↓
9. Response
   • Structured JSON response
   • JWT tokens for authentication
```

### Multi-Tenant Data Isolation

```
Request → Middleware → Service → Repository → Database
    ↓         ↓         ↓         ↓          ↓
Tenant   Tenant    Tenant    Tenant     Row
Context  Required  Validation Filtering Security
```

## Component Details

### Backend Components

#### HTTP Layer (Gin Router)
- **Purpose**: HTTP request routing and response handling
- **Features**:
  - RESTful API endpoints
  - JSON request/response handling
  - Error response formatting
  - Request logging

#### Middleware Stack
- **CORS Middleware**: Configurable cross-origin requests
- **Rate Limiting**: Request throttling by IP/endpoint
- **Logging**: Structured request logging
- **Recovery**: Panic recovery with error responses
- **CSRF Protection**: Double-submit cookie validation
- **AuthRequired**: JWT token validation
- **TenantRequired**: Tenant context injection
- **AdminRequired**: Admin role validation

#### Handler Layer
- **AuthHandler**: Authentication endpoints (login, register, refresh)
- **AdminHandler**: Administrative functions (user/tenant management)
- **Responsibilities**:
  - Request binding and validation
  - Response formatting
  - Error handling
  - Service method invocation

#### Service Layer
- **UserService**: User management business logic
- **TenantService**: Tenant management operations
- **AuthService**: Authentication coordination
- **Features**:
  - Business rule validation
  - Password hashing/verification
  - Token generation
  - Session management
  - Audit logging

#### Repository Layer
- **UserRepository**: User data operations
- **TenantRepository**: Tenant data operations
- **SessionRepository**: Session management
- **Features**:
  - Database abstraction
  - Query building with tenant filtering
  - Transaction management
  - Connection pooling

#### Data Layer
- **PostgreSQL**: Primary data storage
- **Redis**: Session storage and caching
- **GORM**: ORM with migration support

### Frontend Components

#### Component Layer
- **Auth Components**: Login, Register, Password Reset
- **Admin Components**: User management, Tenant management
- **Layout Components**: Navigation, Header, Footer
- **Features**: React hooks, state management, form handling

#### Context Layer
- **AuthContext**: Global authentication state
- **Features**:
  - User session management
  - Tenant-specific localStorage
  - Automatic token refresh
  - Login/logout handlers

#### Service Layer
- **API Client**: Axios-based HTTP client
- **Features**:
  - Request/response interceptors
  - Automatic token attachment
  - CSRF token handling
  - Error handling and retry logic

#### Utility Layer
- **Tenant Utils**: Domain extraction and resolution
- **Validation**: Client-side form validation
- **Storage**: Tenant-aware localStorage helpers

## Security Architecture

### Authentication Security
```
Password → Argon2id Hashing → Database Storage
                    ↓
Login Request → Hash Comparison → JWT Generation
                    ↓
RSA Signing → Token Distribution → Stateless Verification
```

### Multi-Tenant Security
```
Tenant Resolution → Context Injection → Query Filtering → Data Isolation
     ↓                     ↓                ↓             ↓
Email Domain      Middleware       Repository      Row Security
Subdomain         Injection        WHERE Clauses   PostgreSQL
Header Override   Validation       Tenant ID       Policies
```

### Session Security
```
Login → Session Creation → Redis Storage → Token Validation
   ↓           ↓               ↓            ↓
Device     Fingerprinting    TTL          Expiration
Tracking   IP Address       Rotation      Checks
```

## Database Design

### Schema Relationships

```
┌─────────────┐     ┌─────────────┐
│   Tenants   │────│    Users    │
└─────────────┘     └─────────────┘
       │                   │
       │                   │
       ▼                   ▼
┌─────────────┐     ┌─────────────┐
│User Sessions│     │Refresh      │
│             │     │Tokens       │
└─────────────┘     └─────────────┘
       │                   │
       │                   │
       ▼                   ▼
┌─────────────┐     ┌─────────────┐
│Failed Login │     │Password     │
│Attempts     │     │Resets       │
└─────────────┘     └─────────────┘
```

### Indexing Strategy

#### Primary Indexes
- `users(id)` - UUID primary key
- `tenants(id)` - UUID primary key
- `user_sessions(token_hash)` - Unique token index
- `refresh_tokens(token_hash)` - Unique token index

#### Foreign Key Indexes
- `users(tenant_id)` - Tenant relationship
- `user_sessions(user_id, tenant_id)` - User session lookup
- `refresh_tokens(user_id, tenant_id)` - Refresh token lookup

#### Query Optimization Indexes
- `users(email, tenant_id)` - Login lookup
- `users(user_type)` - Role-based queries
- `users(status)` - Status filtering
- `user_sessions(expires_at)` - Expiration cleanup
- `failed_login_attempts(email, ip_address, tenant_id)` - Lockout checks

## Performance Considerations

### Database Performance
- **Connection Pooling**: GORM connection pool management
- **Query Optimization**: Proper indexing strategy
- **Transaction Management**: Minimal transaction scopes
- **Batch Operations**: Bulk inserts/updates where possible

### Caching Strategy
- **Redis Sessions**: Fast session retrieval
- **Token Blacklisting**: JWT invalidation support
- **Rate Limit Counters**: IP-based rate limiting

### Scalability Features
- **Stateless Authentication**: JWT-based auth
- **Horizontal Scaling**: No server-side sessions
- **Database Sharding**: Tenant-based partitioning ready
- **CDN Support**: Static asset optimization

## Deployment Architecture

### Development Environment
```
┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API   │
│   React Dev     │◄──►│   Gin Server    │
│   localhost:3000│    │ localhost:8080  │
└─────────────────┘    └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│ PostgreSQL Dev  │    │   Redis Dev     │
│   localhost:5432 │    │  localhost:6379 │
└─────────────────┘    └─────────────────┘
```

### Production Environment
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   API Servers   │    │   PostgreSQL    │
│   (Nginx/ALB)   │◄──►│   (Kubernetes)  │◄──►│   (RDS/Aurora)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         ▲                       ▲                       ▲
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CDN/CloudFront│    │     Redis       │    │   Read Replicas │
│                 │    │   (ElastiCache) │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Container Architecture
```dockerfile
# Multi-stage Docker build
FROM golang:1.23-alpine AS builder
# Build Go binary

FROM alpine:latest AS runtime
# Runtime dependencies
COPY --from=builder /app/auth-service /usr/local/bin/
EXPOSE 8080
CMD ["auth-service"]
```

## Monitoring & Observability

### Health Checks
- **Application Health**: `/health` endpoint
- **Database Connectivity**: PostgreSQL connection check
- **Cache Availability**: Redis connection check
- **Dependency Status**: External service health

### Logging Strategy
- **Structured Logging**: JSON format with consistent fields
- **Log Levels**: DEBUG, INFO, WARN, ERROR
- **Request Tracing**: Request ID correlation
- **Security Events**: Authentication failures, suspicious activity

### Metrics Collection
- **Application Metrics**: Request count, latency, error rates
- **Database Metrics**: Connection pools, query performance
- **Security Metrics**: Failed login attempts, rate limit hits
- **Business Metrics**: User registrations, tenant creation

## Disaster Recovery

### Backup Strategy
- **Database Backups**: Daily automated backups
- **Configuration Backup**: Environment variables and secrets
- **Code Repository**: Git-based version control
- **Container Images**: Registry-based image storage

### Recovery Procedures
- **Database Recovery**: Point-in-time recovery capability
- **Application Rollback**: Container image rollback
- **Configuration Restore**: Environment variable restoration
- **Data Integrity**: Post-recovery validation checks

## Future Architecture Considerations

### Microservices Evolution
```
Current: Monolithic Auth Service
Future: Auth Microservices
├── User Service
├── Tenant Service
├── Session Service
├── Token Service
└── Audit Service
```

### API Gateway Integration
```
Client → API Gateway → Auth Service → Business Services
    ↓         ↓            ↓            ↓
Routing  Authentication  Authorization  Business Logic
Rate      JWT Validation  Permissions   Domain Logic
Limit     Tenant Context  Role Checks   Data Access
Logging   Request        Audit Trail   Response
          Transformation              Formatting
```

### Event-Driven Architecture
```
Commands → Command Handlers → Events → Event Handlers
    ↓            ↓            ↓           ↓
User       Create User     User        Send Welcome
Registration Handler      Created     Email
                          Event       Update Search
                                      Index
```

This architecture provides a solid foundation for a scalable, secure, and maintainable multi-tenant authentication service while allowing for future evolution and enhancement.