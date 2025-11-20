-- Rename refresh token column from 'token' to 'token_hash' for clarity
ALTER TABLE oauth_refresh_tokens RENAME COLUMN token TO token_hash;

-- Rename authorization code column from 'code' to 'code_hash' for clarity  
ALTER TABLE authorization_codes RENAME COLUMN code TO code_hash;

-- Add refresh token rotation and binding fields
ALTER TABLE oauth_refresh_tokens 
ADD COLUMN IF NOT EXISTS used_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS replaced_by_id UUID,
ADD COLUMN IF NOT EXISTS family_id UUID,
ADD COLUMN IF NOT EXISTS user_agent_hash VARCHAR(64),
ADD COLUMN IF NOT EXISTS ip_hash VARCHAR(64),
ADD COLUMN IF NOT EXISTS device_id VARCHAR(255);

-- Update existing tokens to have a family_id (each existing token is its own family)
UPDATE oauth_refresh_tokens SET family_id = id WHERE family_id IS NULL;

-- Make family_id NOT NULL after backfilling
ALTER TABLE oauth_refresh_tokens ALTER COLUMN family_id SET NOT NULL;

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_used_at ON oauth_refresh_tokens(used_at);
CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_replaced_by_id ON oauth_refresh_tokens(replaced_by_id);
CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_family_id ON oauth_refresh_tokens(family_id);
CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_user_agent_hash ON oauth_refresh_tokens(user_agent_hash);
CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_ip_hash ON oauth_refresh_tokens(ip_hash);
CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_device_id ON oauth_refresh_tokens(device_id);

-- Update unique index on authorization_codes to use code_hash
DROP INDEX IF EXISTS authorization_codes_code_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_authorization_codes_code_hash ON authorization_codes(code_hash);

-- Add comments for documentation
COMMENT ON COLUMN oauth_refresh_tokens.token_hash IS 'HMAC-SHA256 hash of refresh token for deterministic lookup';
COMMENT ON COLUMN oauth_refresh_tokens.used_at IS 'Timestamp when token was rotated/used';
COMMENT ON COLUMN oauth_refresh_tokens.replaced_by_id IS 'ID of the new token that replaced this one';
COMMENT ON COLUMN oauth_refresh_tokens.family_id IS 'Groups all tokens in the same rotation chain';
COMMENT ON COLUMN oauth_refresh_tokens.user_agent_hash IS 'SHA256 hash of user agent for token binding';
COMMENT ON COLUMN oauth_refresh_tokens.ip_hash IS 'SHA256 hash of IP address for token binding';
COMMENT ON COLUMN authorization_codes.code_hash IS 'HMAC-SHA256 hash of authorization code for deterministic lookup';
