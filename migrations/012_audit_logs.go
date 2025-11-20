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

	// Create audit_logs table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS audit_logs (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		user_id UUID REFERENCES users(id) ON DELETE SET NULL,
		organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
		action VARCHAR(100) NOT NULL,
		resource VARCHAR(100) NOT NULL,
		resource_id UUID,
		ip_address VARCHAR(45),
		user_agent TEXT,
		request_id VARCHAR(100),
		details JSONB,
		success BOOLEAN NOT NULL DEFAULT true,
		error TEXT,
		service VARCHAR(50) NOT NULL DEFAULT 'auth-service',
		method VARCHAR(200),
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`

	if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
		log.Fatal("Failed to create audit_logs table:", err)
	}

	// Create indexes
	indexesSQL := []string{
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_organization_id ON audit_logs(organization_id);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_request_id ON audit_logs(request_id);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_user_action ON audit_logs(user_id, action, timestamp DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_org_action ON audit_logs(organization_id, action, timestamp DESC);`,
	}

	for _, sql := range indexesSQL {
		if _, err := db.ExecContext(ctx, sql); err != nil {
			log.Fatal("Failed to create index:", err)
		}
	}

	// Add table and column comments
	commentsSQL := []string{
		`COMMENT ON TABLE audit_logs IS 'Stores audit trail for all critical auth operations including login, logout, permission changes, role assignments';`,
		`COMMENT ON COLUMN audit_logs.action IS 'Action performed: login, logout, register, role_assign, permission_grant, etc.';`,
		`COMMENT ON COLUMN audit_logs.resource IS 'Resource type affected: user, role, permission, organization, session';`,
		`COMMENT ON COLUMN audit_logs.details IS 'Additional context as JSON: changed fields, target user, previous/new values';`,
	}

	for _, sql := range commentsSQL {
		if _, err := db.ExecContext(ctx, sql); err != nil {
			log.Printf("Warning: Failed to add comment: %v", err)
		}
	}

	log.Println("Migration 012_audit_logs completed successfully")
}
