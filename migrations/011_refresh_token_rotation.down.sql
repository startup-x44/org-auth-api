-- Remove indexes
DROP INDEX IF EXISTS idx_oauth_refresh_tokens_device_id;
DROP INDEX IF EXISTS idx_oauth_refresh_tokens_ip_hash;
DROP INDEX IF EXISTS idx_oauth_refresh_tokens_user_agent_hash;
DROP INDEX IF EXISTS idx_oauth_refresh_tokens_family_id;
DROP INDEX IF EXISTS idx_oauth_refresh_tokens_replaced_by_id;
DROP INDEX IF EXISTS idx_oauth_refresh_tokens_used_at;
DROP INDEX IF EXISTS idx_authorization_codes_code_hash;

-- Remove refresh token rotation and binding fields
ALTER TABLE oauth_refresh_tokens 
DROP COLUMN IF EXISTS device_id,
DROP COLUMN IF EXISTS ip_hash,
DROP COLUMN IF EXISTS user_agent_hash,
DROP COLUMN IF EXISTS family_id,
DROP COLUMN IF EXISTS replaced_by_id,
DROP COLUMN IF EXISTS used_at;

-- Rename columns back
ALTER TABLE oauth_refresh_tokens RENAME COLUMN token_hash TO token;
ALTER TABLE authorization_codes RENAME COLUMN code_hash TO code;

-- Recreate original unique index
CREATE UNIQUE INDEX IF NOT EXISTS authorization_codes_code_key ON authorization_codes(code);
