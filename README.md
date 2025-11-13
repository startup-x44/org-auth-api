# SaaS Authentication Microservice

A production-ready authentication microservice built with Go, providing multi-tenant user management, JWT authentication, and comprehensive security features for the Blocksure platform.

## Features

- **Multi-tenant Architecture**: Domain-based and header-based tenant isolation
- **JWT Authentication**: RSA-signed tokens with configurable expiry
- **Secure Password Hashing**: Argon2id algorithm for password security
- **Session Management**: Redis-backed session storage with TTL
- **User Management**: Complete CRUD operations with role-based access
- **Password Reset**: Secure token-based password reset flow
- **Rate Limiting**: Configurable rate limiting for API protection
- **Admin Panel**: Administrative functions for user and tenant management
- **Database Seeding**: Automated test data seeding for development
- **Comprehensive Testing**: Unit and feature tests with isolated databases
- **Health Checks**: Comprehensive health monitoring
- **Docker Support**: Containerized deployment with docker-compose

## Tech Stack

- **Backend**: Go 1.23, Gin web framework
- **Database**: PostgreSQL with GORM ORM
- **Cache**: Redis for session management
- **Security**: JWT with RSA signing, Argon2id password hashing
- **Testing**: Testify, isolated test databases
- **Container**: Docker with multi-stage builds

## Quick Start

### Prerequisites

- Go 1.23+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (optional)

### Local Development

1. **Clone and setup:**
   ```bash
   git clone <repository-url>
   cd auth-service
   ```

2. **Environment variables:**
   Create a `.env` file or set environment variables:
   ```bash
   export SERVER_PORT=8080
   export ENVIRONMENT=development
   export DATABASE_HOST=localhost
   export DATABASE_PORT=5432
   export DATABASE_USER=auth_user
   export DATABASE_PASSWORD=auth_password
   export DATABASE_NAME=auth_db
   export DATABASE_SSLMODE=disable
   export REDIS_HOST=localhost
   export REDIS_PORT=6379
   export JWT_SECRET=your-super-secret-jwt-key-change-in-production
   export JWT_ACCESS_TOKEN_EXPIRY=3600
   export JWT_REFRESH_TOKEN_EXPIRY=604800
   ```

3. **Database setup:**
   ```bash
   createdb auth_db
   createuser auth_user --password
   ```

4. **Quick development start (recommended):**
   ```bash
   ./dev.sh dev
   ```

5. **Manual Docker Compose commands:**
   ```bash
   # Development with live reload
   docker-compose -f docker-compose.dev.yml up --build

   # Production
   docker-compose up --build
   ```

6. **Run locally:**
   ```bash
   go mod tidy
   go run cmd/server/main.go
   ```

### Development with Live Reload

For development with automatic code reloading:

1. **Use the development Docker Compose:**
   ```bash
   docker-compose -f docker-compose.dev.yml up --build
   ```

2. **Or run locally with air:**
   ```bash
   # Install air globally
   go install github.com/cosmtrek/air@latest

   # Run with live reload
   air
   ```

The development setup includes:
- **Live reload** using `CompileDaemon` - automatically rebuilds and restarts on code changes
- **Volume mounting** - source code changes are reflected instantly
- **Dependency caching** - Go modules are cached for faster rebuilds
- **Same services** - PostgreSQL and Redis are included in the dev environment

### Production Deployment

For production deployment, use the standard setup:

## Seeded Test Data

The service automatically seeds test data on startup for development:

### Tenants
- `default.local` - Default Organization
- `demo.company.com` - Demo Company
- `test.org` - Test Organization

### Users (All tenants have similar user sets)

| Email | Password | User Type | Description |
|-------|----------|-----------|-------------|
| `admin@{tenant}` | `Admin123!` | Admin | Administrative access |
| `student@{tenant}` | `Student123!` | Student | Student user |
| `rto@{tenant}` | `RTO123!` | RTO | Registered Training Organization |

### Example Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@default.local", "password": "Admin123!", "tenant_id": "default.local"}'
```

## API Documentation

### Authentication Endpoints

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "confirm_password": "SecurePass123!",
  "user_type": "student",
  "tenant_id": "tenant-uuid",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "tenant_id": "tenant-uuid"
}
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "refresh-token-here"
}
```

### User Endpoints

#### Get Profile
```http
GET /api/v1/user/profile
Authorization: Bearer <access-token>
X-Tenant-ID: tenant-uuid
```

#### Update Profile
```http
PUT /api/v1/user/profile
Authorization: Bearer <access-token>
X-Tenant-ID: tenant-uuid
Content-Type: application/json

{
  "first_name": "Jane",
  "last_name": "Smith",
  "phone": "+1234567890"
}
```

### Admin Endpoints

#### List Users
```http
GET /api/v1/admin/users?limit=10&offset=0
Authorization: Bearer <admin-access-token>
X-Tenant-ID: tenant-uuid
```

#### Create Tenant
```http
POST /api/v1/admin/tenants
Authorization: Bearer <admin-access-token>
Content-Type: application/json

{
  "name": "New Organization",
  "domain": "neworg.com"
}
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Server port | `8080` |
| `ENVIRONMENT` | Environment (development/production) | `development` |
| `DATABASE_HOST` | PostgreSQL host | `localhost` |
| `DATABASE_PORT` | PostgreSQL port | `5432` |
| `DATABASE_USER` | PostgreSQL user | `auth_user` |
| `DATABASE_PASSWORD` | PostgreSQL password | `auth_password` |
| `DATABASE_NAME` | PostgreSQL database name | `auth_db` |
| `DATABASE_SSLMODE` | SSL mode | `disable` |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Redis password | (empty) |
| `REDIS_DB` | Redis database | `0` |
| `JWT_SECRET` | JWT signing secret | (required) |
| `JWT_ACCESS_TOKEN_EXPIRY` | Access token expiry (seconds) | `3600` |
| `JWT_REFRESH_TOKEN_EXPIRY` | Refresh token expiry (seconds) | `604800` |
| `PASSWORD_MIN_LENGTH` | Minimum password length | `8` |
| `RATE_LIMIT_REQUESTS` | Rate limit requests per window | `100` |
| `RATE_LIMIT_WINDOW` | Rate limit window (seconds) | `60` |

## Security Features

- **Password Requirements**: Minimum 8 characters, uppercase, lowercase, number, special character
- **JWT Security**: RSA signing with configurable expiry
- **Rate Limiting**: Configurable request limits per time window
- **Input Validation**: Comprehensive validation for all inputs
- **SQL Injection Protection**: Parameterized queries with GORM
- **CORS Protection**: Configurable CORS headers
- **Session Management**: Secure session handling with Redis TTL

## Database Schema

### Users Table
- `id` (UUID, Primary Key)
- `email` (String, Unique per tenant)
- `password` (String, Hashed)
- `type` (String, User role)
- `tenant_id` (UUID, Foreign Key)
- `first_name` (String, Optional)
- `last_name` (String, Optional)
- `phone` (String, Optional)
- `is_active` (Boolean)
- `created_at`, `updated_at`, `last_login_at` (Timestamps)

### Tenants Table
- `id` (UUID, Primary Key)
- `name` (String)
- `domain` (String, Unique)
- `created_at`, `updated_at` (Timestamps)

### Sessions Table
- `id` (UUID, Primary Key)
- `user_id` (UUID, Foreign Key)
- `expires_at` (Timestamp)

### Refresh Tokens Table
- `user_id` (UUID, Primary Key)
- `token` (String, Primary Key)
- `expires_at` (Timestamp)

### Password Resets Table
- `email` (String, Primary Key)
- `tenant_id` (UUID, Primary Key)
- `token` (String)
- `expires_at` (Timestamp)

## Documentation

- **[API Specification](docs/api-spec.yaml)** - OpenAPI 3.0 specification
- **[Postman Collection](docs/postman-collection.json)** - API testing collection
- **[Architecture Guide](docs/architecture.md)** - System design and patterns
- **[Deployment Guide](docs/deployment.md)** - Production deployment instructions
- **[Development Guide](docs/development.md)** - Development setup and workflow
- **[Security Review](docs/SECURITY_CODE_REVIEW_REPORT.md)** - Security assessment report

## Development

### Development Script

Use the included `dev.sh` script for common development tasks:

```bash
# Start development environment with live reload
./dev.sh dev

# Start in background
./dev.sh dev-d

# Stop development environment
./dev.sh stop

# Run tests
./dev.sh test

# View logs
./dev.sh logs

# Open shell in container
./dev.sh shell

# Clean up (removes containers and volumes)
./dev.sh clean
```

### Running Tests

The project includes comprehensive testing with isolated databases:

```bash
# Run all tests
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# Run specific test types
docker-compose -f docker-compose.test.yml run --rm auth-service-test go test ./tests/unit/... -v
docker-compose -f docker-compose.test.yml run --rm auth-service-test go test ./tests/feature/... -v

# Run tests locally (requires Go and PostgreSQL)
go test ./tests/... -v
```

### Database Seeding

The application automatically seeds development data on startup:

- **Tenants**: Default, Demo, and Test organizations
- **Users**: Admin, Student, and RTO accounts for each tenant
- **Sessions**: Sample session data for testing

### Building for Production

```bash
# Build the application
go build -o auth-service ./cmd/server

# Build Docker image
docker build -t auth-service .

# Run container
docker run -p 8080:8080 auth-service
```

## Deployment

### Docker Deployment
```bash
docker build -t auth-service .
docker run -p 8080:8080 auth-service
```

### Kubernetes
Use the provided Kubernetes manifests in the `k8s/` directory for production deployment.

## Health Checks

The service provides health check endpoints:

- `GET /health` - General health check
- Returns database and Redis connectivity status

## Monitoring

- Structured logging with request IDs
- Health check endpoints for load balancers
- Error tracking and reporting
- Performance metrics collection

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.