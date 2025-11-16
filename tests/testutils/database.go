package testutils

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"auth-service/internal/models"
)

// TestDB holds test database connection and utilities
type TestDB struct {
	DB       *gorm.DB
	SQLDB    *sql.DB
	Database string
}

// SetupTestDB creates a new test database and returns a connection
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	// Generate unique database name for this test
	dbName := fmt.Sprintf("auth_test_%s", uuid.New().String()[:8])

	// Get database connection parameters from environment or use defaults
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "password")

	// Connect to postgres database to create test database
	postgresDS := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable TimeZone=UTC",
		host, port, user, password)

	sqlDB, err := sql.Open("postgres", postgresDS)
	if err != nil {
		t.Skipf("Cannot connect to PostgreSQL for testing: %v", err)
		return nil
	}
	defer sqlDB.Close()

	// Create test database
	_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		t.Skipf("Cannot create test database: %v", err)
		return nil
	}

	// Connect to the test database
	testDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		host, port, user, password, dbName)

	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Suppress logs during tests
	})
	require.NoError(t, err)

	testSQLDB, err := db.DB()
	require.NoError(t, err)

	// Run migrations
	err = runMigrations(db)
	require.NoError(t, err)

	testDB := &TestDB{
		DB:       db,
		SQLDB:    testSQLDB,
		Database: dbName,
	}

	// Cleanup function to drop database after test
	t.Cleanup(func() {
		testDB.Cleanup(t)
	})

	return testDB
}

// Cleanup drops the test database
func (tdb *TestDB) Cleanup(t *testing.T) {
	t.Helper()

	if tdb.SQLDB != nil {
		tdb.SQLDB.Close()
	}

	// Get database connection parameters
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "password")

	// Connect to postgres database to drop test database
	postgresDS := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable TimeZone=UTC",
		host, port, user, password)

	sqlDB, err := sql.Open("postgres", postgresDS)
	if err != nil {
		t.Logf("Cannot connect to PostgreSQL for cleanup: %v", err)
		return
	}
	defer sqlDB.Close()

	// Drop test database
	_, err = sqlDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", tdb.Database))
	if err != nil {
		t.Logf("Cannot drop test database %s: %v", tdb.Database, err)
	}
}

// runMigrations runs the database migrations for testing
func runMigrations(db *gorm.DB) error {
	// Auto-migrate all models
	return db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationMembership{},
		&models.OrganizationInvitation{},
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.UserSession{},
		&models.RefreshToken{},
		&models.PasswordReset{},
		&models.FailedLoginAttempt{},
	)
}

// CreateTestUser creates a test user for testing
func CreateTestUser(t *testing.T, db *gorm.DB, email string) *models.User {
	t.Helper()

	firstName := "Test"
	lastName := "User"
	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=2$somehashedpassword", // Mock password hash
		Firstname:    &firstName,
		Lastname:     &lastName,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}

// CreateTestOrganization creates a test organization for testing
func CreateTestOrganization(t *testing.T, db *gorm.DB, name, slug string) *models.Organization {
	t.Helper()

	org := &models.Organization{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		Status:    "active",
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := db.Create(org).Error
	require.NoError(t, err)

	return org
}

// CreateTestRole creates a test role for testing
func CreateTestRole(t *testing.T, db *gorm.DB, orgID uuid.UUID, name string) *models.Role {
	t.Helper()

	role := &models.Role{
		ID:             uuid.New(),
		OrganizationID: &orgID,
		Name:           name,
		DisplayName:    name,
		IsSystem:       false,
		CreatedBy:      uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := db.Create(role).Error
	require.NoError(t, err)

	return role
}

// CreateTestPermission creates a test permission for testing
func CreateTestPermission(t *testing.T, db *gorm.DB, name string, orgID *uuid.UUID, isSystem bool) *models.Permission {
	t.Helper()

	perm := &models.Permission{
		ID:             uuid.New(),
		Name:           name,
		DisplayName:    name,
		Category:       "test",
		IsSystem:       isSystem,
		OrganizationID: orgID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := db.Create(perm).Error
	require.NoError(t, err)

	return perm
}

// CleanTestData cleans all test data from tables
func CleanTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	// Clean in order to respect foreign key constraints
	tables := []string{
		"role_permissions",
		"organization_invitations",
		"organization_memberships",
		"failed_login_attempts",
		"password_resets",
		"refresh_tokens",
		"user_sessions",
		"permissions",
		"roles",
		"users",
		"organizations",
	}

	for _, table := range tables {
		err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error
		require.NoError(t, err)
	}
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
