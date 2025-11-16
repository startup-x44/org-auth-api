package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// This migration separates system roles from organization roles
	// System roles: IsSystem=true, organization_id=NULL (global, superadmin only)
	// Custom roles: IsSystem=false, organization_id=required (org-specific)

	db, err := sql.Open("postgres", "your_connection_string_here")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := Migration006RBACSystemSeparation(db); err != nil {
		log.Fatal(err)
	}

	log.Println("Migration completed successfully")
}

// Migration006RBACSystemSeparation separates system roles from organization roles
func Migration006RBACSystemSeparation(db *sql.DB) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	sql := `
-- Migration: Separate system roles from organization roles
-- System roles: IsSystem=true, organization_id=NULL (global, superadmin only)
-- Custom roles: IsSystem=false, organization_id=required (org-specific)

-- Make organization_id nullable for roles to allow system roles
ALTER TABLE roles ALTER COLUMN organization_id DROP NOT NULL;

-- Add index for system role queries
CREATE INDEX IF NOT EXISTS idx_role_system ON roles(is_system);
CREATE INDEX IF NOT EXISTS idx_role_name_org ON roles(name, organization_id);

-- Add check constraint: system roles must have NULL organization_id
ALTER TABLE roles ADD CONSTRAINT chk_role_system_org 
  CHECK ((is_system = true AND organization_id IS NULL) OR (is_system = false AND organization_id IS NOT NULL));

-- Add comment for clarity
COMMENT ON COLUMN roles.is_system IS 'true for global system roles (superadmin only), false for custom organization roles';
COMMENT ON COLUMN roles.organization_id IS 'NULL for system roles, required for custom roles';

-- Add comment for permissions
COMMENT ON COLUMN permissions.is_system IS 'true for global system permissions (superadmin only), false for custom organization permissions';
COMMENT ON COLUMN permissions.organization_id IS 'NULL for system permissions, set for custom organization permissions';
`

	if _, err := tx.Exec(sql); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Rollback006RBACSystemSeparation rolls back the RBAC system separation
func Rollback006RBACSystemSeparation(db *sql.DB) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	sql := `
-- Rollback: Revert system role separation

-- Drop check constraint
ALTER TABLE roles DROP CONSTRAINT IF EXISTS chk_role_system_org;

-- Drop indexes
DROP INDEX IF EXISTS idx_role_system;
DROP INDEX IF EXISTS idx_role_name_org;

-- Make organization_id required again for roles
ALTER TABLE roles ALTER COLUMN organization_id SET NOT NULL;
`

	if _, err := tx.Exec(sql); err != nil {
		return fmt.Errorf("failed to rollback migration 006: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	return nil
}
