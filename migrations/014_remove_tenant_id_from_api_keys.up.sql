-- Remove tenant_id from api_keys table
DROP INDEX IF EXISTS idx_api_keys_tenant_id;
ALTER TABLE api_keys DROP COLUMN IF EXISTS tenant_id;
