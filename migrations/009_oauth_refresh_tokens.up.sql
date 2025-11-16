-- Create oauth_refresh_tokens table
CREATE TABLE IF NOT EXISTS oauth_refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token VARCHAR(255) NOT NULL UNIQUE,
    client_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    organization_id UUID,
    scope TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client_apps(client_id) ON DELETE CASCADE
);

CREATE INDEX idx_oauth_refresh_tokens_token ON oauth_refresh_tokens(token);
CREATE INDEX idx_oauth_refresh_tokens_client_id ON oauth_refresh_tokens(client_id);
CREATE INDEX idx_oauth_refresh_tokens_user_id ON oauth_refresh_tokens(user_id);
CREATE INDEX idx_oauth_refresh_tokens_expires_at ON oauth_refresh_tokens(expires_at);
CREATE INDEX idx_oauth_refresh_tokens_revoked ON oauth_refresh_tokens(revoked);

COMMENT ON TABLE oauth_refresh_tokens IS 'OAuth2 refresh tokens - hashed in database';
COMMENT ON COLUMN oauth_refresh_tokens.token IS 'bcrypt hashed refresh token';
COMMENT ON COLUMN oauth_refresh_tokens.revoked IS 'Set to true on logout or security event';
