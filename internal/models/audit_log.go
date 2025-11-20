package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit trail entry for critical operations
type AuditLog struct {
	ID             uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Timestamp      time.Time              `gorm:"type:timestamptz;not null;default:NOW()" json:"timestamp"`
	UserID         *uuid.UUID             `gorm:"type:uuid;index" json:"user_id,omitempty"`
	OrganizationID *uuid.UUID             `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	Action         string                 `gorm:"type:varchar(100);not null;index" json:"action"`
	Resource       string                 `gorm:"type:varchar(100);not null;index" json:"resource"`
	ResourceID     *uuid.UUID             `gorm:"type:uuid" json:"resource_id,omitempty"`
	IPAddress      string                 `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent      string                 `gorm:"type:text" json:"user_agent,omitempty"`
	RequestID      string                 `gorm:"type:varchar(100);index" json:"request_id,omitempty"`
	Details        map[string]interface{} `gorm:"type:jsonb" json:"details,omitempty"`
	Success        bool                   `gorm:"not null;default:true" json:"success"`
	Error          string                 `gorm:"type:text" json:"error,omitempty"`
	Service        string                 `gorm:"type:varchar(50);not null;default:'auth-service'" json:"service"`
	Method         string                 `gorm:"type:varchar(200)" json:"method,omitempty"`
	CreatedAt      time.Time              `gorm:"type:timestamptz;not null;default:NOW()" json:"created_at"`
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}

// Common audit actions
const (
	// Authentication actions
	ActionLogin             = "login"
	ActionLoginFailed       = "login_failed"
	ActionLogout            = "logout"
	ActionRegister          = "register"
	ActionTokenRefresh      = "token_refresh"
	ActionPasswordChange    = "password_change"
	ActionPasswordReset     = "password_reset"
	ActionEmailVerification = "email_verification"
	ActionMFAEnable         = "mfa_enable"
	ActionMFADisable        = "mfa_disable"

	// Authorization actions
	ActionRoleAssign          = "role_assign"
	ActionRoleRevoke          = "role_revoke"
	ActionRoleCreate          = "role_create"
	ActionRoleUpdate          = "role_update"
	ActionRoleDelete          = "role_delete"
	ActionRoleView            = "role_view"
	ActionPermissionGrant     = "permission_grant"
	ActionPermissionRevoke    = "permission_revoke"
	ActionPermissionCreate    = "permission_create"
	ActionPermissionUpdate    = "permission_update"
	ActionPermissionDelete    = "permission_delete"
	ActionPermissionView      = "permission_view"
	ActionAuthorizationFailed = "authorization_failed"

	// Organization actions
	ActionOrgCreate    = "org_create"
	ActionOrgUpdate    = "org_update"
	ActionOrgDelete    = "org_delete"
	ActionOrgJoin      = "org_join"
	ActionOrgLeave     = "org_leave"
	ActionMemberInvite = "member_invite"
	ActionMemberRemove = "member_remove"
	ActionMemberUpdate = "member_update"

	// User management actions
	ActionUserCreate     = "user_create"
	ActionUserUpdate     = "user_update"
	ActionUserDelete     = "user_delete"
	ActionUserActivate   = "user_activate"
	ActionUserDeactivate = "user_deactivate"

	// Session actions
	ActionSessionCreate    = "session_create"
	ActionSessionRevoke    = "session_revoke"
	ActionSessionRevokeAll = "session_revoke_all"

	// OAuth2 actions
	ActionOAuthAuthorize     = "oauth_authorize"
	ActionOAuthTokenGrant    = "oauth_token_grant"
	ActionOAuthTokenRevoke   = "oauth_token_revoke"
	ActionClientCreate       = "client_create"
	ActionClientUpdate       = "client_update"
	ActionClientDelete       = "client_delete"
	ActionClientSecretRotate = "client_secret_rotate"

	// API Key actions
	ActionAPIKeyCreate = "api_key_create"
	ActionAPIKeyRevoke = "api_key_revoke"
	ActionAPIKeyDelete = "api_key_delete"
)

// Common resource types
const (
	ResourceUser         = "user"
	ResourceRole         = "role"
	ResourcePermission   = "permission"
	ResourceOrganization = "organization"
	ResourceMember       = "member"
	ResourceSession      = "session"
	ResourceToken        = "token"
	ResourceOAuthClient  = "oauth_client"
	ResourceAPIKey       = "api_key"
	ResourceAuth         = "auth"
)
