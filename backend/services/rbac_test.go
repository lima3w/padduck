package services

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"padduck/models"
)

// newTestService returns a zero-value IdentityService sufficient for pure RBAC logic
// (no DB or config fields are accessed by HasPermission, CanAccessResource,
// RequirePermission, or IsValidRole).
func newTestService() *IdentityService {
	return &IdentityService{}
}

// ---------------------------------------------------------------------------
// Permission constant existence and non-emptiness
// ---------------------------------------------------------------------------

func TestPermissionConstants(t *testing.T) {
	t.Parallel()

	constants := map[string]string{
		"PermNetworkCreate": PermNetworkCreate,
		"PermNetworkRead":   PermNetworkRead,
		"PermNetworkUpdate": PermNetworkUpdate,
		"PermNetworkDelete": PermNetworkDelete,
		"PermSubnetCreate":  PermSubnetCreate,
		"PermSubnetRead":    PermSubnetRead,
		"PermSubnetUpdate":  PermSubnetUpdate,
		"PermSubnetDelete":  PermSubnetDelete,
		"PermIPCreate":      PermIPCreate,
		"PermIPRead":        PermIPRead,
		"PermIPUpdate":      PermIPUpdate,
		"PermIPDelete":      PermIPDelete,
		"PermIPAssign":      PermIPAssign,
		"PermIPRelease":     PermIPRelease,
		"PermTokenCreate":   PermTokenCreate,
		"PermTokenRead":     PermTokenRead,
		"PermTokenDelete":   PermTokenDelete,
	}

	for name, val := range constants {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, val, "constant %s should have a non-empty value", name)
		})
	}
}

func TestRoleConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "admin", RoleAdmin)
	assert.Equal(t, "user", RoleUser)
	assert.Equal(t, "viewer", RoleViewer)
}

// ---------------------------------------------------------------------------
// HasPermission
// ---------------------------------------------------------------------------

func TestHasPermission_NilUser(t *testing.T) {
	svc := newTestService()
	assert.False(t, svc.HasPermission(nil, PermNetworkRead))
	assert.False(t, svc.HasPermission(nil, PermIPCreate))
	assert.False(t, svc.HasPermission(nil, "arbitrary:perm"))
}

func TestHasPermission_AdminGetsEveryPermission(t *testing.T) {
	svc := newTestService()
	admin := &models.User{Role: RoleAdmin}

	allPerms := []string{
		PermNetworkCreate, PermNetworkRead, PermNetworkUpdate, PermNetworkDelete,
		PermSubnetCreate, PermSubnetRead, PermSubnetUpdate, PermSubnetDelete,
		PermIPCreate, PermIPRead, PermIPUpdate, PermIPDelete, PermIPAssign, PermIPRelease,
		PermTokenCreate, PermTokenRead, PermTokenDelete,
	}

	for _, perm := range allPerms {
		t.Run(perm, func(t *testing.T) {
			assert.True(t, svc.HasPermission(admin, perm),
				"admin should have permission %s", perm)
		})
	}
}

func TestHasPermission_UserRole(t *testing.T) {
	svc := newTestService()
	user := &models.User{Role: RoleUser}

	// user role has all permissions
	granted := []string{
		PermNetworkRead, PermNetworkCreate, PermNetworkUpdate, PermNetworkDelete,
		PermSubnetRead, PermSubnetCreate, PermSubnetUpdate, PermSubnetDelete,
		PermIPRead, PermIPCreate, PermIPUpdate, PermIPDelete, PermIPAssign, PermIPRelease,
		PermTokenCreate, PermTokenRead, PermTokenDelete,
	}
	for _, perm := range granted {
		t.Run("granted_"+perm, func(t *testing.T) {
			assert.True(t, svc.HasPermission(user, perm),
				"user role should have permission %s", perm)
		})
	}

	// Should not have an invented permission
	t.Run("denied_arbitrary", func(t *testing.T) {
		assert.False(t, svc.HasPermission(user, "vlan:delete"))
	})
}

func TestHasPermission_ViewerRole(t *testing.T) {
	svc := newTestService()
	viewer := &models.User{Role: RoleViewer}

	// viewer only gets read permissions
	readPerms := []string{PermNetworkRead, PermSubnetRead, PermIPRead, PermTokenRead}
	for _, perm := range readPerms {
		t.Run("granted_"+perm, func(t *testing.T) {
			assert.True(t, svc.HasPermission(viewer, perm),
				"viewer should have permission %s", perm)
		})
	}

	// viewer must not get write permissions
	deniedPerms := []string{
		PermNetworkCreate, PermNetworkUpdate, PermNetworkDelete,
		PermSubnetCreate, PermSubnetUpdate, PermSubnetDelete,
		PermIPCreate, PermIPUpdate, PermIPDelete, PermIPAssign, PermIPRelease,
		PermTokenCreate, PermTokenDelete,
	}
	for _, perm := range deniedPerms {
		t.Run("denied_"+perm, func(t *testing.T) {
			assert.False(t, svc.HasPermission(viewer, perm),
				"viewer should NOT have permission %s", perm)
		})
	}
}

func TestHasPermission_UnknownRole(t *testing.T) {
	svc := newTestService()
	unknown := &models.User{Role: "superuser"}

	perms := []string{
		PermNetworkCreate, PermNetworkRead, PermSubnetRead, PermIPRead, PermTokenRead,
	}
	for _, perm := range perms {
		t.Run(perm, func(t *testing.T) {
			assert.False(t, svc.HasPermission(unknown, perm),
				"unknown role should have no permissions, but got %s", perm)
		})
	}
}

func TestHasPermission_EmptyRole(t *testing.T) {
	svc := newTestService()
	noRole := &models.User{Role: ""}
	assert.False(t, svc.HasPermission(noRole, PermNetworkRead))
}

// ---------------------------------------------------------------------------
// CanAccessResource — delegates to HasPermission
// ---------------------------------------------------------------------------

func TestCanAccessResource_DelegatesToHasPermission(t *testing.T) {
	svc := newTestService()

	cases := []struct {
		name   string
		user   *models.User
		action string
		want   bool
	}{
		{"nil user denied", nil, PermNetworkRead, false},
		{"admin granted", &models.User{Role: RoleAdmin}, PermIPDelete, true},
		{"viewer read granted", &models.User{Role: RoleViewer}, PermIPRead, true},
		{"viewer write denied", &models.User{Role: RoleViewer}, PermIPCreate, false},
		{"user granted", &models.User{Role: RoleUser}, PermSubnetCreate, true},
		{"unknown role denied", &models.User{Role: "ghost"}, PermNetworkRead, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := svc.CanAccessResource(tc.user, tc.action)
			assert.Equal(t, tc.want, got)
			// Confirm it matches HasPermission directly
			assert.Equal(t, svc.HasPermission(tc.user, tc.action), got)
		})
	}
}

// ---------------------------------------------------------------------------
// RequirePermission
// ---------------------------------------------------------------------------

func TestRequirePermission(t *testing.T) {
	svc := newTestService()

	t.Run("returns nil when user has permission", func(t *testing.T) {
		admin := &models.User{Role: RoleAdmin}
		err := svc.RequirePermission(admin, PermNetworkCreate)
		assert.NoError(t, err)
	})

	t.Run("returns nil when viewer has read permission", func(t *testing.T) {
		viewer := &models.User{Role: RoleViewer}
		err := svc.RequirePermission(viewer, PermIPRead)
		assert.NoError(t, err)
	})

	t.Run("error contains permission name when denied", func(t *testing.T) {
		viewer := &models.User{Role: RoleViewer}
		err := svc.RequirePermission(viewer, PermIPDelete)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), PermIPDelete),
			"error message should contain the missing permission name, got: %s", err.Error())
	})

	t.Run("error for nil user contains permission name", func(t *testing.T) {
		err := svc.RequirePermission(nil, PermNetworkCreate)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), PermNetworkCreate),
			"error message should contain the missing permission name, got: %s", err.Error())
	})

	t.Run("error for unknown role contains permission name", func(t *testing.T) {
		unknown := &models.User{Role: "unknown"}
		perm := PermSubnetUpdate
		err := svc.RequirePermission(unknown, perm)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), perm)
	})
}

// ---------------------------------------------------------------------------
// IsValidRole
// ---------------------------------------------------------------------------

func TestIsValidRole(t *testing.T) {
	cases := []struct {
		role  string
		valid bool
	}{
		{RoleAdmin, true},
		{RoleUser, true},
		{RoleViewer, true},
		{"superuser", false},
		{"Admin", false}, // case-sensitive
		{"USER", false},  // case-sensitive
		{"", false},
		{"root", false},
		{"moderator", false},
	}

	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			assert.Equal(t, tc.valid, IsValidRole(tc.role),
				"IsValidRole(%q) should be %v", tc.role, tc.valid)
		})
	}
}

// ---------------------------------------------------------------------------
// v0.8.11 — IsValidPermission
// ---------------------------------------------------------------------------

func TestIsValidPermission_AllKnownPermissions(t *testing.T) {
	for _, p := range AllPermissions {
		t.Run(p, func(t *testing.T) {
			assert.True(t, IsValidPermission(p), "known permission %q should be valid", p)
		})
	}
}

func TestIsValidPermission_UnknownPermissions(t *testing.T) {
	unknown := []string{
		"",
		"admin",
		"section:read",
		"ipam:section:*",
		"ipam:section",
		"IPAM:SECTION:READ",
		"ipam:unknownresource:read",
	}
	for _, p := range unknown {
		t.Run(p+"_invalid", func(t *testing.T) {
			assert.False(t, IsValidPermission(p), "unknown permission %q should be invalid", p)
		})
	}
}

func TestAllPermissions_ContainsExpectedCount(t *testing.T) {
	expected := []string{
		PermV2NetworkList, PermV2NetworkRead, PermV2NetworkWrite, PermV2NetworkDelete,
		PermV2SubnetList, PermV2SubnetRead, PermV2SubnetWrite, PermV2SubnetDelete,
		PermV2IPList, PermV2IPRead, PermV2IPAssign, PermV2IPRelease,
		PermV2VRFList, PermV2VRFRead, PermV2VRFWrite, PermV2VRFDelete,
		PermV2VLANList, PermV2VLANRead, PermV2VLANWrite, PermV2VLANDelete,
		PermV2UserList, PermV2UserRead, PermV2UserWrite, PermV2AuditRead,
		// v1.3.0 device permissions
		PermV2DeviceRead, PermV2DeviceWrite, PermV2DeviceDelete, PermV2DeviceAdmin,
		// v1.5.1 location permissions
		PermV2LocationList, PermV2LocationRead, PermV2LocationWrite, PermV2LocationDelete,
		// v1.6.0 nameserver permissions
		PermV2NameserverList, PermV2NameserverRead, PermV2NameserverWrite, PermV2NameserverDelete,
		// v1.7.0 request workflow permissions
		PermV2SubnetRequestSubmit, PermV2SubnetRequestReview,
		// v1.8.0 vlan domain permissions
		PermV2VLANDomainList, PermV2VLANDomainRead, PermV2VLANDomainWrite, PermV2VLANDomainDelete,
		// v1.8.0 vlan group permissions
		PermV2VLANGroupList, PermV2VLANGroupRead, PermV2VLANGroupWrite, PermV2VLANGroupDelete,
		// admin-only operation permissions
		PermV2AdminRead, PermV2AdminWrite,
		// v1.14.0 customer / tenant management permissions
		PermV2CustomerList, PermV2CustomerRead, PermV2CustomerWrite, PermV2CustomerDelete,
		// v1.14.0 BGP autonomous system permissions
		PermV2ASList, PermV2ASRead, PermV2ASWrite, PermV2ASDelete,
		// v1.29.0 network module permissions
		PermV2NATList, PermV2NATRead, PermV2NATWrite, PermV2NATDelete,
		PermV2DHCPList, PermV2DHCPRead, PermV2DHCPWrite, PermV2DHCPDelete,
		PermV2CircuitList, PermV2CircuitRead, PermV2CircuitWrite, PermV2CircuitDelete,
		// v1.30.0 optional tools permissions
		PermV2FirewallList, PermV2FirewallRead, PermV2FirewallWrite, PermV2FirewallDelete,
		// v1.33.12 organization permissions
		PermV2OrgRead, PermV2OrgWrite,
		// v1.33.14 platform admin permission
		PermV2PlatformAdmin,
	}
	assert.Equal(t, len(expected), len(AllPermissions))
	for _, p := range expected {
		assert.Contains(t, AllPermissions, p)
	}
}

// ---------------------------------------------------------------------------
// v0.8.11 — CheckPermission (validation guards only, no DB)
// ---------------------------------------------------------------------------

func TestCheckPermission_InvalidUserID(t *testing.T) {
	t.Parallel()

	svc := newTestService()
	ctx := context.Background()

	for _, id := range []int64{0, -1, -99} {
		err := svc.CheckPermission(ctx, id, PermV2NetworkRead)
		assert.Error(t, err, "userID %d should be rejected", id)
	}
}

func TestPermMatches_ResourceScopedPermissions(t *testing.T) {
	t.Parallel()

	resourceType := "subnet"
	resourceID := int64(42)
	otherResourceID := int64(99)

	cases := []struct {
		name   string
		perms  []*models.RolePermission
		scopes []ResourceScope
		want   bool
	}{
		{
			name: "global grant matches any resource",
			perms: []*models.RolePermission{{
				Permission: PermV2SubnetRead,
			}},
			scopes: []ResourceScope{{Type: "subnet", ID: otherResourceID}},
			want:   true,
		},
		{
			name: "type only grant does not wildcard resource type",
			perms: []*models.RolePermission{{
				Permission:   PermV2SubnetRead,
				ResourceType: &resourceType,
				ResourceID:   nil,
			}},
			scopes: []ResourceScope{{Type: "subnet", ID: resourceID}},
			want:   false,
		},
		{
			name: "exact resource grant matches",
			perms: []*models.RolePermission{{
				Permission:   PermV2SubnetRead,
				ResourceType: &resourceType,
				ResourceID:   &resourceID,
			}},
			scopes: []ResourceScope{{Type: "subnet", ID: resourceID}},
			want:   true,
		},
		{
			name: "different resource id does not match",
			perms: []*models.RolePermission{{
				Permission:   PermV2SubnetRead,
				ResourceType: &resourceType,
				ResourceID:   &resourceID,
			}},
			scopes: []ResourceScope{{Type: "subnet", ID: otherResourceID}},
			want:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, permMatches(tc.perms, PermV2SubnetRead, tc.scopes))
		})
	}
}

// ---------------------------------------------------------------------------
// v0.8.11 — CreateRole validation
// ---------------------------------------------------------------------------

func TestCreateRole_EmptyName(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	_, err := svc.Ops.Identity.CreateRole(ctx, "", "some description")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role name is required")
}

func TestCreateRole_InvalidNameChars(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	invalid := []string{"role name", "role.name", "role/name", "role@name"}
	for _, name := range invalid {
		t.Run(name, func(t *testing.T) {
			_, err := svc.Ops.Identity.CreateRole(ctx, name, "")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "letters, numbers")
		})
	}
}

func TestCreateRole_ValidName_ReachesRepo(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	assert.Panics(t, func() {
		_, _ = svc.Ops.Identity.CreateRole(ctx, "my-role", "description")
	})
}

// ---------------------------------------------------------------------------
// v0.8.11 — GetRole validation
// ---------------------------------------------------------------------------

func TestGetRole_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	for _, id := range []int64{0, -1, -99} {
		_, err := svc.Ops.Identity.GetRole(ctx, id)
		assert.Error(t, err, "id %d should be rejected", id)
		assert.Contains(t, err.Error(), "invalid role ID")
	}
}

// ---------------------------------------------------------------------------
// v0.8.11 — UpdateRole validation
// ---------------------------------------------------------------------------

func TestUpdateRole_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	_, err := svc.Ops.Identity.UpdateRole(ctx, 0, "valid-name", "desc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role ID")
}

func TestUpdateRole_EmptyName(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	_, err := svc.Ops.Identity.UpdateRole(ctx, 1, "", "desc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role name is required")
}

// ---------------------------------------------------------------------------
// v0.8.11 — DeleteRole validation
// ---------------------------------------------------------------------------

func TestDeleteRole_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	for _, id := range []int64{0, -1} {
		err := svc.Ops.Identity.DeleteRole(ctx, id)
		assert.Error(t, err, "id %d should be rejected", id)
		assert.Contains(t, err.Error(), "invalid role ID")
	}
}

// ---------------------------------------------------------------------------
// v0.8.11 — AddPermissionToRole validation
// ---------------------------------------------------------------------------

func TestAddPermissionToRole_InvalidRoleID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	_, err := svc.Ops.Identity.AddPermissionToRole(ctx, 0, PermV2NetworkRead, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role ID")
}

func TestAddPermissionToRole_UnknownPermission(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	_, err := svc.Ops.Identity.AddPermissionToRole(ctx, 1, "not:a:real:permission", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown permission")
}

func TestAddPermissionToRole_ResourceTypeRequiresResourceID(t *testing.T) {
	t.Parallel()

	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	resourceType := "subnet"
	_, err := svc.Ops.Identity.AddPermissionToRole(ctx, 1, PermV2SubnetRead, &resourceType, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource ID is required")
}

func TestAddPermissionToRole_ValidArgs_ReachesRepo(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	assert.Panics(t, func() {
		_, _ = svc.Ops.Identity.AddPermissionToRole(ctx, 1, PermV2NetworkRead, nil, nil)
	})
}

// ---------------------------------------------------------------------------
// v0.8.11 — RemovePermissionFromRole validation
// ---------------------------------------------------------------------------

func TestRemovePermissionFromRole_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	for _, id := range []int64{0, -1} {
		err := svc.Ops.Identity.RemovePermissionFromRole(ctx, id)
		assert.Error(t, err, "id %d should be rejected", id)
		assert.Contains(t, err.Error(), "invalid permission ID")
	}
}

// ---------------------------------------------------------------------------
// v0.8.11 — AssignRoleToUser / RemoveRoleFromUser / GetUserRoles validation
// ---------------------------------------------------------------------------

func TestAssignRoleToUser_InvalidUserID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	err := svc.Ops.Identity.AssignRoleToUser(ctx, 0, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID")
}

func TestAssignRoleToUser_InvalidRoleID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	err := svc.Ops.Identity.AssignRoleToUser(ctx, 1, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role ID")
}

func TestRemoveRoleFromUser_InvalidUserID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	err := svc.Ops.Identity.RemoveRoleFromUser(ctx, 0, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID")
}

func TestRemoveRoleFromUser_InvalidRoleID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	err := svc.Ops.Identity.RemoveRoleFromUser(ctx, 1, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role ID")
}

func TestGetUserRoles_InvalidUserID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()
	for _, id := range []int64{0, -1} {
		_, err := svc.Ops.Identity.GetUserRoles(ctx, id)
		assert.Error(t, err, "id %d should be rejected", id)
		assert.Contains(t, err.Error(), "invalid user ID")
	}
}

// ---------------------------------------------------------------------------
// v0.8.11 — v2 permission constant values
// ---------------------------------------------------------------------------

func TestV2PermissionConstants_NonEmpty(t *testing.T) {
	v2perms := map[string]string{
		"PermV2NetworkList":   PermV2NetworkList,
		"PermV2NetworkRead":   PermV2NetworkRead,
		"PermV2NetworkWrite":  PermV2NetworkWrite,
		"PermV2NetworkDelete": PermV2NetworkDelete,
		"PermV2SubnetList":    PermV2SubnetList,
		"PermV2SubnetRead":    PermV2SubnetRead,
		"PermV2SubnetWrite":   PermV2SubnetWrite,
		"PermV2SubnetDelete":  PermV2SubnetDelete,
		"PermV2IPList":        PermV2IPList,
		"PermV2IPRead":        PermV2IPRead,
		"PermV2IPAssign":      PermV2IPAssign,
		"PermV2IPRelease":     PermV2IPRelease,
		"PermV2UserList":      PermV2UserList,
		"PermV2UserRead":      PermV2UserRead,
		"PermV2UserWrite":     PermV2UserWrite,
		"PermV2AuditRead":     PermV2AuditRead,
	}
	for name, val := range v2perms {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, val, "constant %s should have a non-empty value", name)
		})
	}
}

func TestV2PermissionConstants_Prefixed(t *testing.T) {
	for _, p := range AllPermissions {
		assert.True(t,
			strings.HasPrefix(p, "ipam:") || strings.HasPrefix(p, "auth:") || strings.HasPrefix(p, "devices:"),
			"permission %q should start with 'ipam:', 'auth:', or 'devices:'", p,
		)
	}
}

func TestLegacyUserRole_DoesNotGrantCustomerOrASMutation(t *testing.T) {
	denied := []string{
		PermV2CustomerWrite,
		PermV2CustomerDelete,
		PermV2ASWrite,
		PermV2ASDelete,
		PermV2NATWrite,
		PermV2NATDelete,
		PermV2DHCPWrite,
		PermV2DHCPDelete,
		PermV2CircuitWrite,
		PermV2CircuitDelete,
		PermV2FirewallWrite,
		PermV2FirewallDelete,
	}
	for _, perm := range denied {
		t.Run(perm, func(t *testing.T) {
			assert.False(t, legacyRoleHasPermission(RoleUser, perm))
		})
	}
}

func TestLegacyUserRole_GrantsCustomerAndASRead(t *testing.T) {
	granted := []string{
		PermV2CustomerList,
		PermV2CustomerRead,
		PermV2ASList,
		PermV2ASRead,
		PermV2NATList,
		PermV2NATRead,
		PermV2DHCPList,
		PermV2DHCPRead,
		PermV2CircuitList,
		PermV2CircuitRead,
		PermV2FirewallList,
		PermV2FirewallRead,
	}
	for _, perm := range granted {
		t.Run(perm, func(t *testing.T) {
			assert.True(t, legacyRoleHasPermission(RoleUser, perm))
		})
	}
}

func TestLegacyViewerRole_GrantsCustomerAndASRead(t *testing.T) {
	granted := []string{
		PermV2CustomerList,
		PermV2CustomerRead,
		PermV2ASList,
		PermV2ASRead,
		PermV2NATList,
		PermV2NATRead,
		PermV2DHCPList,
		PermV2DHCPRead,
		PermV2CircuitList,
		PermV2CircuitRead,
		PermV2FirewallList,
		PermV2FirewallRead,
	}
	for _, perm := range granted {
		t.Run(perm, func(t *testing.T) {
			assert.True(t, legacyRoleHasPermission(RoleViewer, perm))
		})
	}
}
