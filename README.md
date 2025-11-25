# Auth Service - Backend

Production-ready authentication and authorization service built with Go, featuring multi-tenancy, RBAC, OAuth 2.1 + PKCE, and comprehensive security features.

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Database Migrations](#database-migrations)
- [Development](#development)
- [Deployment](#deployment)
- [API Documentation](#api-documentation)
- [Security](#security)

## Features

### Core Authentication
- ✅ User registration and login
- ✅ Email verification
- ✅ Password reset flow
- ✅ JWT access & refresh tokens
- ✅ Session management
- ✅ CSRF protection
- ✅ Rate limiting

### Multi-Tenancy & RBAC
- ✅ Organization-based multi-tenancy
- ✅ Role-Based Access Control (RBAC)
- ✅ Custom permissions per organization
- ✅ System-level permissions for superadmins
- ✅ Organization invitations

### OAuth 2.1 & API Keys
- ✅ OAuth 2.1 Authorization Code Flow
- ✅ PKCE (Proof Key for Code Exchange)
- ✅ Client application management
- ✅ API key generation and management
- ✅ Scope-based access control

### Security Features
- ✅ Bcrypt password hashing
- ✅ Failed login attempt tracking
- ✅ Account lockout protection
- ✅ CORS configuration
- ✅ Security headers
- ✅ Input validation & sanitization

### Observability
- ✅ Structured logging
- ✅ Prometheus metrics
- ✅ Health check endpoints
- ✅ Audit logging

## Prerequisites

- **Go**: 1.23 or higher
- **PostgreSQL**: 15 or higher
- **Redis**: 7 or higher
- **Docker & Docker Compose** (for containerized deployment)

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd auth-service
```

### 2. Set Up Environment Variables

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Server
SERVER_PORT=8080
ENVIRONMENT=development

# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=auth_user
DATABASE_PASSWORD=your_secure_password
DATABASE_NAME=auth_db
DATABASE_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# Email (optional for development)
EMAIL_ENABLED=false
```

### 3. Start Dependencies (PostgreSQL & Redis)

**Using Docker Compose:**

```bash
docker-compose up -d postgres redis
```

**Or install locally:**
- PostgreSQL: https://www.postgresql.org/download/
- Redis: https://redis.io/download

### 4. Run Database Migrations

```bash
# Build the migration tool
go build -o migrate ./cmd/migrate

# Run migrations
./migrate
```

### 5. Run the Application

**Development mode:**

```bash
go run ./cmd/server
```

**Or build and run:**

```bash
go build -o auth-service ./cmd/server
./auth-service
```

The server will start on `http://localhost:8080`

### 6. Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Should return: {"status":"healthy"}
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_PORT` | HTTP server port | `8080` | Yes |
| `ENVIRONMENT` | Environment (development/staging/production) | `development` | Yes |
| `DATABASE_HOST` | PostgreSQL host | `localhost` | Yes |
| `DATABASE_PORT` | PostgreSQL port | `5432` | Yes |
| `DATABASE_USER` | PostgreSQL username | `auth_user` | Yes |
| `DATABASE_PASSWORD` | PostgreSQL password | - | Yes |
| `DATABASE_NAME` | PostgreSQL database name | `auth_db` | Yes |
| `DATABASE_SSLMODE` | PostgreSQL SSL mode | `disable` | Yes |
| `REDIS_HOST` | Redis host | `localhost` | Yes |
| `REDIS_PORT` | Redis port | `6379` | Yes |
| `REDIS_PASSWORD` | Redis password | - | No |
| `REDIS_DB` | Redis database number | `0` | No |
| `JWT_SECRET` | Secret key for JWT signing | - | Yes |
| `JWT_ACCESS_TOKEN_EXPIRY` | Access token expiry (seconds) | `3600` | No |
| `JWT_REFRESH_TOKEN_EXPIRY` | Refresh token expiry (seconds) | `604800` | No |
| `EMAIL_ENABLED` | Enable email sending | `false` | No |
| `SMTP_HOST` | SMTP server host | - | If email enabled |
| `SMTP_PORT` | SMTP server port | `587` | If email enabled |
| `SMTP_USERNAME` | SMTP username | - | If email enabled |
| `SMTP_PASSWORD` | SMTP password | - | If email enabled |
| `FRONTEND_URL` | Frontend URL for links | `http://localhost:3000` | Yes |
| `RATE_LIMIT_REQUESTS` | Max requests per window | `100` | No |
| `RATE_LIMIT_WINDOW` | Rate limit window (seconds) | `60` | No |

See `.env.example` for a complete configuration template.

## Database Migrations

### Running Migrations

The service uses GORM AutoMigrate for database schema management.

```bash
# Build the migration tool
go build -o migrate ./cmd/migrate

# Run migrations
./migrate
```

### Migration Tables

The following tables will be created:

- **users** - User accounts (global)
- **organizations** - Tenant organizations
- **organization_memberships** - User-organization relationships
- **organization_invitations** - Pending invitations
- **roles** - RBAC roles (system and org-level)
- **permissions** - RBAC permissions (system and org-level)
- **role_permissions** - Role-permission mappings
- **user_sessions** - Active user sessions
- **refresh_tokens** - OAuth refresh tokens
- **password_resets** - Password reset tokens
- **failed_login_attempts** - Failed login tracking
- **client_apps** - OAuth client applications
- **oauth_authorization_codes** - OAuth authorization codes
- **api_keys** - API key management
- **audit_logs** - Audit trail

### Manual SQL Migrations

For custom migrations, use the `migrations/` directory:

```bash
# Example: Run a custom migration
psql -h localhost -U auth_user -d auth_db -f migrations/001_custom_migration.sql
```

## Development

### Using Docker Compose (Recommended)

Start the entire development environment:

```bash
# Start all services (app + postgres + redis)
./dev.sh dev

# Or in background
./dev.sh dev-d

# View logs
./dev.sh logs

# Stop services
./dev.sh stop

# Clean up (removes volumes)
./dev.sh clean
```

### Development Scripts

- `./dev.sh dev` - Start development environment with live reload
- `./dev.sh dev-d` - Start in background
- `./dev.sh stop` - Stop development environment
- `./dev.sh test` - Run test suite
- `./dev.sh clean` - Clean up containers and volumes
- `./dev.sh logs` - Show logs
- `./dev.sh shell` - Open shell in container

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./tests/integration -v

# Run tests in Docker
./dev.sh test
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Vet code
go vet ./...
```

## Deployment

### Docker Deployment

#### 1. Build Docker Image

```bash
docker build -t auth-service:latest .
```

#### 2. Run with Docker Compose

```bash
# Production deployment
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f auth-service
```

#### 3. Environment Variables

Update `docker-compose.yml` or use a `.env` file:

```yaml
environment:
  - DATABASE_PASSWORD=${DATABASE_PASSWORD}
  - JWT_SECRET=${JWT_SECRET}
  - SMTP_PASSWORD=${SMTP_PASSWORD}
```

### Kubernetes Deployment

See [k8s/README.md](./k8s/README.md) for detailed Kubernetes deployment instructions.

#### Quick Deploy to Kubernetes

```bash
# Create namespace
kubectl create namespace auth-service

# Create secrets (update values first)
kubectl create secret generic auth-secrets \
  --from-literal=jwt-secret=your-secret \
  --from-literal=database-password=your-db-pass \
  -n auth-service

# Deploy all resources
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -n auth-service
kubectl logs -f deployment/auth-service -n auth-service
```

#### Kubernetes Resources

- **Deployment**: 3 replicas with health checks
- **Service**: ClusterIP for internal access
- **Ingress**: External access with TLS
- **HPA**: Auto-scaling (3-10 pods)
- **ConfigMap**: Non-sensitive configuration
- **Secret**: Sensitive credentials
- **NetworkPolicy**: Security rules

### Production Checklist

Before deploying to production:

- [ ] Change `JWT_SECRET` to a strong random value
- [ ] Use strong database passwords
- [ ] Enable SSL/TLS for database (`DATABASE_SSLMODE=require`)
- [ ] Configure proper CORS origins
- [ ] Set `ENVIRONMENT=production`
- [ ] Enable email service with production SMTP
- [ ] Set up monitoring and alerting
- [ ] Configure backup strategy for database
- [ ] Review and adjust rate limits
- [ ] Set up log aggregation
- [ ] Enable audit logging
- [ ] Configure proper ingress/load balancer
- [ ] Set resource limits in Kubernetes
- [ ] Enable pod disruption budgets
- [ ] Set up database connection pooling
- [ ] Review security headers

## API Documentation

### Health Endpoints

```bash
# Liveness check
GET /health
GET /health/live

# Readiness check (includes DB connection)
GET /health/ready

# Metrics (Prometheus format)
GET /metrics
```

### Authentication Endpoints

```bash
# Register
POST /api/v1/auth/register
Body: { "email": "user@example.com", "password": "SecureP@ss123", ... }

# Login
POST /api/v1/auth/login
Body: { "email": "user@example.com", "password": "SecureP@ss123" }

# Refresh token
POST /api/v1/auth/refresh
Body: { "refresh_token": "..." }

# Logout
POST /api/v1/user/logout
Headers: Authorization: Bearer <token>
```

### OAuth 2.1 Endpoints

```bash
# Authorization endpoint
GET /api/v1/oauth/authorize?client_id=...&redirect_uri=...&response_type=code&code_challenge=...

# Token endpoint
POST /api/v1/oauth/token
Body: { "grant_type": "authorization_code", "code": "...", "code_verifier": "..." }
```

### User Management (Superadmin)

```bash
# List users
GET /api/v1/admin/users
Headers: Authorization: Bearer <superadmin-token>

# Get user
GET /api/v1/admin/users/:id

# Update user
PUT /api/v1/admin/users/:id

# Delete user
DELETE /api/v1/admin/users/:id
```

For complete API documentation, see [docs/API_FLOWS.md](./docs/API_FLOWS.md)

## Security

### Security Features

- **Password Hashing**: Bcrypt with configurable cost
- **JWT Tokens**: RS256 signing with short expiry
- **CSRF Protection**: Double-submit cookie pattern
- **Rate Limiting**: Per-IP and per-user limits
- **Failed Login Tracking**: Account lockout after failed attempts
- **Input Validation**: Comprehensive validation on all inputs
- **SQL Injection Prevention**: Parameterized queries with GORM
- **XSS Prevention**: Input sanitization
- **CORS**: Configurable allowed origins

### Security Best Practices

1. **Secrets Management**: Never commit secrets to version control
2. **HTTPS Only**: Always use HTTPS in production
3. **Database Security**: Use SSL/TLS for database connections
4. **Regular Updates**: Keep dependencies updated
5. **Audit Logs**: Monitor and review audit logs regularly
6. **Backup**: Regular database backups
7. **Monitoring**: Set up alerts for suspicious activity

## Monitoring

### Prometheus Metrics

The service exposes Prometheus metrics at `/metrics`:

- HTTP request duration
- HTTP request count by status code
- Database connection pool stats
- Redis connection stats
- Custom business metrics

### Logging

Structured JSON logging with levels:
- `DEBUG`: Detailed debugging information
- `INFO`: General information
- `WARN`: Warning messages
- `ERROR`: Error messages

Configure log level via `LOG_LEVEL` environment variable.

## Troubleshooting

### Common Issues

#### Database Connection Failed

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check database credentials
psql -h localhost -U auth_user -d auth_db

# Check database logs
docker-compose logs postgres
```

#### Redis Connection Failed

```bash
# Check Redis is running
docker-compose ps redis

# Test Redis connection
redis-cli ping

# Check Redis logs
docker-compose logs redis
```

#### Migration Errors

```bash
# Reset database (CAUTION: deletes all data)
docker-compose down -v
docker-compose up -d postgres
./migrate

# Or manually drop and recreate
psql -h localhost -U postgres -c "DROP DATABASE auth_db;"
psql -h localhost -U postgres -c "CREATE DATABASE auth_db;"
./migrate
```

#### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use a different port
SERVER_PORT=8081 go run ./cmd/server
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

Proprietary - NiloAuth Auth Service

## Support

For issues and questions:
- Create an issue in the repository
- Contact the development team
