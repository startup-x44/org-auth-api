-- Add organization_id column to permissions table for custom permission isolation
-- System permissions will have NULL organization_id, custom permissions will have specific organization_id

-- Add organization_id column (nullable to allow system permissions)
ALTER TABLE permissions 
ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- Create composite unique constraint to prevent duplicate permission names within same organization context
-- This allows same permission name to exist as system permission (NULL org) and in different organizations
ALTER TABLE permissions 
ADD CONSTRAINT permissions_name_organization_unique 
UNIQUE (name, organization_id);

-- Create index for performance on organization-scoped queries
CREATE INDEX idx_permissions_organization_id ON permissions(organization_id);
CREATE INDEX idx_permissions_name_organization ON permissions(name, organization_id);

-- Add comment to clarify the purpose
COMMENT ON COLUMN permissions.organization_id IS 'Organization that owns this custom permission. NULL for system permissions available to all organizations.';