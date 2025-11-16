-- Drop API keys table and related objects

-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_api_keys_updated_at ON api_keys;
DROP FUNCTION IF EXISTS update_api_keys_updated_at();

-- Drop constraints
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS fk_api_keys_user_id;
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS fk_api_keys_client_app_id;
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS chk_api_keys_name_length;
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS chk_api_keys_key_id_format;

-- Drop indexes
DROP INDEX IF EXISTS idx_api_keys_key_id;
DROP INDEX IF EXISTS idx_api_keys_user_id;
DROP INDEX IF EXISTS idx_api_keys_tenant_id;
DROP INDEX IF EXISTS idx_api_keys_client_app_id;
DROP INDEX IF EXISTS idx_api_keys_active;

-- Drop table
DROP TABLE IF EXISTS api_keys;