package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"auth-service/internal/config"
	"auth-service/internal/repository"
)

func main() {
	log.Println("Starting database migration...")

	// Load configuration
	cfg := config.Load()

	// Build database connection string
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Run migrations
	log.Println("Running GORM AutoMigrate...")
	if err := repository.Migrate(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("✅ Database migration completed successfully!")
	log.Println("")
	log.Println("Migrated tables:")
	log.Println("  - users (global, no tenant_id)")
	log.Println("  - organizations")
	log.Println("  - organization_memberships")
	log.Println("  - organization_invitations")
	log.Println("  - user_sessions (organization-scoped)")
	log.Println("  - refresh_tokens (organization-scoped)")
	log.Println("  - password_resets (global)")
	log.Println("  - failed_login_attempts (global)")

	os.Exit(0)
}
