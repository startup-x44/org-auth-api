# Development Guide

This guide covers setting up a development environment and contributing to the SaaS Authentication Microservice.

## Prerequisites

- **Go**: 1.23 or later
- **PostgreSQL**: 15 or later
- **Redis**: 7 or later
- **Docker**: 20.10 or later
- **Docker Compose**: 2.0 or later
- **Git**: Latest version
- **Make**: Build automation (optional)

## Quick Start

### 1. Clone and Setup

```bash
# Clone the repository
git clone <repository-url>
cd auth-service

# Copy environment file
cp .env.example .env

# Edit environment variables
nano .env
```

### 2. Development Environment

```bash
# Start all services (PostgreSQL, Redis)
./dev.sh dev

# Or run in background
./dev.sh dev-d

# View logs
./dev.sh logs

# Stop services
./dev.sh stop
```

### 3. Manual Setup (Alternative)

```bash
# Start dependencies with Docker Compose
docker-compose -f docker-compose.dev.yml up -d postgres redis

# Install Go dependencies
go mod tidy

# Run database migrations
go run cmd/server/main.go migrate

# Start the application
go run cmd/server/main.go
```

### 4. Verify Setup

```bash
# Test health endpoint
curl http://localhost:8080/health

# Test user registration
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123!",
    "confirm_password": "TestPass123!",
    "user_type": "student",
    "tenant_id": "default.local"
  }'
```

## Development Workflow

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
│   ├── service/         # Business logic
│   └── seeder/          # Database seeding
├── pkg/                 # Shared packages
├── tests/               # Test suites
├── docs/                # Documentation
├── docker/              # Docker files
├── k8s/                 # Kubernetes manifests
└── migrations/          # Database migrations
```

### Code Organization

#### Handlers (API Layer)

Handlers are responsible for:
- HTTP request/response handling
- Input validation and binding
- Response formatting
- Error handling

```go
func (h *AuthHandler) Register(c *gin.Context) {
    var req service.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // Handle validation error
        return
    }

    response, err := h.authService.UserService().Register(c.Request.Context(), &req)
    if err != nil {
        // Handle business logic error
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "success": true,
        "data":    response,
    })
}
```

#### Services (Business Logic Layer)

Services contain:
- Business rules and logic
- Data transformation
- External service integration
- Transaction management

```go
func (s *userService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
    // Validate request
    if err := s.validateRegistrationRequest(req); err != nil {
        return nil, err
    }

    // Create user
    user, err := s.createUserAccount(ctx, req, tenantID)
    if err != nil {
        return nil, err
    }

    // Generate tokens
    tokenPair, err := s.generateTokenPair(user)
    if err != nil {
        return nil, err
    }

    return &RegisterResponse{
        User:  s.convertToUserProfile(user),
        Token: tokenPair,
    }, nil
}
```

#### Repositories (Data Access Layer)

Repositories handle:
- Database operations
- Query building
- Connection management
- Data mapping

```go
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByEmail(ctx context.Context, email, tenantID string) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).
        Where("email = ? AND tenant_id = ?", email, tenantID).
        First(&user).Error
    return &user, err
}
```

### Development Scripts

Use the included `dev.sh` script for common tasks:

```bash
# Development
./dev.sh dev          # Start development environment
./dev.sh dev-d        # Start in background
./dev.sh logs         # View logs
./dev.sh shell        # Open shell in container
./dev.sh stop         # Stop services

# Testing
./dev.sh test         # Run all tests
./dev.sh test-unit    # Run unit tests only
./dev.sh test-int     # Run integration tests only

# Database
./dev.sh db-reset     # Reset database
./dev.sh db-seed      # Seed database

# Cleanup
./dev.sh clean        # Remove containers and volumes
```

## Testing

### Test Structure

```
tests/
├── fixtures/         # Test data fixtures
├── testutils/        # Test utilities and helpers
├── unit/            # Unit tests (services, repositories)
│   ├── service/
│   └── repository/
└── feature/         # Feature tests (API endpoints)
```

### Running Tests

```bash
# Run all tests
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# Run specific test types
go test ./tests/unit/... -v
go test ./tests/feature/... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run benchmarks
go test ./pkg/password -bench=.
go test ./pkg/jwt -bench=.
```

### Writing Tests

#### Unit Tests

```go
func TestUserService_Register(t *testing.T) {
    // Setup
    mockRepo := &mocks.UserRepository{}
    svc := service.NewUserService(mockRepo, jwtSvc, pwdSvc)

    // Expectations
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    // Test
    req := &service.RegisterRequest{
        Email:    "test@example.com",
        Password: "TestPass123!",
    }

    response, err := svc.Register(context.Background(), req)

    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, response)
    mockRepo.AssertExpectations(t)
}
```

#### Integration Tests

```go
func TestAuthAPI_Register(t *testing.T) {
    // Setup test database
    db := testutils.SetupTestDB(t)
    defer testutils.CleanupTestDB(t, db)

    // Setup services
    repo := repository.NewRepository(db)
    // ... setup other dependencies

    // Create test server
    router := setupTestRouter(repo, jwtSvc, pwdSvc)

    // Test request
    req := `{
        "email": "test@example.com",
        "password": "TestPass123!",
        "confirm_password": "TestPass123!",
        "user_type": "student",
        "tenant_id": "default.local"
    }`

    w := httptest.NewRecorder()
    r := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(req))
    r.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, r)

    // Assertions
    assert.Equal(t, http.StatusCreated, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response["success"].(bool))
}
```

### Test Data

The service includes seeded test data for development:

```go
// Seed development data
seeder := seeder.NewDatabaseSeeder(db)
ctx := context.Background()
err := seeder.Seed(ctx)
```

## Debugging

### Logging

The application uses structured logging:

```go
// Info logging
logger.Info("User registered", "user_id", userID, "email", email)

// Error logging
logger.Error("Failed to create user", "error", err, "email", email)

// Audit logging
auditLogger.LogAdminAction(userID, "create_user", "user", newUserID, ip, userAgent, true, nil, "User created successfully")
```

### Request Tracing

Each request includes a unique request ID:

```go
// Add to context
ctx := context.WithValue(r.Context(), "request_id", requestID)

// Log with request ID
logger.Info("Processing request", "request_id", requestID, "path", r.URL.Path)
```

### Debug Mode

Enable debug logging:

```bash
export LOG_LEVEL=debug
```

## Code Quality

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linting
golangci-lint run
```

### Code Formatting

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...
```

### Pre-commit Hooks

```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install

# Run on all files
pre-commit run --all-files
```

## Database Development

### Migrations

```go
// Create new migration
migrate create -ext sql -dir migrations -seq add_user_preferences

// Run migrations
migrate -path migrations -database "postgres://..." up

// Rollback
migrate -path migrations -database "postgres://..." down 1
```

### Seeding

```go
func (s *databaseSeeder) Seed(ctx context.Context) error {
    // Create tenants
    tenants := []models.Tenant{
        {Name: "Default Organization", Domain: "default.local"},
        {Name: "Demo Company", Domain: "demo.company.com"},
    }

    for _, tenant := range tenants {
        if err := s.db.Create(&tenant).Error; err != nil {
            return err
        }
    }

    // Create users
    users := []models.User{
        {
            Email:     "admin@default.local",
            UserType:  "admin",
            TenantID:  tenants[0].ID,
            IsActive:  true,
        },
    }

    for _, user := range users {
        hashedPassword, _ := s.passwordService.Hash("Admin123!")
        user.PasswordHash = hashedPassword
        if err := s.db.Create(&user).Error; err != nil {
            return err
        }
    }

    return nil
}
```

## API Development

### Adding New Endpoints

1. **Define Request/Response Types**

```go
type CreateTenantRequest struct {
    Name   string `json:"name" validate:"required,min=2,max=100"`
    Domain string `json:"domain" validate:"required,fqdn"`
}

type CreateTenantResponse struct {
    Tenant *Tenant `json:"tenant"`
}
```

2. **Add Repository Methods**

```go
func (r *tenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
    return r.db.WithContext(ctx).Create(tenant).Error
}
```

3. **Add Service Methods**

```go
func (s *tenantService) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*Tenant, error) {
    // Validation
    if err := validation.ValidateTenant(req); err != nil {
        return nil, err
    }

    // Check domain uniqueness
    existing, _ := s.repo.Tenant().GetByDomain(ctx, req.Domain)
    if existing != nil {
        return nil, errors.New("domain already exists")
    }

    // Create tenant
    tenant := &models.Tenant{
        Name:   req.Name,
        Domain: req.Domain,
    }

    if err := s.repo.Tenant().Create(ctx, tenant); err != nil {
        return nil, err
    }

    return s.convertToTenant(tenant), nil
}
```

4. **Add Handler Methods**

```go
func (h *AdminHandler) CreateTenant(c *gin.Context) {
    var req service.CreateTenantRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid request data",
            "errors":  err.Error(),
        })
        return
    }

    tenant, err := h.adminService.CreateTenant(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": err.Error(),
        })
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "success": true,
        "data":    tenant,
        "message": "Tenant created successfully",
    })
}
```

5. **Add Routes**

```go
admin := v1.Group("/admin")
admin.Use(authMiddleware.AuthRequired())
admin.Use(authMiddleware.AdminRequired())
{
    admin.POST("/tenants", adminHandler.CreateTenant)
}
```

6. **Add Tests**

```go
func TestAdminAPI_CreateTenant(t *testing.T) {
    // Setup test database and services
    // Test the endpoint
    // Assert response
}
```

## Performance Optimization

### Profiling

```go
import _ "net/http/pprof"

// Access profiling at http://localhost:8080/debug/pprof/
```

### Benchmarking

```go
func BenchmarkPasswordHash(b *testing.B) {
    svc := password.NewService()
    password := "BenchmarkTestPassword123!"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := svc.Hash(password)
        if err != nil {
            b.Fatalf("Failed to hash password: %v", err)
        }
    }
}
```

### Database Optimization

```sql
-- Add indexes for performance
CREATE INDEX CONCURRENTLY idx_users_email_tenant ON users(email, tenant_id);
CREATE INDEX CONCURRENTLY idx_sessions_user_expires ON user_sessions(user_id, expires_at);

-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com' AND tenant_id = 'tenant-uuid';
```

## Contributing

### Git Workflow

```bash
# Create feature branch
git checkout -b feature/new-endpoint

# Make changes
# Add tests
# Update documentation

# Commit changes
git add .
git commit -m "Add new endpoint for user preferences"

# Push branch
git push origin feature/new-endpoint

# Create pull request
```

### Code Review Checklist

- [ ] Tests pass
- [ ] Code is properly formatted (`go fmt`)
- [ ] No linting errors (`golangci-lint run`)
- [ ] Documentation updated
- [ ] Security considerations addressed
- [ ] Performance impact assessed
- [ ] Database migrations included if needed

### Commit Message Format

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Testing
- `chore`: Maintenance

## Troubleshooting

### Common Issues

1. **Database Connection Issues**
   ```bash
   # Check PostgreSQL
   docker-compose -f docker-compose.dev.yml ps postgres

   # View logs
   docker-compose -f docker-compose.dev.yml logs postgres
   ```

2. **Redis Connection Issues**
   ```bash
   # Check Redis
   docker-compose -f docker-compose.dev.yml ps redis

   # Test connection
   docker-compose -f docker-compose.dev.yml exec redis redis-cli ping
   ```

3. **Port Conflicts**
   ```bash
   # Find process using port
   lsof -i :8080

   # Kill process
   kill -9 <PID>
   ```

4. **Go Module Issues**
   ```bash
   # Clean module cache
   go clean -modcache

   # Re-download dependencies
   go mod download

   # Tidy modules
   go mod tidy
   ```

### Debug Commands

```bash
# View application logs
docker-compose -f docker-compose.dev.yml logs -f auth-service

# Execute into container
docker-compose -f docker-compose.dev.yml exec auth-service sh

# Test API endpoints
curl -v http://localhost:8080/health

# Database shell
docker-compose -f docker-compose.dev.yml exec postgres psql -U auth_user -d auth_db
```

This development guide provides comprehensive information for setting up and contributing to the authentication service.