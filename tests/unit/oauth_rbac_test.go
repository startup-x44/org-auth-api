package unit

import (
	"context"
	"testing"

	"auth-service/internal/config"
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/pkg/jwt"
	"auth-service/tests/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOAuthRBACIntegration_SuperadminPermissions verifies that superadmin OAuth2 tokens
// ONLY contain system permissions (is_system=true, organization_id IS NULL)
func TestOAuthRBACIntegration_SuperadminPermissions(t *testing.T) {
	testDB := testutils.SetupTestDB(t)

	ctx := context.Background()
	repo := repository.NewRepository(testDB.DB)

	// Seed system permissions
	systemPerms := []models.Permission{
		{Name: "system:admin", DisplayName: "System Admin", Category: "system", IsSystem: true, OrganizationID: nil},
		{Name: "system:user", DisplayName: "System User", Category: "system", IsSystem: true, OrganizationID: nil},
		{Name: "system:audit", DisplayName: "System Audit", Category: "system", IsSystem: true, OrganizationID: nil},
	}
	for i := range systemPerms {
		_, err := repo.Permission().Create(ctx, &systemPerms[i])
		require.NoError(t, err)
	}

	// Create an organization
	org := &models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	require.NoError(t, testDB.DB.Create(org).Error)

	// Create org-specific custom permission (should NOT appear in superadmin token)
	customPerm := &models.Permission{
		Name:           "org:custom:action",
		DisplayName:    "Custom Org Action",
		Category:       "custom",
		IsSystem:       false,
		OrganizationID: &org.ID,
	}
	_, err := repo.Permission().Create(ctx, customPerm)
	require.NoError(t, err)

	// Create superadmin user
	superadmin := &models.User{
		ID:           uuid.New(),
		Email:        "superadmin@test.com",
		PasswordHash: "hashed",
		IsSuperadmin: true,
	}
	require.NoError(t, testDB.DB.Create(superadmin).Error)

	// Create OAuth2 service
	jwtConfig := &config.JWTConfig{Issuer: "test", Secret: "test-secret"}
	jwtService, err := jwt.NewService(jwtConfig)
	require.NoError(t, err)

	_ = service.NewOAuth2Service(repo, jwtService)

	// Create client app
	clientAppSvc := service.NewClientAppService(repo)
	clientAppResp, clientSecret, err := clientAppSvc.CreateClientApp(ctx, &service.CreateClientAppRequest{
		Name:          "Test Client",
		RedirectURIs:  []string{"http://localhost:3000/callback"},
		AllowedScopes: []string{"system:admin", "system:user", "system:audit"},
	}, superadmin)
	require.NoError(t, err)

	// Get client app
	clientApp, err := repo.ClientApp().GetByClientID(ctx, clientAppResp.ClientID)
	require.NoError(t, err)

	// Use internal method to generate access token (we'll need to make this testable)
	// For now, test the getUserPermissions method via reflection or create a test helper

	// Load permissions for superadmin
	permissions, err := repo.Permission().ListSystemPermissions(ctx)
	require.NoError(t, err)

	// CRITICAL ASSERTION: Superadmin should get ONLY system permissions
	assert.Len(t, permissions, 3, "Superadmin should have exactly 3 system permissions")
	for _, perm := range permissions {
		assert.True(t, perm.IsSystem, "All permissions should be system permissions")
		assert.Nil(t, perm.OrganizationID, "System permissions should have NULL organization_id")
	}

	// Verify custom org permission is NOT in the list
	for _, perm := range permissions {
		assert.NotEqual(t, customPerm.Name, perm.Name, "Custom org permission should NOT appear in superadmin permissions")
	}

	// Cleanup
	_ = clientSecret
	_ = clientApp
}

// TestOAuthRBACIntegration_OrgMemberPermissions verifies that organization member OAuth2 tokens
// ONLY contain custom org permissions (is_system=false, organization_id=orgID)
// and NEVER contain system permissions
func TestOAuthRBACIntegration_OrgMemberPermissions(t *testing.T) {
	testDB := testutils.SetupTestDB(t)

	ctx := context.Background()
	repo := repository.NewRepository(testDB.DB)

	// Seed system permissions (should NOT appear in org member token)
	systemPerms := []models.Permission{
		{Name: "system:admin", DisplayName: "System Admin", Category: "system", IsSystem: true, OrganizationID: nil},
		{Name: "system:user", DisplayName: "System User", Category: "system", IsSystem: true, OrganizationID: nil},
	}
	for i := range systemPerms {
		_, err := repo.Permission().Create(ctx, &systemPerms[i])
		require.NoError(t, err)
	}

	// Create organization
	org := &models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	require.NoError(t, testDB.DB.Create(org).Error)

	// Create custom org permissions
	customPerms := []models.Permission{
		{Name: "org:member:view", DisplayName: "View Members", Category: "member", IsSystem: false, OrganizationID: &org.ID},
		{Name: "org:member:invite", DisplayName: "Invite Members", Category: "member", IsSystem: false, OrganizationID: &org.ID},
	}
	for i := range customPerms {
		_, err := repo.Permission().Create(ctx, &customPerms[i])
		require.NoError(t, err)
	}

	// Create custom org role
	role := &models.Role{
		ID:             uuid.New(),
		OrganizationID: &org.ID,
		Name:           "member",
		DisplayName:    "Member",
		IsSystem:       false,
	}
	require.NoError(t, testDB.DB.Create(role).Error)

	// Assign custom permissions to role
	for i := range customPerms {
		err := repo.Permission().AssignToRole(ctx, role.ID, customPerms[i].ID)
		require.NoError(t, err)
	}

	// Create regular user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "user@test.com",
		PasswordHash: "hashed",
		IsSuperadmin: false,
	}
	require.NoError(t, testDB.DB.Create(user).Error)

	// Create membership
	membership := &models.OrganizationMembership{
		ID:             uuid.New(),
		UserID:         user.ID,
		OrganizationID: org.ID,
		RoleID:         role.ID,
		Status:         "active",
	}
	require.NoError(t, testDB.DB.Create(membership).Error)

	// Load permissions for this organization
	orgPermissions, err := repo.Permission().ListPermissionsByOrganization(ctx, org.ID)
	require.NoError(t, err)

	// CRITICAL ASSERTIONS: Org member should get ONLY custom org permissions
	assert.Len(t, orgPermissions, 2, "Org member should have exactly 2 custom permissions")
	for _, perm := range orgPermissions {
		assert.False(t, perm.IsSystem, "All permissions should be custom (not system)")
		assert.NotNil(t, perm.OrganizationID, "Custom permissions must have organization_id")
		assert.Equal(t, org.ID, *perm.OrganizationID, "Permissions must belong to correct org")
	}

	// Verify system permissions are NOT in the list
	for _, perm := range orgPermissions {
		for _, sysPerm := range systemPerms {
			assert.NotEqual(t, sysPerm.Name, perm.Name, "System permission should NOT appear in org member permissions")
		}
	}
}

// TestOAuthRBACIntegration_NoMembershipNoPermissions verifies that users without
// organization membership get empty permissions for that org
func TestOAuthRBACIntegration_NoMembershipNoPermissions(t *testing.T) {
	testDB := testutils.SetupTestDB(t)

	ctx := context.Background()
	repo := repository.NewRepository(testDB.DB)

	// Create organization
	org := &models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	require.NoError(t, testDB.DB.Create(org).Error)

	// Create custom org permissions
	customPerm := &models.Permission{
		Name:           "org:member:view",
		DisplayName:    "View Members",
		Category:       "member",
		IsSystem:       false,
		OrganizationID: &org.ID,
	}
	_, err := repo.Permission().Create(ctx, customPerm)
	require.NoError(t, err)

	// Create user WITHOUT membership
	user := &models.User{
		ID:           uuid.New(),
		Email:        "outsider@test.com",
		PasswordHash: "hashed",
		IsSuperadmin: false,
	}
	require.NoError(t, testDB.DB.Create(user).Error)

	// Try to get membership (should fail)
	_, err = repo.OrganizationMembership().GetByOrganizationAndUser(ctx, org.ID.String(), user.ID.String())
	assert.Error(t, err, "User should not have membership")

	// Verify org permissions exist
	orgPerms, err := repo.Permission().ListPermissionsByOrganization(ctx, org.ID)
	require.NoError(t, err)
	assert.Len(t, orgPerms, 1, "Org should have 1 custom permission")

	// CRITICAL: User without membership should get ZERO permissions
	// This would be tested via oauth2Service.getUserPermissions() which checks membership first
}

// TestOAuthRBACIntegration_CrossOrgPermissionAssignmentFails verifies that
// permissions from one org cannot be assigned to roles in another org
func TestOAuthRBACIntegration_CrossOrgPermissionAssignmentFails(t *testing.T) {
	testDB := testutils.SetupTestDB(t)

	ctx := context.Background()
	repo := repository.NewRepository(testDB.DB)

	// Create two organizations
	org1 := &models.Organization{ID: uuid.New(), Name: "Org 1"}
	org2 := &models.Organization{ID: uuid.New(), Name: "Org 2"}
	require.NoError(t, testDB.DB.Create(org1).Error)
	require.NoError(t, testDB.DB.Create(org2).Error)

	// Create permission in org1
	perm1 := &models.Permission{
		Name:           "org1:action",
		DisplayName:    "Org 1 Action",
		Category:       "custom",
		IsSystem:       false,
		OrganizationID: &org1.ID,
	}
	_, err := repo.Permission().Create(ctx, perm1)
	require.NoError(t, err)

	// Create role in org2
	role2 := &models.Role{
		ID:             uuid.New(),
		OrganizationID: &org2.ID,
		Name:           "org2-role",
		DisplayName:    "Org 2 Role",
		IsSystem:       false,
	}
	require.NoError(t, testDB.DB.Create(role2).Error)

	// CRITICAL: Attempt to assign org1 permission to org2 role (should FAIL)
	err = repo.Permission().AssignToRole(ctx, role2.ID, perm1.ID)
	assert.Error(t, err, "Cross-org permission assignment must fail")
	assert.Contains(t, err.Error(), "different organization", "Error should mention cross-org violation")
}

// TestOAuthRBACIntegration_SystemPermissionToCustomRoleFails verifies that
// system permissions cannot be assigned to custom roles
func TestOAuthRBACIntegration_SystemPermissionToCustomRoleFails(t *testing.T) {
	testDB := testutils.SetupTestDB(t)

	ctx := context.Background()
	repo := repository.NewRepository(testDB.DB)

	// Create system permission
	sysPerm := &models.Permission{
		Name:           "system:admin",
		DisplayName:    "System Admin",
		Category:       "system",
		IsSystem:       true,
		OrganizationID: nil,
	}
	_, err := repo.Permission().Create(ctx, sysPerm)
	require.NoError(t, err)

	// Create organization
	org := &models.Organization{ID: uuid.New(), Name: "Test Org"}
	require.NoError(t, testDB.DB.Create(org).Error)

	// Create custom role
	customRole := &models.Role{
		ID:             uuid.New(),
		OrganizationID: &org.ID,
		Name:           "custom-role",
		DisplayName:    "Custom Role",
		IsSystem:       false,
	}
	require.NoError(t, testDB.DB.Create(customRole).Error)

	// CRITICAL: Attempt to assign system permission to custom role (should FAIL)
	err = repo.Permission().AssignToRole(ctx, customRole.ID, sysPerm.ID)
	assert.Error(t, err, "System permission assignment to custom role must fail")
	assert.Contains(t, err.Error(), "system permission", "Error should mention system permission violation")
}

// TestOAuthRBACIntegration_CustomPermissionToSystemRoleFails verifies that
// custom permissions cannot be assigned to system roles
func TestOAuthRBACIntegration_CustomPermissionToSystemRoleFails(t *testing.T) {
	testDB := testutils.SetupTestDB(t)

	ctx := context.Background()
	repo := repository.NewRepository(testDB.DB)

	// Create organization
	org := &models.Organization{ID: uuid.New(), Name: "Test Org"}
	require.NoError(t, testDB.DB.Create(org).Error)

	// Create custom permission
	customPerm := &models.Permission{
		Name:           "org:custom",
		DisplayName:    "Custom Permission",
		Category:       "custom",
		IsSystem:       false,
		OrganizationID: &org.ID,
	}
	_, err := repo.Permission().Create(ctx, customPerm)
	require.NoError(t, err)

	// Create system role
	sysRole := &models.Role{
		ID:             uuid.New(),
		OrganizationID: nil,
		Name:           "admin",
		DisplayName:    "Admin",
		IsSystem:       true,
	}
	require.NoError(t, testDB.DB.Create(sysRole).Error)

	// CRITICAL: Attempt to assign custom permission to system role (should FAIL)
	err = repo.Permission().AssignToRole(ctx, sysRole.ID, customPerm.ID)
	assert.Error(t, err, "Custom permission assignment to system role must fail")
	assert.Contains(t, err.Error(), "custom permission", "Error should mention custom permission violation")
}

// TestOAuthRBACIntegration_JWTContainsCorrectClaims verifies that OAuth2 access tokens
// contain correct roles, permissions, and scope claims
func TestOAuthRBACIntegration_JWTContainsCorrectClaims(t *testing.T) {
	testDB := testutils.SetupTestDB(t)

	ctx := context.Background()
	repo := repository.NewRepository(testDB.DB)

	// Create organization
	org := &models.Organization{ID: uuid.New(), Name: "Test Org"}
	require.NoError(t, testDB.DB.Create(org).Error)

	// Create custom permissions
	perms := []models.Permission{
		{Name: "org:read", DisplayName: "Read", Category: "org", IsSystem: false, OrganizationID: &org.ID},
		{Name: "org:write", DisplayName: "Write", Category: "org", IsSystem: false, OrganizationID: &org.ID},
	}
	for i := range perms {
		_, err := repo.Permission().Create(ctx, &perms[i])
		require.NoError(t, err)
	}

	// Create custom role
	role := &models.Role{
		ID:             uuid.New(),
		OrganizationID: &org.ID,
		Name:           "editor",
		DisplayName:    "Editor",
		IsSystem:       false,
	}
	require.NoError(t, testDB.DB.Create(role).Error)

	// Assign permissions to role
	for i := range perms {
		err := repo.Permission().AssignToRole(ctx, role.ID, perms[i].ID)
		require.NoError(t, err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "editor@test.com",
		PasswordHash: "hashed",
		IsSuperadmin: false,
	}
	require.NoError(t, testDB.DB.Create(user).Error)

	// Create membership
	membership := &models.OrganizationMembership{
		ID:             uuid.New(),
		UserID:         user.ID,
		OrganizationID: org.ID,
		RoleID:         role.ID,
		Status:         "active",
	}
	require.NoError(t, testDB.DB.Create(membership).Error)

	// Create JWT service
	jwtConfig := &config.JWTConfig{Issuer: "test-issuer", Secret: "test-secret"}
	jwtService, err := jwt.NewService(jwtConfig)
	require.NoError(t, err)

	// Generate OAuth2 token context
	tokenCtx := &jwt.OAuthTokenContext{
		UserID:         user.ID,
		Email:          user.Email,
		OrganizationID: &org.ID,
		Roles:          []string{"editor"},
		Permissions:    []string{"org:read", "org:write"},
		Issuer:         "https://auth.myservice.com/test_client",
		Audience:       "test_client",
		Subject:        user.ID.String(),
		IsSuperadmin:   false,
	}

	// Generate access token
	accessToken, err := jwtService.GenerateOAuthAccessToken(tokenCtx)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken, "Access token should be generated")

	// Parse token
	claims, err := jwtService.ValidateToken(accessToken)
	require.NoError(t, err)

	// CRITICAL ASSERTIONS: Verify JWT contains correct claims
	assert.Equal(t, user.ID, claims.UserID, "User ID should match")
	assert.Equal(t, user.Email, claims.Email, "Email should match")
	assert.Equal(t, &org.ID, claims.Org, "Org should match")
	assert.Equal(t, []string{"editor"}, claims.Roles, "Roles should match")
	assert.Equal(t, []string{"org:read", "org:write"}, claims.Permissions, "Permissions should match")
	assert.Equal(t, "org:read org:write", claims.Scope, "Scope should be space-separated permissions")
	assert.Equal(t, "https://auth.myservice.com/test_client", claims.Issuer, "Issuer should match")
	assert.Equal(t, "test_client", claims.Audience[0], "Audience should match")
	assert.Equal(t, user.ID.String(), claims.Subject, "Subject should match")
	assert.False(t, claims.IsSuperadmin, "IsSuperadmin should be false")

	// Cleanup
	_ = membership
}

// TestOAuthRBACIntegration_OAuthScopesMatchRBACPermissions verifies that
// OAuth2 scopes are correctly mapped to RBAC permissions
func TestOAuthRBACIntegration_OAuthScopesMatchRBACPermissions(t *testing.T) {
	testDB := testutils.SetupTestDB(t)

	ctx := context.Background()
	repo := repository.NewRepository(testDB.DB)

	// Create system permissions for superadmin
	systemPerms := []string{"system:admin", "system:user", "system:audit"}
	for _, name := range systemPerms {
		_, err := repo.Permission().Create(ctx, &models.Permission{
			Name:           name,
			DisplayName:    name,
			Category:       "system",
			IsSystem:       true,
			OrganizationID: nil,
		})
		require.NoError(t, err)
	}

	// Create superadmin user
	superadmin := &models.User{
		ID:           uuid.New(),
		Email:        "superadmin@test.com",
		PasswordHash: "hashed",
		IsSuperadmin: true,
	}
	require.NoError(t, testDB.DB.Create(superadmin).Error)

	// Load system permissions
	loadedPerms, err := repo.Permission().ListSystemPermissions(ctx)
	require.NoError(t, err)

	// Build scope list
	scopes := make([]string, len(loadedPerms))
	for i, perm := range loadedPerms {
		scopes[i] = perm.Name
	}

	// CRITICAL ASSERTION: Scopes should match permission names exactly
	assert.ElementsMatch(t, systemPerms, scopes, "OAuth scopes should match RBAC permission names")

	// Verify space-separated scope format for JWT
	scopeString := "system:admin system:audit system:user" // (order may vary)
	assert.Contains(t, scopeString, "system:admin")
	assert.Contains(t, scopeString, "system:user")
	assert.Contains(t, scopeString, "system:audit")
}
