package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a custom role within an organization
type Role struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index:idx_org_role,unique"`
	Name           string    `json:"name" gorm:"not null;index:idx_org_role,unique"` // e.g., "admin", "issuer", "rto", "student", or custom
	DisplayName    string    `json:"display_name" gorm:"not null"`
	Description    string    `json:"description"`
	IsSystem       bool      `json:"is_system" gorm:"default:false"` // true for default roles (admin), false for custom roles
	CreatedBy      uuid.UUID `json:"created_by" gorm:"type:uuid"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Permissions  []Permission  `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// Permission represents a specific permission that can be assigned to roles
type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"` // e.g., "member:invite", "cert:issue"
	DisplayName string    `json:"display_name" gorm:"not null"`
	Description string    `json:"description"`
	Category    string    `json:"category" gorm:"not null"` // e.g., "organization", "member", "certificate"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id" gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `json:"permission_id" gorm:"type:uuid;primaryKey"`
	CreatedAt    time.Time `json:"created_at"`

	// Relations
	Role       *Role       `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	Permission *Permission `json:"permission,omitempty" gorm:"foreignKey:PermissionID"`
}

// Permission constants - these are global system permissions
const (
	// Organization permissions
	PermissionOrgUpdate = "org:update"
	PermissionOrgDelete = "org:delete"
	PermissionOrgView   = "org:view"

	// Member permissions
	PermissionMemberInvite = "member:invite"
	PermissionMemberRemove = "member:remove"
	PermissionMemberUpdate = "member:update"
	PermissionMemberView   = "member:view"

	// Invitation permissions
	PermissionInvitationView   = "invitation:view"
	PermissionInvitationResend = "invitation:resend"
	PermissionInvitationCancel = "invitation:cancel"

	// Role management permissions
	PermissionRoleCreate = "role:create"
	PermissionRoleUpdate = "role:update"
	PermissionRoleDelete = "role:delete"
	PermissionRoleView   = "role:view"

	// Certificate permissions (for future use)
	PermissionCertIssue  = "cert:issue"
	PermissionCertRevoke = "cert:revoke"
	PermissionCertVerify = "cert:verify"
	PermissionCertView   = "cert:view"
)

// System role names
const (
	RoleNameAdmin = "admin" // System role - cannot be deleted, full permissions
)

// DefaultPermissions returns all system permissions that should be seeded
func DefaultPermissions() []Permission {
	return []Permission{
		// Organization
		{Name: PermissionOrgUpdate, DisplayName: "Update Organization", Description: "Update organization details", Category: "organization"},
		{Name: PermissionOrgDelete, DisplayName: "Delete Organization", Description: "Delete organization", Category: "organization"},
		{Name: PermissionOrgView, DisplayName: "View Organization", Description: "View organization details", Category: "organization"},

		// Members
		{Name: PermissionMemberInvite, DisplayName: "Invite Members", Description: "Invite new members to organization", Category: "member"},
		{Name: PermissionMemberRemove, DisplayName: "Remove Members", Description: "Remove members from organization", Category: "member"},
		{Name: PermissionMemberUpdate, DisplayName: "Update Members", Description: "Update member roles and details", Category: "member"},
		{Name: PermissionMemberView, DisplayName: "View Members", Description: "View organization members", Category: "member"},

		// Invitations
		{Name: PermissionInvitationView, DisplayName: "View Invitations", Description: "View pending invitations", Category: "invitation"},
		{Name: PermissionInvitationResend, DisplayName: "Resend Invitations", Description: "Resend pending invitations", Category: "invitation"},
		{Name: PermissionInvitationCancel, DisplayName: "Cancel Invitations", Description: "Cancel pending invitations", Category: "invitation"},

		// Roles
		{Name: PermissionRoleCreate, DisplayName: "Create Roles", Description: "Create custom roles", Category: "role"},
		{Name: PermissionRoleUpdate, DisplayName: "Update Roles", Description: "Update role permissions", Category: "role"},
		{Name: PermissionRoleDelete, DisplayName: "Delete Roles", Description: "Delete custom roles", Category: "role"},
		{Name: PermissionRoleView, DisplayName: "View Roles", Description: "View organization roles", Category: "role"},

		// Certificates
		{Name: PermissionCertIssue, DisplayName: "Issue Certificates", Description: "Issue new certificates", Category: "certificate"},
		{Name: PermissionCertRevoke, DisplayName: "Revoke Certificates", Description: "Revoke issued certificates", Category: "certificate"},
		{Name: PermissionCertVerify, DisplayName: "Verify Certificates", Description: "Verify certificate authenticity", Category: "certificate"},
		{Name: PermissionCertView, DisplayName: "View Certificates", Description: "View certificates", Category: "certificate"},
	}
}

// DefaultAdminPermissions returns all permissions for the admin role
func DefaultAdminPermissions() []string {
	return []string{
		// Organization
		PermissionOrgUpdate,
		PermissionOrgDelete,
		PermissionOrgView,
		// Members
		PermissionMemberInvite,
		PermissionMemberRemove,
		PermissionMemberUpdate,
		PermissionMemberView,
		// Invitations
		PermissionInvitationView,
		PermissionInvitationResend,
		PermissionInvitationCancel,
		// Roles
		PermissionRoleCreate,
		PermissionRoleUpdate,
		PermissionRoleDelete,
		PermissionRoleView,
		// Certificates
		PermissionCertIssue,
		PermissionCertRevoke,
		PermissionCertVerify,
		PermissionCertView,
	}
}
