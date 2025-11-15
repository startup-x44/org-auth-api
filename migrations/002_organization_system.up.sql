-- +migrate Up
-- Convert from tenant-based to organization-based system

-- Step 1: Clean up old sessions (they'll need to re-login with organization selection)
TRUNCATE TABLE user_sessions CASCADE;
TRUNCATE TABLE refresh_tokens CASCADE;

-- Step 2: Make tenant_id nullable in users table (for backward compatibility during migration)
ALTER TABLE users ALTER COLUMN tenant_id DROP NOT NULL;

-- Step 3: Add new columns to users table for organization system
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_superadmin BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS global_role VARCHAR(50) DEFAULT 'user';
ALTER TABLE users ADD COLUMN IF NOT EXISTS legacy_tenant_id UUID;
ALTER TABLE users ADD COLUMN IF NOT EXISTS legacy_user_type VARCHAR(50);

-- Step 4: Create organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    settings JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Step 5: Create organization_memberships table
CREATE TABLE IF NOT EXISTS organization_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMP WITH TIME ZONE,
    joined_at TIMESTAMP WITH TIME ZONE,
    last_activity_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, user_id)
);

-- Step 6: Create organization_invitations table
CREATE TABLE IF NOT EXISTS organization_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    role VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    invited_by UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Step 7: Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_superadmin ON users(is_superadmin);

CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug);
CREATE INDEX IF NOT EXISTS idx_organizations_status ON organizations(status);
CREATE INDEX IF NOT EXISTS idx_organizations_created_by ON organizations(created_by);

CREATE INDEX IF NOT EXISTS idx_membership_org_user ON organization_memberships(organization_id, user_id);
CREATE INDEX IF NOT EXISTS idx_membership_user ON organization_memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_membership_status ON organization_memberships(status);

CREATE INDEX IF NOT EXISTS idx_invitation_email_org ON organization_invitations(email, organization_id);
CREATE INDEX IF NOT EXISTS idx_invitation_token ON organization_invitations(token_hash);
CREATE INDEX IF NOT EXISTS idx_invitation_status ON organization_invitations(status);
CREATE INDEX IF NOT EXISTS idx_invitation_expires ON organization_invitations(expires_at);

-- Step 8: Update existing user records to have is_superadmin = true for Admin user_type
UPDATE users SET is_superadmin = true WHERE tenant_id IS NOT NULL AND tenant_id::text != '';

-- Step 9: Create updated_at triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_organization_memberships_updated_at ON organization_memberships;
CREATE TRIGGER update_organization_memberships_updated_at BEFORE UPDATE ON organization_memberships FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_organization_invitations_updated_at ON organization_invitations;
CREATE TRIGGER update_organization_invitations_updated_at BEFORE UPDATE ON organization_invitations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +migrate Down
-- Revert organization-based system back to tenant-based

-- Remove triggers
DROP TRIGGER IF EXISTS update_organization_invitations_updated_at ON organization_invitations;
DROP TRIGGER IF EXISTS update_organization_memberships_updated_at ON organization_memberships;
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Remove new tables
DROP TABLE IF EXISTS organization_invitations;
DROP TABLE IF EXISTS organization_memberships;
DROP TABLE IF EXISTS organizations;

-- Remove new columns from users
ALTER TABLE users DROP COLUMN IF EXISTS is_superadmin;
ALTER TABLE users DROP COLUMN IF EXISTS global_role;
ALTER TABLE users DROP COLUMN IF EXISTS legacy_tenant_id;
ALTER TABLE users DROP COLUMN IF EXISTS legacy_user_type;

-- Make tenant_id NOT NULL again
ALTER TABLE users ALTER COLUMN tenant_id SET NOT NULL;