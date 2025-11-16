-- Rollback: Revert system role separation

-- Drop check constraint
ALTER TABLE roles DROP CONSTRAINT IF EXISTS chk_role_system_org;

-- Drop indexes
DROP INDEX IF EXISTS idx_role_system;
DROP INDEX IF EXISTS idx_role_name_org;

-- Make organization_id required again for roles
ALTER TABLE roles ALTER COLUMN organization_id SET NOT NULL;

-- Remove comments
COMMENT ON COLUMN roles.is_system IS NULL;
COMMENT ON COLUMN roles.organization_id IS NULL;
COMMENT ON COLUMN permissions.is_system IS NULL;
COMMENT ON COLUMN permissions.organization_id IS NULL;
