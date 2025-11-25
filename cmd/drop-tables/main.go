package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"auth-service/internal/config"
)

func main() {
	log.Println("⚠️  WARNING: This will drop ALL tables from the database!")
	log.Println("Connecting to database...")

	// Load configuration
	cfg := config.Load()

	// Build database connection string with pooler support
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s prefer_simple_protocol=true",
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

	log.Println("Connected successfully")
	log.Println("Dropping tables...")

	// Drop tables in reverse dependency order
	tables := []string{
		"audit_logs",
		"role_permissions",
		"api_keys",
		"oauth_refresh_tokens",
		"oauth_authorization_codes",
		"authorization_codes",
		"client_apps",
		"organization_invitations",
		"organization_memberships",
		"user_sessions",
		"refresh_tokens",
		"password_resets",
		"failed_login_attempts",
		"permissions",
		"roles",
		"organizations",
		"users",
	}

	for _, table := range tables {
		sql := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		} else {
			log.Printf("✅ Dropped table: %s", table)
		}
	}

	log.Println("")
	log.Println("🎉 All tables dropped successfully!")
	log.Println("")
	log.Println("Next step: Run 'go run ./cmd/migrate' to recreate tables")
}
