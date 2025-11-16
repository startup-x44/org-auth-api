# RBAC Middleware Usage Guide

## Overview

The auth service now includes **complete RBAC (Role-Based Access Control)** middleware for fine-grained permission enforcement with **strict organization isolation security**.

## 🔒 CRITICAL SECURITY REQUIREMENTS

### Organization Isolation 
**ALL RBAC operations MUST respect organization boundaries**. The system enforces strict isolation to prevent privilege escalation and cross-organization data access.

#### Security Rules:
1. **System Permissions**: Global permissions (e.g., `user:login`) can be assigned to any role
2. **Custom Permissions**: Organization-specific permissions can ONLY be assigned to roles within the SAME organization
3. **Permission Queries**: All permission lookups MUST filter by organization context
4. **Role Operations**: All role operations MUST validate organization membership

### Deprecated Methods - DO NOT USE
These methods bypass organization validation and create security vulnerabilities:

```go
// ❌ DANGEROUS - No organization validation
service.AssignPermissionsToRole(roleID, permissions)
service.RevokePermissionsFromRole(roleID, permissions) 
service.GetRole(roleID)
service.UpdateRole(roleID, req)
service.DeleteRole(roleID)

// ✅ SECURE - Organization-aware methods
service.AssignPermissionsToRoleWithOrganization(roleID, orgID, permissions)
service.RevokePermissionsFromRoleWithOrganization(roleID, orgID, permissions)
service.GetRoleWithOrganization(roleID, orgID)
service.UpdateRoleWithOrganization(roleID, orgID, req)
service.DeleteRoleWithOrganization(roleID, orgID)
```

### Secure Development Guidelines

#### 1. Always Use Organization Context
```go
// ✅ CORRECT - Organization context included
membership, err := repo.OrganizationMembership().GetByOrganizationAndUser(ctx, orgID, userID)
if err != nil {
    return fmt.Errorf("user not member of organization")
}

role, err := repo.Role().GetByIDAndOrganization(ctx, roleID, orgID)
```

#### 2. Validate Organization Membership
```go
// ✅ CORRECT - Validate user belongs to organization before role operations
func (s *roleService) getUserOrgMembership(ctx context.Context, userID, orgID string) (*models.OrganizationMembership, error) {
    membership, err := s.repo.OrganizationMembership().GetByOrganizationAndUser(ctx, orgID, userID)
    if err != nil {
        return nil, fmt.Errorf("user not found in organization: %w", err)
    }
    if membership.Status != models.MembershipStatusActive {
        return nil, fmt.Errorf("user membership not active")
    }
    return membership, nil
}
```

#### 3. Repository Layer Security
The repository layer enforces security constraints:

```go
// Permission assignment with validation
func (r *permissionRepository) AssignToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
    // Validates that custom permissions can only be assigned within same organization
    // System permissions can be assigned to any role
}

// Role permission creation with organization validation  
func (r *rolePermissionRepository) CreateWithValidation(ctx context.Context, rp *models.RolePermission) error {
    // Prevents cross-organization privilege escalation
}
```

## Available Middleware

### 1. `RequirePermission(permission string)`
Requires a **single specific permission**. Superadmins automatically bypass.

```go
// Only users with "member:invite" permission can access
router.POST("/invite", 
    authMiddleware.AuthRequired(),
    authMiddleware.RequirePermission("member:invite"),
    orgHandler.InviteUser,
)
```

### 2. `RequireAnyPermission(permissions ...string)`
Requires **at least one** of the specified permissions.

```go
// Users with either "member:view" OR "member:update" can access
router.GET("/members", 
    authMiddleware.AuthRequired(),
    authMiddleware.RequireAnyPermission("member:view", "member:update"),
    orgHandler.ListMembers,
)
```

### 3. `RequireAllPermissions(permissions ...string)`
Requires **all** of the specified permissions.

```go
// Users must have both "cert:issue" AND "cert:verify" 
router.POST("/certificates/issue-and-verify", 
    authMiddleware.AuthRequired(),
    authMiddleware.RequireAllPermissions("cert:issue", "cert:verify"),
    certHandler.IssueAndVerify,
)
```

## Permission Naming Convention

Permissions follow the pattern: `<resource>:<action>`

### Organization Permissions
- `org:view` - View organization details
- `org:update` - Update organization settings
- `org:delete` - Delete organization

### Member Management
- `member:view` - View members list
- `member:invite` - Invite new members
- `member:update` - Update member roles/status
- `member:remove` - Remove members

### Invitation Management
- `invitation:view` - View pending invitations
- `invitation:resend` - Resend invitations
- `invitation:revoke` - Revoke invitations

### Role Management
- `role:view` - View roles
- `role:create` - Create custom roles
- `role:update` - Update role permissions
- `role:delete` - Delete custom roles

### Certificate Management
- `cert:view` - View certificates
- `cert:issue` - Issue new certificates
- `cert:revoke` - Revoke certificates
- `cert:verify` - Verify certificates

## Example: Protected Routes

```go
func SetupOrganizationRoutes(router *gin.Engine, authMw *middleware.AuthMiddleware, orgHandler *handler.OrganizationHandler) {
    org := router.Group("/api/v1/organizations")
    org.Use(authMw.AuthRequired()) // All routes require authentication
    {
        // View organization - requires "org:view" permission
        org.GET("/:orgId", 
            authMw.RequirePermission("org:view"),
            orgHandler.GetOrganization,
        )

        // Update organization - requires "org:update" permission
        org.PUT("/:orgId", 
            authMw.RequirePermission("org:update"),
            orgHandler.UpdateOrganization,
        )

        // Delete organization - requires "org:delete" permission
        org.DELETE("/:orgId", 
            authMw.RequirePermission("org:delete"),
            orgHandler.DeleteOrganization,
        )

        // Invite member - requires "member:invite" permission
        org.POST("/:orgId/members/invite", 
            authMw.RequirePermission("member:invite"),
            orgHandler.InviteUser,
        )

        // Update member - requires "member:update" permission
        org.PUT("/:orgId/members/:userId", 
            authMw.RequirePermission("member:update"),
            orgHandler.UpdateMembership,
        )

        // Remove member - requires "member:remove" permission
        org.DELETE("/:orgId/members/:userId", 
            authMw.RequirePermission("member:remove"),
            orgHandler.RemoveMember,
        )
    }
}
```

## Superadmin Bypass

**Superadmins automatically bypass ALL permission checks**:
- `is_superadmin: true` in JWT claims grants unrestricted access
- Useful for platform administrators who need global access

## JWT Token Structure

Access tokens now include cached permissions:

```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "organization_id": "987fcdeb-51a2-43d7-8a9e-123456789abc",
  "role_id": "456789ab-cdef-1234-5678-9abcdef01234",
  "organization_role": "admin",
  "permissions": [
    "org:view",
    "org:update",
    "member:view",
    "member:invite",
    "member:update",
    "member:remove",
    "role:view",
    "role:create"
  ],
  "is_superadmin": false
}
```

## Error Responses

### 401 Unauthorized (No token or invalid token)
```json
{
  "success": false,
  "message": "Invalid or expired token"
}
```

### 403 Forbidden (Insufficient permissions)
```json
{
  "success": false,
  "message": "Insufficient permissions",
  "required": "member:invite"
}
```

### 403 Forbidden (Missing multiple permissions)
```json
{
  "success": false,
  "message": "Insufficient permissions",
  "missing": ["cert:issue", "cert:verify"]
}
```

## Migration from Role-Based to Permission-Based

**Before (Role-based check):**
```go
org.POST("/invite", 
    authMw.AuthRequired(),
    authMw.OrgAdminRequired(), // ❌ Only admins can invite
    orgHandler.InviteUser,
)
```

**After (Permission-based check):**
```go
org.POST("/invite", 
    authMw.AuthRequired(),
    authMw.RequirePermission("member:invite"), // ✅ Any role with this permission
    orgHandler.InviteUser,
)
```

## Benefits

1. **Fine-grained control**: Assign specific permissions to custom roles
2. **Flexibility**: Create roles like "Recruiter" (can invite but not remove)
3. **Security**: Permissions cached in JWT for fast validation
4. **Scalability**: No database hits for permission checks on every request
5. **Slack-like UX**: Per-organization custom roles with flexible permissions

## Admin Role

The system `admin` role automatically receives **all permissions**:
```go
if role.Name == models.RoleNameAdmin && role.IsSystem {
    permissions = models.DefaultAdminPermissions() // All permissions
}
```

## Custom Roles Example

Organization admins can create custom roles:

```json
{
  "name": "Recruiter",
  "description": "Can invite and manage members but cannot issue certificates",
  "permissions": [
    "member:view",
    "member:invite",
    "member:update",
    "invitation:view",
    "invitation:resend",
    "invitation:revoke"
  ]
}
```

Then assign this role to users who need those specific capabilities.

## 🔍 Security Testing

### Required Security Tests
All RBAC implementations must include tests that verify:

1. **Cross-Organization Permission Assignment Prevention**
```go
// Test: Attempt to assign custom permission from org1 to role in org2
err := repo.Permission().AssignToRole(ctx, role2ID, customPerm1ID)
assert.Error(t, err) // Should fail with security error
```

2. **Organization-Filtered Permission Queries**
```go
// Test: GetRolePermissions only returns org-scoped permissions
perms, err := repo.Permission().GetRolePermissions(ctx, roleID)
for _, perm := range perms {
    if perm.OrganizationID != nil {
        assert.Equal(t, expectedOrgID, *perm.OrganizationID)
    }
}
```

3. **Privilege Escalation Prevention**
```go
// Test: Direct role permission creation bypassing validation should fail
invalidRP := &models.RolePermission{
    RoleID: roleInOrg1,
    PermissionID: customPermFromOrg2,
}
err := repo.RolePermission().CreateWithValidation(ctx, invalidRP)
assert.Error(t, err) // Should prevent cross-org escalation
```

### Security Audit Checklist

When implementing RBAC features, verify:

- [ ] All repository methods include organization context
- [ ] Service layer validates user membership before role operations
- [ ] Handler layer uses organization-scoped service methods
- [ ] No direct use of deprecated security-bypassing methods
- [ ] Permission assignments respect system vs custom permission rules
- [ ] Tests cover cross-organization attack scenarios
- [ ] Error messages don't leak organization structure information

## 🚨 Common Security Vulnerabilities

### 1. Missing Organization Validation
```go
// ❌ VULNERABLE - No org validation
role, err := repo.Role().GetByID(ctx, roleID)

// ✅ SECURE - Organization validation
role, err := repo.Role().GetByIDAndOrganization(ctx, roleID, orgID)
```

### 2. Cross-Organization Data Leakage
```go
// ❌ VULNERABLE - Could return permissions from other orgs
perms, err := repo.Permission().GetAll(ctx)

// ✅ SECURE - Organization-filtered
perms, err := repo.Permission().GetByOrganization(ctx, orgID)
```

### 3. Privilege Escalation Through Service Bypass
```go
// ❌ VULNERABLE - Bypasses organization checks
repo.RolePermission().Create(ctx, &models.RolePermission{
    RoleID: targetRoleID,
    PermissionID: adminPermissionID, // Could be from different org
})

// ✅ SECURE - Validation enforced
repo.RolePermission().CreateWithValidation(ctx, rolePermission)
```

Remember: **Organization isolation is the foundation of multi-tenant security**. Every RBAC operation must respect these boundaries.
