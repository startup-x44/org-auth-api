//go:build ignore

package main

import (
	"context"
	"database/sql"
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

	// Add organization_id to client_apps table
	alterClientAppsSQL := `
	ALTER TABLE client_apps
	ADD COLUMN IF NOT EXISTS organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
	`

	if _, err := db.ExecContext(ctx, alterClientAppsSQL); err != nil {
		log.Fatal("Failed to add organization_id to client_apps:", err)
	}

	// Create index for client_apps
	createClientAppsIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_client_apps_organization_id ON client_apps(organization_id);
	`

	if _, err := db.ExecContext(ctx, createClientAppsIndexSQL); err != nil {
		log.Fatal("Failed to create index on client_apps:", err)
	}

	// Add organization_id to api_keys table
	alterAPIKeysSQL := `
	ALTER TABLE api_keys
	ADD COLUMN IF NOT EXISTS organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
	`

	if _, err := db.ExecContext(ctx, alterAPIKeysSQL); err != nil {
		log.Fatal("Failed to add organization_id to api_keys:", err)
	}

	// Create index for api_keys
	createAPIKeysIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_api_keys_organization_id ON api_keys(organization_id);
	`

	if _, err := db.ExecContext(ctx, createAPIKeysIndexSQL); err != nil {
		log.Fatal("Failed to create index on api_keys:", err)
	}

	log.Println("Migration 013_add_organization_id_to_client_apps_and_api_keys completed successfully")
}
