//go:build ignore

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// This is a simple migration script
	// In a real application, you would use a proper migration tool like golang-migrate

	db, err := sql.Open("postgres", "your_connection_string_here")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Add new columns to user_sessions table
	alterTableSQL := `
	ALTER TABLE user_sessions
	ADD COLUMN IF NOT EXISTS device_fingerprint VARCHAR(255),
	ADD COLUMN IF NOT EXISTS location VARCHAR(255),
	ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true,
	ADD COLUMN IF NOT EXISTS last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	ADD COLUMN IF NOT EXISTS revoked_reason TEXT,
	ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
	`

	if _, err := db.ExecContext(ctx, alterTableSQL); err != nil {
		log.Fatal("Failed to alter user_sessions table:", err)
	}

	// Create index on last_activity for performance
	createIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_sessions_activity ON user_sessions(last_activity);
	`

	if _, err := db.ExecContext(ctx, createIndexSQL); err != nil {
		log.Fatal("Failed to create index:", err)
	}

	// Update existing records to be active
	updateSQL := `
	UPDATE user_sessions
	SET is_active = true, last_activity = created_at, updated_at = NOW()
	WHERE is_active IS NULL;
	`

	if _, err := db.ExecContext(ctx, updateSQL); err != nil {
		log.Fatal("Failed to update existing records:", err)
	}

	fmt.Println("Migration completed successfully")
}
