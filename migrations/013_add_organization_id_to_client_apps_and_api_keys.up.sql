-- Add organization_id to client_apps table
ALTER TABLE client_apps
ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- Create index for faster queries
CREATE INDEX idx_client_apps_organization_id ON client_apps(organization_id);

-- Add organization_id to api_keys table
ALTER TABLE api_keys
ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- Create index for faster queries
CREATE INDEX idx_api_keys_organization_id ON api_keys(organization_id);
