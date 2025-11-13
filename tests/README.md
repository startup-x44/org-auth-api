# Auth Service Tests

This directory contains comprehensive unit and feature tests for the authentication service.

## Test Structure

```
tests/
├── fixtures/           # Test data fixtures
├── feature/           # API endpoint tests
├── unit/              # Unit tests for services and repositories
│   ├── repository/    # Repository layer tests
│   └── service/       # Service layer tests
└── testutils/         # Test utilities and database setup
```

## Running Tests

### Using Docker Compose (Recommended)

```bash
# Run all tests
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# Run only unit tests
docker-compose -f docker-compose.test.yml run --rm auth-service-test go test ./tests/unit/... -v

# Run only feature tests
docker-compose -f docker-compose.test.yml run --rm auth-service-test go test ./tests/feature/... -v
```

### Manual Test Execution

If you have Go installed locally:

```bash
# Set test database environment variables
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=auth_user
export TEST_DB_PASSWORD=auth_password
export TEST_DB_SSLMODE=disable

# Run unit tests
go test ./tests/unit/... -v

# Run feature tests
go test ./tests/feature/... -v

# Run all tests
go test ./tests/... -v
```

## Test Coverage

### Unit Tests
- **User Service**: Login, registration, profile management, password operations
- **Repository Layer**: CRUD operations, data validation, error handling
- **Tenant Resolution**: Domain to UUID conversion, validation

### Feature Tests
- **API Endpoints**: Authentication flows, user management, error responses
- **HTTP Handlers**: Request/response validation, status codes
- **Middleware**: Authentication, authorization

## Test Database

Tests use isolated PostgreSQL databases that are created and destroyed for each test run. The test utilities handle:

- Database creation and cleanup
- Schema migration
- Test data seeding
- Connection management

## Test Fixtures

Predefined test data is available in `tests/fixtures/`:

- Sample tenants, users, and sessions
- Valid and invalid test scenarios
- Edge cases and error conditions

## Writing New Tests

### Unit Test Example

```go
func TestUserService_Login(t *testing.T) {
    testDB := testutils.SetupTestDB(t)
    defer testDB.TeardownTestDB(t)

    // Setup test data
    tenant := testutils.CreateTestTenant(t, testDB.DB, "Test Org", "test.local")
    user := testutils.CreateTestUser(t, testDB.DB, "test@example.com", "hash", "Admin", tenant.ID)

    // Test logic here
    assert.NotNil(t, user)
}
```

### Feature Test Example

```go
func TestAuthAPI_Login(t *testing.T) {
    server, cleanup := setupTestServer(t)
    defer cleanup()

    // Make HTTP request
    resp, err := http.Post(server.URL+"/api/v1/auth/login", "application/json", body)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Test Data

The tests include seeded data that matches the production seeders:

- **Tenants**: default.local, demo.company.com, test.org
- **Users**: Admin, Student, and RTO accounts for each tenant
- **Passwords**: Admin123!, Student123!, RTO123!

## Continuous Integration

Tests are designed to run in CI/CD pipelines with:

- Isolated test databases
- No external dependencies
- Fast execution (< 30 seconds)
- Comprehensive coverage

## Troubleshooting

### Database Connection Issues

Ensure PostgreSQL is running and accessible:

```bash
# Check if PostgreSQL is running
docker-compose ps postgres-test

# View database logs
docker-compose logs postgres-test
```

### Test Failures

- Check database connectivity
- Verify environment variables
- Ensure no port conflicts
- Review test logs for detailed error messages

### Performance Issues

- Tests create isolated databases for each run
- Use `t.Parallel()` for concurrent test execution
- Clean up resources in test teardown functions