-- Migration: Separate system roles from organization roles
-- System roles: IsSystem=true, organization_id=NULL (global, superadmin only)
-- Custom roles: IsSystem=false, organization_id=required (org-specific)

-- Make organization_id nullable for roles to allow system roles
ALTER TABLE roles ALTER COLUMN organization_id DROP NOT NULL;

-- Add index for system role queries
CREATE INDEX IF NOT EXISTS idx_role_system ON roles(is_system);
CREATE INDEX IF NOT EXISTS idx_role_name_org ON roles(name, organization_id);

-- Add check constraint: system roles must have NULL organization_id
ALTER TABLE roles ADD CONSTRAINT chk_role_system_org 
  CHECK ((is_system = true AND organization_id IS NULL) OR (is_system = false AND organization_id IS NOT NULL));

-- Add comment for clarity
COMMENT ON COLUMN roles.is_system IS 'true for global system roles (superadmin only), false for custom organization roles';
COMMENT ON COLUMN roles.organization_id IS 'NULL for system roles, required for custom roles';

-- Make organization_id nullable for permissions to allow system permissions
-- This was already nullable, but let's ensure it
ALTER TABLE permissions ALTER COLUMN organization_id DROP NOT NULL;

-- Add comment for permissions
COMMENT ON COLUMN permissions.is_system IS 'true for global system permissions (superadmin only), false for custom organization permissions';
COMMENT ON COLUMN permissions.organization_id IS 'NULL for system permissions, set for custom organization permissions';
