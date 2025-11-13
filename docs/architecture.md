# Architecture Documentation

## Overview

The SaaS Authentication Microservice is a production-ready authentication system built with Go, designed to handle multi-tenant user management, secure authentication, and comprehensive security features for the Blocksure platform.

## System Architecture

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │    │  Auth Service   │    │   Database      │
│                 │    │                 │    │   (PostgreSQL)  │
│ - Load Balancing│◄──►│ - JWT Auth      │◄──►│ - Users         │
│ - Rate Limiting │    │ - User Mgmt     │    │ - Tenants       │
│ - CORS          │    │ - Session Mgmt  │    │ - Sessions      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         ▲                       ▲                       ▲
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Redis Cache   │    │   Email Service │
│   (React)       │    │   (Sessions)    │    │   (SendGrid)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Component Architecture

```
auth-service/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP request handlers (REST API)
│   ├── middleware/      # HTTP middleware (auth, CORS, security)
│   ├── models/          # Data models (GORM structs)
│   ├── repository/      # Data access layer (interfaces + implementations)
│   ├── service/         # Business logic layer
│   └── seeder/          # Database seeding
├── pkg/                 # Shared packages
│   ├── jwt/            # JWT token utilities
│   ├── password/       # Password hashing utilities
│   └── validation/     # Input validation
├── tests/              # Test suites
└── docs/               # Documentation
```

## Design Patterns

### Clean Architecture

The service follows Clean Architecture principles:

- **Entities/Models**: Core business objects (User, Tenant, Session)
- **Use Cases/Services**: Application business rules
- **Interface Adapters/Handlers**: Controllers adapting HTTP to business logic
- **Frameworks/Drivers**: External concerns (Database, HTTP framework)

### Dependency Injection

Services are composed with dependency injection:

```go
// Service composition in main.go
authService := service.NewAuthService(repo, jwtService, passwordService)
authHandler := handler.NewAuthHandler(authService)
```

### Repository Pattern

Data access is abstracted through repository interfaces:

```go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id string) (*models.User, error)
    GetByEmail(ctx context.Context, email, tenantID string) (*models.User, error)
    // ... more methods
}
```

## Security Architecture

### Authentication Flow

```
1. User Login Request
       ↓
2. Validate Credentials
       ↓
3. Generate JWT Tokens
       ↓
4. Create Session
       ↓
5. Return Tokens
```

### Authorization

- **JWT-based Authentication**: Stateless token validation
- **Role-based Access Control**: Admin, Student, RTO, Issuer, Validator roles
- **Tenant Isolation**: Multi-tenant data separation
- **Session Management**: Redis-backed session storage with TTL

### Security Features

- **Password Security**: Argon2id hashing with salt
- **Token Security**: RSA-signed JWTs with configurable expiry
- **Rate Limiting**: Request throttling per IP/user
- **Input Validation**: Comprehensive validation with custom rules
- **CORS Protection**: Configurable origin restrictions
- **CSRF Protection**: Token-based CSRF prevention
- **Security Headers**: HSTS, CSP, X-Frame-Options, etc.

## Data Architecture

### Database Schema

```
┌─────────────────┐    ┌─────────────────┐
│     tenants     │    │      users      │
├─────────────────┤    ├─────────────────┤
│ id (UUID)       │    │ id (UUID)       │
│ name            │    │ email           │◄──┐
│ domain          │    │ password_hash   │   │
│ created_at      │    │ user_type       │   │
│ updated_at      │    │ tenant_id (FK)  │───┘
└─────────────────┘    │ is_active       │
                       │ first_name      │
                       │ last_name       │
                       │ phone           │
                       │ created_at      │
                       │ updated_at      │
                       │ last_login_at   │
                       └─────────────────┘
                              ▲
                              │
                       ┌─────────────────┐
                       │ user_sessions   │
                       ├─────────────────┤
                       │ id (UUID)       │
                       │ user_id (FK)    │
                       │ tenant_id (FK)  │
                       │ token_hash      │
                       │ expires_at      │
                       │ created_at      │
                       │ last_activity   │
                       │ is_active       │
                       │ device_fingerprint│
                       │ location        │
                       │ revoked_reason  │
                       └─────────────────┘
```

### Data Flow

1. **Registration**: User → Handler → Service → Repository → Database
2. **Login**: User → Handler → Service → Repository → JWT Service → Response
3. **Session Check**: Request → Middleware → JWT Validation → User Context
4. **Admin Operations**: Request → Middleware → Admin Check → Service → Repository

## API Design

### RESTful Endpoints

- `POST /auth/register` - User registration
- `POST /auth/login` - User authentication
- `POST /auth/refresh` - Token refresh
- `POST /auth/logout` - User logout
- `GET /user/profile` - Get user profile
- `PUT /user/profile` - Update user profile
- `GET /admin/users` - List users (admin)
- `PUT /admin/users/{id}/activate` - Activate user (admin)

### Response Format

All API responses follow a consistent format:

```json
{
  "success": true,
  "data": { ... },
  "message": "Operation successful"
}
```

### Error Handling

```json
{
  "success": false,
  "message": "Error description",
  "errors": "Detailed validation errors"
}
```

## Performance Considerations

### Caching Strategy

- **Redis**: Session storage, rate limiting, temporary data
- **Database Indexes**: Optimized queries for user lookups, sessions
- **Connection Pooling**: Configurable database connection limits

### Async Processing

- **Password Hashing**: CPU-intensive operations run in goroutines
- **Email Sending**: Asynchronous email delivery
- **Background Jobs**: Session cleanup, audit log processing

### Scalability

- **Horizontal Scaling**: Stateless design supports multiple instances
- **Database Sharding**: Tenant-based data partitioning
- **Load Balancing**: API Gateway distributes requests
- **Rate Limiting**: Prevents abuse and ensures fair usage

## Monitoring & Observability

### Health Checks

- **Liveness Probe**: Application startup status
- **Readiness Probe**: Database and Redis connectivity
- **Health Endpoint**: `/health` with detailed status

### Logging

- **Structured Logging**: JSON format with consistent fields
- **Audit Logging**: Security events and admin actions
- **Request Tracing**: Request IDs for debugging
- **Error Tracking**: Comprehensive error logging

### Metrics

- **Performance Metrics**: Response times, throughput
- **Business Metrics**: User registrations, login attempts
- **System Metrics**: CPU, memory, database connections
- **Security Metrics**: Failed login attempts, suspicious activity

## Deployment Architecture

### Containerization

```dockerfile
FROM golang:1.23-alpine AS builder
# Build stage

FROM alpine:latest
# Runtime stage with minimal footprint
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: auth-service
        image: auth-service:latest
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
```

### Environment Configuration

- **Development**: Local PostgreSQL, Redis, hot reload
- **Staging**: Isolated environment with production-like setup
- **Production**: Kubernetes with secrets management, monitoring

## Development Workflow

### Local Development

1. **Setup**: Docker Compose for dependencies
2. **Development**: Hot reload with CompileDaemon
3. **Testing**: Isolated test databases
4. **Debugging**: Structured logging and request tracing

### CI/CD Pipeline

1. **Build**: Go compilation and Docker image creation
2. **Test**: Unit tests, integration tests, security scans
3. **Deploy**: Kubernetes deployment with rolling updates
4. **Monitor**: Health checks and alerting

## Security Considerations

### Threat Mitigation

- **SQL Injection**: Parameterized queries with GORM
- **XSS**: Input validation and output encoding
- **CSRF**: Token-based protection for state changes
- **Session Hijacking**: Secure session management
- **Brute Force**: Account lockout and rate limiting

### Compliance

- **Data Encryption**: Passwords hashed with Argon2id
- **Audit Logging**: All admin actions logged
- **Access Control**: Principle of least privilege
- **Data Privacy**: Tenant data isolation

## Future Enhancements

### Planned Features

- **OAuth Integration**: Social login providers
- **MFA**: Multi-factor authentication
- **API Keys**: Service account authentication
- **Audit Dashboard**: Admin interface for audit logs
- **Advanced Analytics**: User behavior insights

### Scalability Improvements

- **Database Sharding**: Horizontal database scaling
- **Caching Layer**: Enhanced Redis usage
- **Message Queue**: Async job processing
- **CDN Integration**: Static asset delivery