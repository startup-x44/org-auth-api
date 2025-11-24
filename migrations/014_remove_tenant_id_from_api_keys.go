package migrations
package migrations

import (
	"database/sql"
	"log"
)

func init() {
	Migrations = append(Migrations, Migration{
		Version:     14,
		Description: "Remove tenant_id from api_keys table",
		Up:          mig014Up,
		Down:        mig014Down,
	})
}

func mig014Up(tx *sql.Tx) error {
	log.Println("Running migration 014: Remove tenant_id from api_keys")

	// Drop the index first
	_, err := tx.Exec(`
	DROP INDEX IF EXISTS idx_api_keys_tenant_id;
	`)
	if err != nil {
		log.Fatal("Failed to drop index idx_api_keys_tenant_id:", err)
		return err
	}

	// Drop the tenant_id column
	_, err = tx.Exec(`
	ALTER TABLE api_keys DROP COLUMN IF EXISTS tenant_id;
	`)
	if err != nil {
		log.Fatal("Failed to drop tenant_id column:", err)
		return err
	}

	log.Println("Migration 014 completed successfully")
	return nil
}

func mig014Down(tx *sql.Tx) error {
	log.Println("Rolling back migration 014: Add tenant_id back to api_keys")

	// Add tenant_id column back
	_, err := tx.Exec(`
	ALTER TABLE api_keys ADD COLUMN tenant_id UUID;
	`)
	if err != nil {
		log.Fatal("Failed to add tenant_id column:", err)
		return err
	}

	// Recreate the index
	_, err = tx.Exec(`
	CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_id ON api_keys(tenant_id);
	`)
	if err != nil {
		log.Fatal("Failed to create index idx_api_keys_tenant_id:", err)
		return err
	}

	log.Println("Migration 014 rollback completed successfully")
	return nil
}
