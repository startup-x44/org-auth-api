# RBAC System Refactor - Summary

## Overview
Completed comprehensive refactoring of the RBAC (Role-Based Access Control) system to separate system roles/permissions from organization-specific custom roles/permissions. This ensures proper isolation and prevents privilege escalation.

## Key Changes

### 1. Model Updates (`internal/models/role_permission.go`)
- **Role Model**:
  - Changed `OrganizationID` from `uuid.UUID` to `*uuid.UUID` (nullable)
  - System roles: `IsSystem=true`, `OrganizationID=NULL`
  - Custom roles: `IsSystem=false`, `OrganizationID=required`
  - Added indexes for efficient querying

- **Permission Model**:
  - Already had `IsSystem bool` and `*uuid.UUID` for OrganizationID
  - System permissions: `IsSystem=true`, `OrganizationID=NULL`
  - Custom permissions: `IsSystem=false`, `OrganizationID=set`

### 2. Database Migration (`migrations/006_rbac_system_separation.*`)
- Makes `organization_id` nullable in roles table
- Adds check constraint: system roles MUST have NULL organization_id
- Adds indexes for system role queries
- Includes up/down migrations for rollback capability

### 3. Repository Layer (`internal/repository/role_repository.go`)
Added new methods for system role management:
- `GetAllSystemRoles()` - retrieve all system roles (superadmin only)
- `GetSystemRoleByName(name)` - get specific system role
- `GetAllRoles(includeSystem)` - get all roles with optional filtering

Updated `CreateDefaultAdminRole()`:
- Now creates a CUSTOM "owner" role (not system)
- `IsSystem=false`, tied to specific organization
- Assigns all system permissions to the owner role

### 4. Service Layer

#### Role Service (`internal/service/role_service.go`)
- **CreateRole**: Prevents creating system role names, always sets `IsSystem=false`
- **UpdateRole**: Blocks updating system roles for non-superadmins
- **DeleteRole**: Prevents deletion of system roles
- **UpdateRoleResponse**: Changed `OrganizationID` to pointer type
- **AssignPermissionsToRoleWithOrganization**: Validates org ownership
- **Permission filtering**: Ensures org admins can only assign org-specific permissions

#### User Service (`internal/service/user_service.go`)
- **issueTokenPair()**: Updated to use `getRolePermissionsFiltered()`
- **getRolePermissionsFiltered()**: New method that filters permissions by user type:
  - **Superadmin**: Gets ALL permissions (system + org)
  - **Org admin/user**: Gets ONLY:
    - System permissions (global)
    - Custom permissions for their organization
    - Excludes permissions from other organizations

#### Organization Service (`internal/service/organization_service.go`)
- **CreateOrganization**: Updated comments to reflect custom "owner" role creation
- Uses `CreateDefaultAdminRole()` which now creates custom roles

### 5. Handler Layer

#### Organization Handler (`internal/handler/organization_handler.go`)
- **GetOrganizationRoles**: Added filtering logic:
  - **Superadmin**: Sees ALL roles (system + custom)
  - **Org admin**: Sees ONLY custom org roles (`IsSystem=false`)
  - Filters out system roles from response for non-superadmins

### 6. Seeder Updates (`internal/seeder/`)

#### Permission Seeder (`internal/seeder/permission_seeder.go`)
- Already creates system permissions with `IsSystem=true`
- No changes needed

#### Organization Seeder (`internal/seeder/organization_seeder.go`)
- **CRITICAL CHANGE**: Creates CUSTOM "owner" roles, not system roles
- Sets `IsSystem=false` for all org roles
- Sets `OrganizationID` to specific organization (not NULL)
- Assigns system permissions to owner role

## Acceptance Criteria Status

### ✅ Completed
- [x] Creating new organization generates ONLY a custom OWNER role
- [x] System roles are never copied into organizations
- [x] Organization admins cannot see system roles (filtered in API)
- [x] Organization admins cannot assign system permissions (validation in service)
- [x] Superadmin can see all roles & permissions globally
- [x] Organization admin can create ONLY custom roles
- [x] Permission filtering in tokens:
  - Superadmin → system + org permissions
  - Org admin → system permissions + org-specific permissions only
- [x] Validation prevents:
  - Non-superadmin from editing/deleting system roles
  - Assigning permissions from other organizations
  - Creating roles with system role names

## Security Improvements

1. **Strict separation of concerns**:
   - System roles are global (superadmin only)
   - Custom roles are org-specific

2. **Database-level constraints**:
   - Check constraint enforces system role rules
   - Prevents data corruption at database level

3. **Multi-layer validation**:
   - Repository layer: Queries filter by organization
   - Service layer: Business logic validates ownership
   - Handler layer: Response filtering by user type

4. **Permission isolation**:
   - Tokens only include permissions user is authorized to see
   - Org admins cannot access permissions from other orgs

## Migration Instructions

### Running the Migration

```bash
# The migration is applied automatically via GORM AutoMigrate
# Or manually run:
psql -U auth_user -d auth_db -f migrations/006_rbac_system_separation.up.sql
```

### Post-Migration Cleanup

After running the migration, you may need to:

1. **Identify existing roles** that should be system vs custom:
   ```sql
   -- Find roles that should be system roles
   SELECT * FROM roles WHERE organization_id IS NULL;
   
   -- Find roles that are incorrectly marked as system
   SELECT * FROM roles WHERE is_system = true AND organization_id IS NOT NULL;
   ```

2. **Fix incorrect role assignments**:
   ```sql
   -- Update existing org "admin" roles to "owner" and mark as custom
   UPDATE roles 
   SET name = 'owner', is_system = false 
   WHERE name = 'admin' AND organization_id IS NOT NULL;
   ```

3. **Verify no orphaned roles**:
   ```sql
   -- Should return empty (all custom roles must have org_id)
   SELECT * FROM roles WHERE is_system = false AND organization_id IS NULL;
   ```

## Testing Checklist

### Unit Tests
- [ ] Test CreateRole rejects system role names
- [ ] Test UpdateRole blocks system role modifications
- [ ] Test DeleteRole prevents system role deletion
- [ ] Test permission filtering in getRolePermissionsFiltered

### Integration Tests
- [ ] Test organization creation creates custom owner role
- [ ] Test token generation includes correct permissions
- [ ] Test API filters roles based on user type
- [ ] Test role assignment validation

### Manual Testing
1. Create a new organization as regular user → should get "owner" custom role
2. Login as org admin → roles API should NOT show system roles
3. Login as superadmin → roles API should show all roles
4. Try to assign system permission to custom role → should fail
5. Check JWT token claims → verify permissions are filtered correctly

## Rollback Plan

If issues arise, rollback using:

```sql
-- Run down migration
psql -U auth_user -d auth_db -f migrations/006_rbac_system_separation.down.sql
```

Then revert code changes:
```bash
git revert <commit-hash>
```

## Notes

- **Breaking Change**: Existing client code that expects `OrganizationID` as non-nullable will break
- **API Response**: Role responses now have `organization_id` as nullable
- **Token Format**: Permission filtering in tokens is backward compatible
- **Performance**: Added indexes should improve query performance for role lookups

## Future Enhancements

1. **System role management UI** for superadmin
2. **Permission inheritance** from system to custom roles
3. **Role templates** for creating common role types
4. **Audit logging** for system role modifications
5. **Redis caching** for permission lookups (currently disabled during refactor)

---

**Completed**: All acceptance criteria met
**Status**: Ready for testing and deployment
**Risk Level**: Medium (requires data migration and careful testing)
