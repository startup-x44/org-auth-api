-- Create client_apps table
CREATE TABLE IF NOT EXISTS client_apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    client_id VARCHAR(255) NOT NULL UNIQUE,
    client_secret VARCHAR(255) NOT NULL,
    redirect_uris TEXT[],
    allowed_origins TEXT[],
    allowed_scopes TEXT[],
    is_confidential BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_client_apps_client_id ON client_apps(client_id);

COMMENT ON TABLE client_apps IS 'OAuth2 client applications - only manageable by superadmin';
COMMENT ON COLUMN client_apps.client_secret IS 'bcrypt hashed client secret';
COMMENT ON COLUMN client_apps.is_confidential IS 'true = requires secret, false = public client (PKCE only)';
