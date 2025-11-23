-- Add tenant_id back to api_keys table (rollback)
ALTER TABLE api_keys ADD COLUMN tenant_id UUID;
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_id ON api_keys(tenant_id);
