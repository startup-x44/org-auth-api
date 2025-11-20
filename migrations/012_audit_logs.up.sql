-- Create audit_logs table for persisting audit events
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id UUID,
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_id VARCHAR(100),
    details JSONB,
    success BOOLEAN NOT NULL DEFAULT true,
    error TEXT,
    service VARCHAR(50) NOT NULL DEFAULT 'auth-service',
    method VARCHAR(200),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for efficient querying
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_organization_id ON audit_logs(organization_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_request_id ON audit_logs(request_id);

-- Composite index for common queries
CREATE INDEX idx_audit_logs_user_action ON audit_logs(user_id, action, timestamp DESC);
CREATE INDEX idx_audit_logs_org_action ON audit_logs(organization_id, action, timestamp DESC);

-- Add comment for documentation
COMMENT ON TABLE audit_logs IS 'Stores audit trail for all critical auth operations including login, logout, permission changes, role assignments';
COMMENT ON COLUMN audit_logs.action IS 'Action performed: login, logout, register, role_assign, permission_grant, etc.';
COMMENT ON COLUMN audit_logs.resource IS 'Resource type affected: user, role, permission, organization, session';
COMMENT ON COLUMN audit_logs.details IS 'Additional context as JSON: changed fields, target user, previous/new values';
