# SaaS Authentication Microservice

A production-ready authentication microservice built with Go, providing multi-tenant user management, JWT authentication, and comprehensive security features.

## Features

- **Multi-tenant Architecture**: Domain-based and header-based tenant isolation
- **JWT Authentication**: RSA-signed tokens with configurable expiry
- **Secure Password Hashing**: Argon2id algorithm for password security
- **Session Management**: Redis-backed session storage with TTL
- **User Management**: Complete CRUD operations with role-based access
- **Password Reset**: Secure token-based password reset flow
- **Rate Limiting**: Configurable rate limiting for API protection
- **Admin Panel**: Administrative functions for user and tenant management
- **Health Checks**: Comprehensive health monitoring
- **Docker Support**: Containerized deployment with docker-compose

## Tech Stack

- **Backend**: Go 1.21, Gin web framework
- **Database**: PostgreSQL with GORM ORM
- **Cache**: Redis for session management
- **Security**: JWT with RSA signing, Argon2id password hashing
- **Container**: Docker with multi-stage builds

## Quick Start

### Prerequisites

- Go 1.21+
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

4. **Run with Docker Compose:**
   ```bash
   docker-compose up -d
   ```

5. **Run locally:**
   ```bash
   go mod tidy
   go run cmd/server/main.go
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

## Development

### Project Structure
```
auth-service/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data models
│   ├── repository/      # Data access layer
│   └── service/         # Business logic layer
├── pkg/
│   ├── jwt/            # JWT utilities
│   ├── password/       # Password utilities
│   └── validation/     # Input validation
├── docker-compose.yml   # Development environment
├── Dockerfile          # Container definition
├── go.mod             # Go modules
└── README.md          # This file
```

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
go build -o auth-service ./cmd/server
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