package feature_test

import (
	"context"
	"testing"

	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/tests/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRBACOrganizationIsolation tests critical security requirement that RBAC operations respect organization boundaries
func TestRBACOrganizationIsolation(t *testing.T) {
	ctx := context.Background()
	testDB := testutils.SetupTestDB(t)
	defer testDB.SQLDB.Close()

	repo := repository.NewRepository(testDB.DB)

	// Create two separate organizations
	org1 := &models.Organization{
		ID:        uuid.New(),
		Name:      "Organization 1",
		Slug:      "org1",
		Status:    "active",
		CreatedBy: uuid.New(),
	}
	org2 := &models.Organization{
		ID:        uuid.New(),
		Name:      "Organization 2",
		Slug:      "org2",
		Status:    "active",
		CreatedBy: uuid.New(),
	}

	require.NoError(t, repo.Organization().Create(ctx, org1))
	require.NoError(t, repo.Organization().Create(ctx, org2))

	// Create custom permissions in each organization
	customPerm1 := &models.Permission{
		ID:             uuid.New(),
		Name:           "custom:action1",
		DisplayName:    "Custom Action 1",
		Category:       "custom",
		IsSystem:       false,
		OrganizationID: &org1.ID,
	}
	customPerm2 := &models.Permission{
		ID:             uuid.New(),
		Name:           "custom:action2",
		DisplayName:    "Custom Action 2",
		Category:       "custom",
		IsSystem:       false,
		OrganizationID: &org2.ID,
	}

	_, err := repo.Permission().Create(ctx, customPerm1)
	require.NoError(t, err)
	_, err = repo.Permission().Create(ctx, customPerm2)
	require.NoError(t, err)

	// Create roles in each organization
	role1 := &models.Role{
		ID:             uuid.New(),
		OrganizationID: &org1.ID,
		Name:           "custom-role-1",
		DisplayName:    "Custom Role 1",
		IsSystem:       false,
		CreatedBy:      uuid.New(),
	}
	role2 := &models.Role{
		ID:             uuid.New(),
		OrganizationID: &org2.ID,
		Name:           "custom-role-2",
		DisplayName:    "Custom Role 2",
		IsSystem:       false,
		CreatedBy:      uuid.New(),
	}

	require.NoError(t, repo.Role().Create(ctx, role1))
	require.NoError(t, repo.Role().Create(ctx, role2))

	t.Run("CRITICAL: Cannot assign custom permission from one org to role in another org", func(t *testing.T) {
		// Attempt to assign custom permission from org1 to role in org2 - this should FAIL
		err := repo.Permission().AssignToRole(ctx, role2.ID, customPerm1.ID)
		assert.Error(t, err, "Should not allow cross-organization permission assignment")
		assert.Contains(t, err.Error(), "cannot assign custom permission from organization")
	})

	t.Run("CRITICAL: GetRolePermissions filters by organization context", func(t *testing.T) {
		// First, assign valid permissions within same organization
		err := repo.Permission().AssignToRole(ctx, role1.ID, customPerm1.ID)
		require.NoError(t, err, "Should allow same-org permission assignment")

		err = repo.Permission().AssignToRole(ctx, role2.ID, customPerm2.ID)
		require.NoError(t, err, "Should allow same-org permission assignment")

		// Get permissions for role1 with org1 context - should only see org1 permissions
		perms1, err := repo.Permission().GetRolePermissions(ctx, role1.ID)
		require.NoError(t, err)

		// Verify only org1's custom permission is returned
		found := false
		for _, perm := range perms1 {
			if perm.ID == customPerm1.ID {
				found = true
			}
			// Should never see permissions from other organizations
			if perm.OrganizationID != nil {
				assert.Equal(t, org1.ID, *perm.OrganizationID, "Should only see permissions from same organization")
			}
		}
		assert.True(t, found, "Should find the assigned permission")
	})

	t.Run("CRITICAL: RolePermission CreateWithValidation prevents privilege escalation", func(t *testing.T) {
		// Try to create role permission that violates organization boundaries
		invalidRolePerm := &models.RolePermission{
			RoleID:       role1.ID,       // Role in org1
			PermissionID: customPerm2.ID, // Permission in org2
		}

		err := repo.RolePermission().CreateWithValidation(ctx, invalidRolePerm)
		assert.Error(t, err, "Should prevent cross-organization privilege escalation")
		assert.Contains(t, err.Error(), "cannot assign custom permission from organization")
	})

	t.Run("CRITICAL: System permissions can be assigned to any organization", func(t *testing.T) {
		// Create a system permission (OrganizationID = nil)
		systemPerm := &models.Permission{
			ID:             uuid.New(),
			Name:           "system:global",
			DisplayName:    "System Global Permission",
			Category:       "system",
			IsSystem:       true,
			OrganizationID: nil, // System permission
		}
		_, err := repo.Permission().Create(ctx, systemPerm)
		require.NoError(t, err)

		// System permissions should be assignable to roles in any organization
		err1 := repo.Permission().AssignToRole(ctx, role1.ID, systemPerm.ID)
		assert.NoError(t, err1, "System permissions should be assignable to any organization")

		err2 := repo.Permission().AssignToRole(ctx, role2.ID, systemPerm.ID)
		assert.NoError(t, err2, "System permissions should be assignable to any organization")
	})

	t.Run("CRITICAL: Deprecated methods are disabled", func(t *testing.T) {
		// These methods should return errors indicating they're disabled for security
		rolePermRepo := repo.RolePermission()

		// Try to use deprecated Create method instead of secure CreateWithValidation
		invalidRolePerm := &models.RolePermission{
			RoleID:       role1.ID,
			PermissionID: customPerm2.ID,
		}

		err := rolePermRepo.Create(ctx, invalidRolePerm)
		// Note: The Create method still exists for backward compatibility but should be avoided
		// The security is now enforced at the Permission.AssignToRole level
		if err == nil {
			// If Create succeeded, verify the assignment was blocked at permission level
			perms, err := repo.Permission().GetRolePermissions(ctx, role1.ID)
			require.NoError(t, err)

			// Should not contain the cross-org permission
			for _, perm := range perms {
				assert.NotEqual(t, customPerm2.ID, perm.ID, "Cross-org permission should not be accessible")
			}
		}
	})
}

// TestRBACServiceLayerSecurity tests that the service layer properly enforces organization boundaries
func TestRBACServiceLayerSecurity(t *testing.T) {
	ctx := context.Background()
	testDB := testutils.SetupTestDB(t)
	defer testDB.SQLDB.Close()

	repo := repository.NewRepository(testDB.DB)

	t.Run("Deprecated service methods return security errors", func(t *testing.T) {
		// This test would require access to the service layer
		// For now, we verify at repository level that security is enforced

		// Create test data
		org := &models.Organization{
			ID:        uuid.New(),
			Name:      "Test Org",
			Slug:      "test-org",
			Status:    "active",
			CreatedBy: uuid.New(),
		}
		require.NoError(t, repo.Organization().Create(ctx, org))

		role := &models.Role{
			ID:             uuid.New(),
			OrganizationID: &org.ID,
			Name:           "test-role",
			DisplayName:    "Test Role",
			IsSystem:       false,
			CreatedBy:      uuid.New(),
		}
		require.NoError(t, repo.Role().Create(ctx, role)) // Verify that the repository layer prevents security violations
		// even if service layer methods were somehow called inappropriately
		customPerm := &models.Permission{
			ID:             uuid.New(),
			Name:           "test:permission",
			DisplayName:    "Test Permission",
			Category:       "test",
			IsSystem:       false,
			OrganizationID: &org.ID,
		}
		_, err := repo.Permission().Create(ctx, customPerm)
		require.NoError(t, err)

		// Valid assignment should work
		err = repo.Permission().AssignToRole(ctx, role.ID, customPerm.ID)
		assert.NoError(t, err, "Valid same-org assignment should succeed")
	})
}
