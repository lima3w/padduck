package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"ipam-next/models"
)

// newTestService returns a zero-value Service sufficient for pure RBAC logic
// (no DB or config fields are accessed by HasPermission, CanAccessResource,
// RequirePermission, or IsValidRole).
func newTestService() *Service {
	return &Service{}
}

// ---------------------------------------------------------------------------
// Permission constant existence and non-emptiness
// ---------------------------------------------------------------------------

func TestPermissionConstants(t *testing.T) {
	constants := map[string]string{
		"PermSectionCreate": PermSectionCreate,
		"PermSectionRead":   PermSectionRead,
		"PermSectionUpdate": PermSectionUpdate,
		"PermSectionDelete": PermSectionDelete,
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
	assert.Equal(t, "admin", RoleAdmin)
	assert.Equal(t, "user", RoleUser)
	assert.Equal(t, "viewer", RoleViewer)
}

// ---------------------------------------------------------------------------
// HasPermission
// ---------------------------------------------------------------------------

func TestHasPermission_NilUser(t *testing.T) {
	svc := newTestService()
	assert.False(t, svc.HasPermission(nil, PermSectionRead))
	assert.False(t, svc.HasPermission(nil, PermIPCreate))
	assert.False(t, svc.HasPermission(nil, "arbitrary:perm"))
}

func TestHasPermission_AdminGetsEveryPermission(t *testing.T) {
	svc := newTestService()
	admin := &models.User{Role: RoleAdmin}

	allPerms := []string{
		PermSectionCreate, PermSectionRead, PermSectionUpdate, PermSectionDelete,
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
		PermSectionRead, PermSectionCreate, PermSectionUpdate, PermSectionDelete,
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
	readPerms := []string{PermSectionRead, PermSubnetRead, PermIPRead, PermTokenRead}
	for _, perm := range readPerms {
		t.Run("granted_"+perm, func(t *testing.T) {
			assert.True(t, svc.HasPermission(viewer, perm),
				"viewer should have permission %s", perm)
		})
	}

	// viewer must not get write permissions
	deniedPerms := []string{
		PermSectionCreate, PermSectionUpdate, PermSectionDelete,
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
		PermSectionCreate, PermSectionRead, PermSubnetRead, PermIPRead, PermTokenRead,
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
	assert.False(t, svc.HasPermission(noRole, PermSectionRead))
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
		{"nil user denied", nil, PermSectionRead, false},
		{"admin granted", &models.User{Role: RoleAdmin}, PermIPDelete, true},
		{"viewer read granted", &models.User{Role: RoleViewer}, PermIPRead, true},
		{"viewer write denied", &models.User{Role: RoleViewer}, PermIPCreate, false},
		{"user granted", &models.User{Role: RoleUser}, PermSubnetCreate, true},
		{"unknown role denied", &models.User{Role: "ghost"}, PermSectionRead, false},
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
		err := svc.RequirePermission(admin, PermSectionCreate)
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
		err := svc.RequirePermission(nil, PermSectionCreate)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), PermSectionCreate),
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
		{"Admin", false},   // case-sensitive
		{"USER", false},    // case-sensitive
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
