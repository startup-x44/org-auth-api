-- Create authorization_codes table for OAuth2 PKCE flow
CREATE TABLE IF NOT EXISTS authorization_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(255) NOT NULL UNIQUE,
    client_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    organization_id UUID,
    redirect_uri TEXT NOT NULL,
    scope TEXT,
    code_challenge VARCHAR(255),
    code_challenge_method VARCHAR(10),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client_apps(client_id) ON DELETE CASCADE
);

CREATE INDEX idx_authorization_codes_code ON authorization_codes(code);
CREATE INDEX idx_authorization_codes_client_id ON authorization_codes(client_id);
CREATE INDEX idx_authorization_codes_user_id ON authorization_codes(user_id);
CREATE INDEX idx_authorization_codes_expires_at ON authorization_codes(expires_at);
CREATE INDEX idx_authorization_codes_used ON authorization_codes(used);

COMMENT ON TABLE authorization_codes IS 'OAuth2 authorization codes for PKCE flow';
COMMENT ON COLUMN authorization_codes.code_challenge IS 'PKCE code challenge (S256 hash of verifier)';
COMMENT ON COLUMN authorization_codes.code_challenge_method IS 'Must be S256';
COMMENT ON COLUMN authorization_codes.used IS 'Single-use only - must be true after token exchange';
