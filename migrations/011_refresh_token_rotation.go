//go:build ignore
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get connection string from environment
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Rename columns for clarity
	renameSQL := `
	ALTER TABLE oauth_refresh_tokens RENAME COLUMN token TO token_hash;
	ALTER TABLE authorization_codes RENAME COLUMN code TO code_hash;
	`

	if _, err := db.ExecContext(ctx, renameSQL); err != nil {
		log.Fatal("Failed to rename columns:", err)
	}

	// Add refresh token rotation and binding fields
	alterTableSQL := `
	ALTER TABLE oauth_refresh_tokens 
	ADD COLUMN IF NOT EXISTS used_at TIMESTAMP,
	ADD COLUMN IF NOT EXISTS replaced_by_id UUID,
	ADD COLUMN IF NOT EXISTS family_id UUID,
	ADD COLUMN IF NOT EXISTS user_agent_hash VARCHAR(64),
	ADD COLUMN IF NOT EXISTS ip_hash VARCHAR(64),
	ADD COLUMN IF NOT EXISTS device_id VARCHAR(255);
	`

	if _, err := db.ExecContext(ctx, alterTableSQL); err != nil {
		log.Fatal("Failed to alter oauth_refresh_tokens table:", err)
	}

	// Backfill family_id for existing tokens
	backfillSQL := `
	UPDATE oauth_refresh_tokens SET family_id = id WHERE family_id IS NULL;
	ALTER TABLE oauth_refresh_tokens ALTER COLUMN family_id SET NOT NULL;
	`

	if _, err := db.ExecContext(ctx, backfillSQL); err != nil {
		log.Fatal("Failed to backfill family_id:", err)
	}

	// Add indexes for performance
	createIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_used_at ON oauth_refresh_tokens(used_at);
	CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_replaced_by_id ON oauth_refresh_tokens(replaced_by_id);
	CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_family_id ON oauth_refresh_tokens(family_id);
	CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_user_agent_hash ON oauth_refresh_tokens(user_agent_hash);
	CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_ip_hash ON oauth_refresh_tokens(ip_hash);
	CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_device_id ON oauth_refresh_tokens(device_id);
	
	DROP INDEX IF EXISTS authorization_codes_code_key;
	CREATE UNIQUE INDEX IF NOT EXISTS idx_authorization_codes_code_hash ON authorization_codes(code_hash);
	`

	if _, err := db.ExecContext(ctx, createIndexSQL); err != nil {
		log.Fatal("Failed to create indexes:", err)
	}

	fmt.Println("Migration 011: Refresh token rotation with deterministic hashing - completed successfully")
}
