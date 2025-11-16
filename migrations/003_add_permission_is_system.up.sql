-- Add is_system column to permissions table with default false for new permissions
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS is_system BOOLEAN DEFAULT false;

-- Mark specific system permissions as is_system = true
UPDATE permissions SET is_system = true WHERE name IN (
    -- Organization permissions
    'org:update', 'org:delete', 'org:view',
    -- Member permissions  
    'member:invite', 'member:remove', 'member:update', 'member:view',
    -- Invitation permissions
    'invitation:view', 'invitation:resend', 'invitation:cancel',
    -- Role permissions
    'role:create', 'role:update', 'role:delete', 'role:view',
    -- Certificate permissions
    'cert:issue', 'cert:revoke', 'cert:verify', 'cert:view'
);
