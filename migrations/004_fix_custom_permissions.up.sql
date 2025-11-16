-- Fix any existing custom permissions that were incorrectly marked as system permissions
-- This will set is_system = false for permissions that are NOT in the system permissions list

UPDATE permissions SET is_system = false WHERE name NOT IN (
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
) AND is_system = true;