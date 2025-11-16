package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// This migration adds the API keys table for developer API access
	// Note: In production, replace with actual connection string from environment

	db, err := sql.Open("postgres", "your_connection_string_here")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Check if migration has already been applied
	var exists bool
	err = db.QueryRowContext(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'api_keys')").Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}

	if exists {
		fmt.Println("API keys table already exists, skipping migration")
		return
	}

	fmt.Println("Creating API keys table...")

	query := `
-- Create API keys table for developer API access
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id VARCHAR(255) UNIQUE NOT NULL,
    hashed_secret TEXT NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    client_app_id UUID,
    user_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    scopes TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    revoked BOOLEAN DEFAULT FALSE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_api_keys_key_id ON api_keys(key_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_client_app_id ON api_keys(client_app_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(revoked, expires_at) WHERE revoked = FALSE;

-- Add foreign key constraints
ALTER TABLE api_keys 
ADD CONSTRAINT fk_api_keys_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE api_keys 
ADD CONSTRAINT fk_api_keys_client_app_id 
FOREIGN KEY (client_app_id) REFERENCES client_apps(id) ON DELETE SET NULL;

-- Add check constraints
ALTER TABLE api_keys 
ADD CONSTRAINT chk_api_keys_name_length 
CHECK (length(name) >= 1 AND length(name) <= 100);

ALTER TABLE api_keys 
ADD CONSTRAINT chk_api_keys_key_id_format 
CHECK (key_id ~ '^ak_[a-zA-Z0-9]{32}$');

-- Add updated_at trigger
CREATE OR REPLACE FUNCTION update_api_keys_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_api_keys_updated_at
    BEFORE UPDATE ON api_keys
    FOR EACH ROW
    EXECUTE FUNCTION update_api_keys_updated_at();
`

	_, err = db.ExecContext(ctx, query)
	if err != nil {
		log.Fatalf("Failed to create api_keys table: %v", err)
	}

	fmt.Println("✅ API keys table created successfully")
}
