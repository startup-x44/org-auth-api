-- Remove organization_id from api_keys table
DROP INDEX IF EXISTS idx_api_keys_organization_id;
ALTER TABLE api_keys DROP COLUMN IF EXISTS organization_id;

-- Remove organization_id from client_apps table
DROP INDEX IF EXISTS idx_client_apps_organization_id;
ALTER TABLE client_apps DROP COLUMN IF EXISTS organization_id;
