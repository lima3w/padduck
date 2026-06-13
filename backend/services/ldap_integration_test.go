package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/internal/testdb"
	"padduck/models"
	"padduck/repository"
)

func testLDAPSvc(t *testing.T) (*LDAPService, *repository.Repository) {
	t.Helper()
	pool := testdb.Connect(t, "services")
	testdb.Truncate(t, pool, "ldap_group_role_mappings", "user_roles", "users")
	repo := repository.NewRepository(pool)
	return NewLDAPService(repo, testMFAKey), repo
}

func TestSyncGroups_AssignsMatchingRole_Integration(t *testing.T) {
	svc, repo := testLDAPSvc(t)
	ctx := context.Background()

	role, err := repo.CreateRole(ctx, "sync-test-role", "test", false)
	require.NoError(t, err)
	user, err := repo.CreateUser(ctx, "sync-user", "sync@example.com")
	require.NoError(t, err)
	require.NoError(t, repo.CreateLDAPGroupMapping(ctx, &models.LDAPGroupRoleMapping{
		LDAPGroupDN: "CN=Engineers,DC=example,DC=com",
		RoleID:      role.ID,
	}))

	require.NoError(t, svc.SyncGroups(ctx, user.ID, []string{
		"CN=Engineers,DC=example,DC=com",
	}))

	roles, err := repo.GetUserRoles(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	assert.Equal(t, role.ID, roles[0].ID)
}

func TestSyncGroups_NonMatchingGroupsAssignNothing_Integration(t *testing.T) {
	svc, repo := testLDAPSvc(t)
	ctx := context.Background()

	role, err := repo.CreateRole(ctx, "no-match-role", "test", false)
	require.NoError(t, err)
	user, err := repo.CreateUser(ctx, "no-match-user", "nomatch@example.com")
	require.NoError(t, err)
	require.NoError(t, repo.CreateLDAPGroupMapping(ctx, &models.LDAPGroupRoleMapping{
		LDAPGroupDN: "CN=Admins,DC=example,DC=com",
		RoleID:      role.ID,
	}))

	require.NoError(t, svc.SyncGroups(ctx, user.ID, []string{
		"CN=Developers,DC=example,DC=com",
	}))

	roles, err := repo.GetUserRoles(ctx, user.ID)
	require.NoError(t, err)
	assert.Empty(t, roles)
}

func TestSyncGroups_IdempotentOnRepeat_Integration(t *testing.T) {
	svc, repo := testLDAPSvc(t)
	ctx := context.Background()

	role, err := repo.CreateRole(ctx, "idem-role", "test", false)
	require.NoError(t, err)
	user, err := repo.CreateUser(ctx, "idem-user", "idem@example.com")
	require.NoError(t, err)
	require.NoError(t, repo.CreateLDAPGroupMapping(ctx, &models.LDAPGroupRoleMapping{
		LDAPGroupDN: "CN=Staff,DC=example,DC=com",
		RoleID:      role.ID,
	}))

	groups := []string{"CN=Staff,DC=example,DC=com"}

	// First call assigns the role.
	require.NoError(t, svc.SyncGroups(ctx, user.ID, groups))
	// Second call is a no-op (ON CONFLICT DO NOTHING); must not error.
	require.NoError(t, svc.SyncGroups(ctx, user.ID, groups))

	roles, err := repo.GetUserRoles(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 1, "role must not be duplicated on repeated SyncGroups")
}

func TestSyncGroups_CaseInsensitiveGroupMatch_Integration(t *testing.T) {
	svc, repo := testLDAPSvc(t)
	ctx := context.Background()

	role, err := repo.CreateRole(ctx, "case-role", "test", false)
	require.NoError(t, err)
	user, err := repo.CreateUser(ctx, "case-user", "case@example.com")
	require.NoError(t, err)
	require.NoError(t, repo.CreateLDAPGroupMapping(ctx, &models.LDAPGroupRoleMapping{
		LDAPGroupDN: "CN=Staff,DC=example,DC=com",
		RoleID:      role.ID,
	}))

	// Group DN arrives with different casing from the LDAP server.
	require.NoError(t, svc.SyncGroups(ctx, user.ID, []string{
		"cn=staff,dc=example,dc=com",
	}))

	roles, err := repo.GetUserRoles(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 1)
}

func TestSyncGroups_EmptyGroups_Integration(t *testing.T) {
	svc, repo := testLDAPSvc(t)
	ctx := context.Background()

	user, err := repo.CreateUser(ctx, "empty-groups-user", "empty@example.com")
	require.NoError(t, err)

	require.NoError(t, svc.SyncGroups(ctx, user.ID, []string{}))

	roles, err := repo.GetUserRoles(ctx, user.ID)
	require.NoError(t, err)
	assert.Empty(t, roles)
}
