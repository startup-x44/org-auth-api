package testutils

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
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

	// Connect to postgres database to create test database
	adminDB := connectToAdminDB(t)
	defer func() {
		sqlDB, _ := adminDB.DB()
		sqlDB.Close()
	}()

	// Create test database
	createTestDatabase(t, adminDB, dbName)

	// Connect to test database
	testDB := connectToTestDB(t, dbName)

	// Run migrations
	runMigrations(t, testDB)

	return &TestDB{
		DB:       testDB,
		Database: dbName,
	}
}

// TeardownTestDB cleans up the test database
func (tdb *TestDB) TeardownTestDB(t *testing.T) {
	t.Helper()

	// Close the test database connection
	if sqlDB, err := tdb.DB.DB(); err == nil {
		sqlDB.Close()
	}

	// Connect to admin database to drop test database
	adminDB := connectToAdminDB(t)
	defer func() {
		sqlDB, _ := adminDB.DB()
		sqlDB.Close()
	}()

	dropTestDatabase(t, adminDB, tdb.Database)
}

// connectToAdminDB connects to the postgres admin database
func connectToAdminDB(t *testing.T) *gorm.DB {
	dsn := getTestDSN("postgres")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	return db
}

// connectToTestDB connects to the specific test database
func connectToTestDB(t *testing.T, dbName string) *gorm.DB {
	dsn := getTestDSN(dbName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	return db
}

// createTestDatabase creates a new test database
func createTestDatabase(t *testing.T, db *gorm.DB, dbName string) {
	query := fmt.Sprintf("CREATE DATABASE %s", dbName)
	err := db.Exec(query).Error
	require.NoError(t, err)
}

// dropTestDatabase drops the test database
func dropTestDatabase(t *testing.T, db *gorm.DB, dbName string) {
	// Terminate active connections first
	terminateQuery := fmt.Sprintf(`
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = '%s' AND pid <> pg_backend_pid()`, dbName)
	db.Exec(terminateQuery)

	// Drop the database
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	err := db.Exec(query).Error
	require.NoError(t, err)
}

// runMigrations runs database migrations for testing
func runMigrations(t *testing.T, db *gorm.DB) {
	// Auto-migrate all models
	err := db.AutoMigrate(
		&models.User{},
		&models.Tenant{},
		&models.UserSession{},
		&models.RefreshToken{},
		&models.PasswordReset{},
	)
	require.NoError(t, err)
}

// getTestDSN returns the database connection string for testing
func getTestDSN(dbName string) string {
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5432")
	user := getEnvOrDefault("TEST_DB_USER", "auth_user")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "auth_password")
	sslmode := getEnvOrDefault("TEST_DB_SSLMODE", "disable")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslmode)
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CreateTestTenant creates a test tenant for testing
func CreateTestTenant(t *testing.T, db *gorm.DB, name, domain string) *models.Tenant {
	t.Helper()

	tenant := &models.Tenant{
		Name:   name,
		Domain: domain,
	}

	err := db.Create(tenant).Error
	require.NoError(t, err)

	return tenant
}

// CreateTestUser creates a test user for testing
func CreateTestUser(t *testing.T, db *gorm.DB, email, password, userType string, tenantID uuid.UUID) *models.User {
	t.Helper()

	user := &models.User{
		Email:        email,
		PasswordHash: password,
		UserType:     userType,
		TenantID:     tenantID,
		Status:       models.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}

// CleanTestData cleans all test data from tables
func CleanTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	// Clean in order to respect foreign key constraints
	tables := []string{
		"password_resets",
		"refresh_tokens",
		"user_sessions",
		"users",
		"tenants",
	}

	for _, table := range tables {
		err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error
		require.NoError(t, err)
	}
}