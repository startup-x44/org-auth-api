package unit_test

import (
	"testing"

	"auth-service/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestSecurityLogic tests pure RBAC security logic without a DB
func TestSecurityLogic(t *testing.T) {

	//
	// ─────────────────────────────────────────────
	//  PERMISSION ASSIGNMENT SECURITY
	// ─────────────────────────────────────────────
	//
	t.Run("Permission assignment validation", func(t *testing.T) {

		org1ID := uuid.New()
		org2ID := uuid.New()

		systemPerm := &models.Permission{
			ID:             uuid.New(),
			Name:           "system:login",
			IsSystem:       true,
			OrganizationID: nil,
		}

		customPermOrg1 := &models.Permission{
			ID:             uuid.New(),
			Name:           "custom:org1",
			IsSystem:       false,
			OrganizationID: &org1ID,
		}

		roleOrg1 := &models.Role{
			ID:             uuid.New(),
			OrganizationID: &org1ID,
			Name:           "role1",
		}

		roleOrg2 := &models.Role{
			ID:             uuid.New(),
			OrganizationID: &org2ID,
			Name:           "role2",
		} // System perms: always allowed
		assert.True(t, canAssignPermissionToRole(systemPerm, roleOrg1))
		assert.True(t, canAssignPermissionToRole(systemPerm, roleOrg2))

		// Same-org custom perm: allowed
		assert.True(t, canAssignPermissionToRole(customPermOrg1, roleOrg1))

		// Cross-org custom perm: BLOCKED
		assert.False(t, canAssignPermissionToRole(customPermOrg1, roleOrg2))
	})

	//
	// ─────────────────────────────────────────────
	//  ORGANIZATION FILTERING LOGIC
	// ─────────────────────────────────────────────
	//
	t.Run("Organization filtering logic", func(t *testing.T) {

		org1ID := uuid.New()
		org2ID := uuid.New()

		perms := []*models.Permission{
			{ID: uuid.New(), Name: "system:global", IsSystem: true},
			{ID: uuid.New(), Name: "org1:custom", IsSystem: false, OrganizationID: &org1ID},
			{ID: uuid.New(), Name: "org2:custom", IsSystem: false, OrganizationID: &org2ID},
		}

		// Filter for org1: should get system + org1 perms
		p1 := filterPermissionsForOrganization(perms, org1ID)
		assert.Len(t, p1, 2)

		for _, perm := range p1 {
			if perm.OrganizationID != nil {
				assert.Equal(t, org1ID, *perm.OrganizationID)
			}
		}

		// Filter for org2: should get system + org2 perms
		p2 := filterPermissionsForOrganization(perms, org2ID)
		assert.Len(t, p2, 2)

		for _, perm := range p2 {
			if perm.OrganizationID != nil {
				assert.Equal(t, org2ID, *perm.OrganizationID)
			}
		}
	})

	//
	// ─────────────────────────────────────────────
	//  SECURITY CONSTANTS
	// ─────────────────────────────────────────────
	//
	t.Run("Security constants validation", func(t *testing.T) {
		assert.Equal(t, "admin", models.RoleNameAdmin)
		assert.Equal(t, "active", models.UserStatusActive)
		assert.Equal(t, "active", models.OrganizationStatusActive)
		assert.Equal(t, "active", models.MembershipStatusActive)
	})

	//
	// ─────────────────────────────────────────────
	//  BelongsToOrganization() LOGIC
	// ─────────────────────────────────────────────
	//
	t.Run("BelongsToOrganization helper", func(t *testing.T) {
		orgID := uuid.New()
		orgID2 := uuid.New()

		systemPerm := &models.Permission{IsSystem: true}
		orgPerm := &models.Permission{IsSystem: false, OrganizationID: &orgID}
		otherPerm := &models.Permission{IsSystem: false, OrganizationID: &orgID2}

		// System => always belongs
		assert.True(t, belongsToOrg(systemPerm, orgID))

		// Org permission => belongs
		assert.True(t, belongsToOrg(orgPerm, orgID))

		// Other org => should NOT belong
		assert.False(t, belongsToOrg(otherPerm, orgID))
	})
}

//
// ─────────────────────────────────────────────
//  HELPER LOGIC MATCHES REPO SECURITY RULES
// ─────────────────────────────────────────────
//

// Matches PermissionRepository.AssignToRole() validation
func canAssignPermissionToRole(permission *models.Permission, role *models.Role) bool {
	if permission.IsSystem || permission.OrganizationID == nil {
		return true // system/global perms are allowed everywhere
	}

	if role.OrganizationID == nil {
		return false // custom permission cannot be assigned to system role
	}

	return *permission.OrganizationID == *role.OrganizationID
}

// Matches PermissionRepository.GetRolePermissions() filtering
func filterPermissionsForOrganization(perms []*models.Permission, orgID uuid.UUID) []*models.Permission {
	var out []*models.Permission

	for _, p := range perms {

		if p.IsSystem || p.OrganizationID == nil {
			out = append(out, p)
			continue
		}

		if *p.OrganizationID == orgID {
			out = append(out, p)
		}
	}

	return out
}

// Matches permissionRepository.BelongsToOrganization()
func belongsToOrg(perm *models.Permission, orgID uuid.UUID) bool {
	if perm.IsSystem {
		return true
	}
	if perm.OrganizationID != nil && *perm.OrganizationID == orgID {
		return true
	}
	return false
}
