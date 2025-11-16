# RBAC System - Quick Reference Guide

## System Roles vs Custom Roles

| Attribute | System Role | Custom Role |
|-----------|-------------|-------------|
| `IsSystem` | `true` | `false` |
| `OrganizationID` | `NULL` | Required (UUID) |
| Created by | Superadmin/Seeder | Org Admin |
| Visible to | Superadmin only | Org members only |
| Modifiable by | Superadmin only | Org admin with `role:update` permission |
| Deletable | No | Yes (if not in use) |
| Examples | N/A (reserved for future platform roles) | "owner", "admin", "member", "viewer" |

## System Permissions vs Custom Permissions

| Attribute | System Permission | Custom Permission |
|-----------|-------------------|-------------------|
| `IsSystem` | `true` | `false` |
| `OrganizationID` | `NULL` | Required (UUID) |
| Created by | Seeder/Migration | Org Admin |
| Assignable to | Any role in any org | Only roles in same org |
| Visible to | All users | Org members + Superadmin |
| Examples | `member:invite`, `org:update` | `custom:approve_documents` |

## Default Organization Setup

When a new organization is created:

```go
// Created automatically:
1. Organization record
2. Custom "owner" role:
   - Name: "owner"
   - IsSystem: false
   - OrganizationID: <org-id>
   - All system permissions assigned
3. Membership for creator with owner role
```

## Permission Filtering in Tokens

```go
// Superadmin token includes:
- All system permissions
- All permissions from current organization

// Regular user token includes:
- System permissions (global access patterns)
- Custom permissions from current organization only
- NO permissions from other organizations
```

## API Endpoints Behavior

### GET /organizations/:orgId/roles

**Superadmin**:
```json
[
  // Currently, system roles are NOT shown here
  // Only organization-specific custom roles
  {
    "id": "uuid",
    "organization_id": "org-uuid",
    "name": "owner",
    "is_system": false
  }
]
```

**Org Admin**:
```json
[
  // Only custom org roles (IsSystem=false)
  {
    "id": "uuid",
    "organization_id": "org-uuid",
    "name": "owner",
    "is_system": false
  }
  // System roles are filtered out
]
```

### POST /organizations/:orgId/roles

**Validation**:
```go
// Prevents:
- Setting name = "admin", "system_admin", "superadmin"
- Setting IsSystem = true (always false for user-created roles)
- Cross-org permission assignment

// Allows:
- Any custom role name (except reserved)
- Assigning system permissions
- Assigning org-specific permissions
```

### PUT /organizations/:orgId/roles/:roleId

**Validation**:
```go
// Prevents:
- Updating system roles (IsSystem=true)
- Changing OrganizationID
- Assigning permissions from other orgs

// Allows:
- Updating DisplayName, Description
- Replacing permission assignments
- Modifying custom roles only
```

### DELETE /organizations/:orgId/roles/:roleId

**Validation**:
```go
// Prevents:
- Deleting system roles
- Deleting roles in use (memberCount > 0)
- Cross-org deletion

// Allows:
- Deleting unused custom roles
- Organization-scoped deletion only
```

## Service Layer Methods

### Creating Roles

```go
// ✅ Correct: Creates custom role
roleService.CreateRole(ctx, &CreateRoleRequest{
    OrganizationID: orgID,
    Name:          "custom_role",  // NOT "admin"
    DisplayName:   "Custom Role",
    Permissions:   []string{"member:view", "org:view"},
})
// IsSystem is automatically set to false

// ❌ Wrong: Cannot create system role names
roleService.CreateRole(ctx, &CreateRoleRequest{
    Name: "admin",  // Rejected!
})
```

### Assigning Permissions

```go
// ✅ Correct: Org-scoped permission assignment
roleService.AssignPermissionsToRoleWithOrganization(
    ctx, roleID, orgID,
    []string{"member:invite", "custom:permission"}
)
// Validates:
// 1. Role belongs to org
// 2. Role is not system
// 3. Permissions are system OR belong to same org

// ❌ Wrong: Deprecated unsafe method
roleService.AssignPermissionsToRole(ctx, roleID, permissions)
// Returns error: "use AssignPermissionsToRoleWithOrganization"
```

## Database Queries

### Get Organization Roles (Excludes System)

```sql
-- Custom org roles only
SELECT * FROM roles 
WHERE organization_id = $1 
AND is_system = false
ORDER BY name;
```

### Get System Roles (Superadmin Only)

```sql
-- System roles only
SELECT * FROM roles 
WHERE is_system = true 
AND organization_id IS NULL
ORDER BY name;
```

### Get User Permissions

```sql
-- Via role membership, filtered by org
SELECT DISTINCT p.name
FROM permissions p
INNER JOIN role_permissions rp ON p.id = rp.permission_id
INNER JOIN organization_memberships om ON om.role_id = rp.role_id
WHERE om.user_id = $1 
AND om.organization_id = $2
AND (
  -- System permissions (available to all)
  (p.is_system = true AND p.organization_id IS NULL)
  OR
  -- Org-specific custom permissions
  (p.is_system = false AND p.organization_id = $2)
);
```

## Security Validations

### Repository Layer
- Uses `WHERE organization_id = $1` for all custom role queries
- Prevents cross-org data leakage
- System role queries require explicit methods

### Service Layer
- Checks `role.IsSystem` before modification
- Validates `permission.OrganizationID` matches role's org
- Prevents privilege escalation through permission assignment

### Handler Layer
- Filters response based on `is_superadmin` context value
- Hides system roles from non-superadmin users
- Enforces permission checks via middleware

## Common Patterns

### Check if User Can Manage Roles

```go
// In handler
if !h.checkPermission(c, orgID, "role:update") {
    c.JSON(403, gin.H{"error": "Insufficient permissions"})
    return
}

// Permission is checked via:
// 1. Is user superadmin? → Allow
// 2. Does user have "role:update" in their org role? → Allow
// 3. Otherwise → Deny
```

### Create Organization with Owner

```go
// Automatically handled by CreateOrganization
org, err := orgService.CreateOrganization(ctx, &CreateOrgRequest{
    Name: "Acme Corp",
})
// Creates:
// - Organization
// - Custom "owner" role (IsSystem=false)
// - Membership for creator with owner role
```

### Filter Permissions in Token

```go
// Automatically handled by issueTokenPair
tokenPair, err := userService.issueTokenPair(
    ctx, user, orgID, roleID, sessionID
)
// Token includes:
// - For superadmin: all permissions
// - For org user: system + org permissions only
```

## Migration Commands

```bash
# Apply migration
./dev.sh shell
cd /app
go run cmd/migrate/main.go up

# Or via psql
psql -U auth_user -d auth_db -f migrations/006_rbac_system_separation.up.sql

# Rollback if needed
psql -U auth_user -d auth_db -f migrations/006_rbac_system_separation.down.sql
```

## Troubleshooting

### Issue: Organization admin sees system roles
**Solution**: Check handler filtering logic in `GetOrganizationRoles`

### Issue: Cannot create roles
**Solution**: Verify role name is not reserved ("admin", "system_admin")

### Issue: Permissions not in token
**Solution**: Check `getRolePermissionsFiltered` - may be filtered out

### Issue: "Cannot update system role" error
**Solution**: System roles are read-only, create a custom role instead

### Issue: Cross-org permission assignment
**Solution**: Validation prevents this - check error message for details

---

For detailed implementation info, see `RBAC_REFACTOR_SUMMARY.md`
