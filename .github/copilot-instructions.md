# AI Coding Instructions for Auth Service

## Architecture Overview

**Multi-tenant Go microservice** with Gin framework, GORM ORM, PostgreSQL, and Redis. Implements domain-based tenant isolation with JWT authentication and comprehensive security features.

### Key Components
- **Backend**: Go 1.23 + Gin web framework
- **Database**: PostgreSQL with GORM ORM + automatic migrations
- **Cache**: Redis for sessions and rate limiting
- **Security**: RSA-signed JWT tokens, Argon2id password hashing
- **Frontend**: React with tenant-aware localStorage isolation

### Service Boundaries
```
cmd/server/           # Application entry point
internal/
├── config/          # Environment-based configuration
├── handler/         # HTTP request handlers (Gin routes)
├── middleware/      # Authentication, CORS, CSRF, security
├── models/          # GORM database models
├── repository/      # Data access layer with tenant filtering
├── service/         # Business logic layer
└── seeder/          # Database seeding for development
pkg/                 # Shared packages (JWT, password, validation)
tests/               # Isolated test databases
frontend/            # React SPA with tenant resolution
```

## Critical Patterns & Conventions

### 1. Multi-Tenant Architecture
**Tenant isolation is enforced at every layer** - never query without tenant context.

```go
// Repository pattern with automatic tenant filtering
func (r *userRepository) GetByEmail(ctx context.Context, email, tenantID string) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).Where("email = ? AND tenant_id = ?", email, tenantID).First(&user).Error
    return &user, err
}

// Service layer tenant validation
func (s *userService) getTenantID(ctx context.Context) string {
    if tenantID, ok := ctx.Value("tenant_id").(string); ok {
        return tenantID
    }
    return ""
}
```

**Frontend tenant resolution** (see `frontend/src/utils/tenant.js`):
```javascript
// Auto-resolve tenant from subdomain or email domain
export const resolveTenant = (email = null) => {
  const hostname = window.location.hostname;
  // Subdomain logic + email domain fallback
};
```

### 2. JWT & Authentication Flow
**RSA-signed tokens** with tenant claims - validate at every protected endpoint.

```go
// JWT includes tenant_id claim
type Claims struct {
    UserID   uuid.UUID `json:"user_id"`
    TenantID uuid.UUID `json:"tenant_id"`  // Critical for isolation
    UserType string    `json:"user_type"`
}

// Middleware stack order (see cmd/server/main.go)
user.Use(authMiddleware.AuthRequired())     // JWT validation
user.Use(authMiddleware.TenantRequired())   // Tenant context injection
```

### 3. Database Patterns
**GORM with UUID primary keys** and tenant-scoped queries.

```go
// Models use UUIDs, not auto-incrementing IDs
type User struct {
    ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    TenantID uuid.UUID `gorm:"type:uuid;not null"`
    // ... other fields
}

// Repository constructor pattern
func NewRepository(db *gorm.DB) Repository {
    return &repository{
        userRepo: NewUserRepository(db),  // Each repo gets same DB handle
    }
}
```

### 4. Error Handling & Validation
**Structured error responses** with consistent JSON format.

```go
// Standard API response format
c.JSON(http.StatusOK, gin.H{
    "success": true,
    "data":    responseData,
})

// Error responses
c.JSON(http.StatusBadRequest, gin.H{
    "success": false,
    "message": "validation error details",
})
```

### 5. Testing Strategy
**Isolated test databases** - each test gets its own PostgreSQL database.

```go
// Test setup creates unique database per test
func SetupTestDB(t *testing.T) *TestDB {
    dbName := fmt.Sprintf("auth_test_%s", uuid.New().String()[:8])
    // Creates, migrates, and returns isolated DB connection
}

// Test data seeding (see tests/fixtures/)
tenants := fixtures.TestTenants()  // Predefined test tenants
users := fixtures.TestUsers(tenant.ID)  // Users per tenant
```

## Development Workflow Commands

### Local Development (see `dev.sh`)
```bash
./dev.sh dev      # Start with live reload (foreground)
./dev.sh dev-d    # Start in background
./dev.sh stop     # Stop containers
./dev.sh logs     # View logs
./dev.sh shell    # Open container shell
./dev.sh test     # Run full test suite
./dev.sh clean    # Remove containers + volumes
```

### Testing (see `docker-compose.test.yml`)
```bash
# Run all tests with isolated databases
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# Run specific test types
go test ./tests/unit/... -v     # Unit tests only
go test ./tests/feature/... -v  # Integration tests only
```

### Database Operations
```bash
# Local PostgreSQL setup
createdb auth_db
createuser auth_user --password

# Migrations run automatically on startup
# Seeding happens via internal/seeder/
```

## Security Implementation Details

### Password Security
- **Argon2id hashing** with configurable parameters
- **Validation**: 8+ chars, uppercase, lowercase, number, special char
- **No plaintext storage** - hash immediately on registration

### Session Management
```go
// Redis-backed sessions with TTL
session := &models.UserSession{
    UserID:    user.ID,
    TenantID:  user.TenantID,  // Tenant-scoped sessions
    ExpiresAt: time.Now().Add(24 * time.Hour),
}
```

### CORS Configuration
**Dynamic origin validation** supporting tenant subdomains:
```go
// Supports patterns like "*.sprout.com"
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc
```

## Key Files to Reference

### Architecture Understanding
- `cmd/server/main.go` - Service initialization and middleware stack
- `internal/service/user_service.go` - Business logic patterns
- `internal/repository/repository.go` - Data access patterns
- `internal/middleware/auth.go` - Authentication flow

### Multi-Tenant Implementation
- `internal/service/user_service.go` - Tenant validation logic
- `frontend/src/utils/tenant.js` - Frontend tenant resolution
- `frontend/src/contexts/AuthContext.js` - Tenant-specific storage
- `tests/feature/multi_tenant_test.go` - Tenant isolation testing

### Development Setup
- `dev.sh` - Development workflow script
- `docker-compose.dev.yml` - Development environment
- `tests/testutils/database.go` - Test database isolation

## Common Patterns to Follow

### 1. Context Propagation
Always pass `context.Context` through service/repository layers for tenant isolation and cancellation.

### 2. Tenant Context Injection
Protected routes automatically inject tenant_id from JWT claims or X-Tenant-ID header.

### 3. Repository Interface Pattern
```go
type UserRepository interface {
    GetByEmail(ctx context.Context, email, tenantID string) (*models.User, error)
    // All methods include tenantID parameter
}
```

### 4. Service Layer Validation
Business logic validates tenant ownership before operations:
```go
if user.TenantID.String() != tenantID {
    return errors.New("user does not belong to your tenant")
}
```

### 5. Frontend API Integration
```javascript
// Tenant-aware API calls
const response = await api.get('/user/profile', {
    headers: { 'X-Tenant-ID': tenantId }
});
```

## Environment-Specific Behavior

### Development
- Auto-seeding with test data
- CORS allows localhost origins
- Detailed logging and error messages
- Live reload with CompileDaemon

### Production
- Gin release mode
- Secure CORS origins only
- Structured logging
- CSRF protection enabled

## Testing Guidelines

### Unit Tests
- Test business logic in isolation
- Mock external dependencies (Redis, Email)
- Use table-driven tests for multiple scenarios

### Integration Tests
- Full HTTP request/response testing
- Isolated PostgreSQL databases per test
- Test multi-tenant isolation scenarios
- Verify CORS and security middleware

### Test Data
- Use `tests/fixtures/` for consistent test data
- Each tenant gets predefined users with known credentials
- Passwords are pre-hashed for testing

Remember: **Tenant isolation is the #1 security concern** - every database query, API response, and frontend operation must respect tenant boundaries.
<parameter name="filePath">/Users/niloflora/fligno/blocksure/abc/auth-service/.github/copilot-instructions.md