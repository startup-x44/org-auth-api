-- Rollback: Mark all non-system permissions as system permissions (revert the fix)
UPDATE permissions SET is_system = true WHERE name NOT IN (
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