-- Rollback permission organization isolation changes

-- Remove constraints and indexes
DROP INDEX IF EXISTS idx_permissions_name_organization;
DROP INDEX IF EXISTS idx_permissions_organization_id;
ALTER TABLE permissions DROP CONSTRAINT IF EXISTS permissions_name_organization_unique;

-- Remove organization_id column
ALTER TABLE permissions DROP COLUMN IF EXISTS organization_id;