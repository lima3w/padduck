package services

import (
	"context"
	"fmt"

	"ipam-next/models"
)

// ---- Legacy permission constants (kept for backward compatibility) ----

const (
	RoleAdmin  = "admin"
	RoleUser   = "user"
	RoleViewer = "viewer"
)

var ValidRoles = []string{RoleAdmin, RoleUser, RoleViewer}

// Legacy permission strings (pre-v0.8.11, kept for tests)
const (
	PermSectionCreate = "section:create"
	PermSectionRead   = "section:read"
	PermSectionUpdate = "section:update"
	PermSectionDelete = "section:delete"
	PermSubnetCreate  = "subnet:create"
	PermSubnetRead    = "subnet:read"
	PermSubnetUpdate  = "subnet:update"
	PermSubnetDelete  = "subnet:delete"
	PermIPCreate      = "ip:create"
	PermIPRead        = "ip:read"
	PermIPUpdate      = "ip:update"
	PermIPDelete      = "ip:delete"
	PermIPAssign      = "ip:assign"
	PermIPRelease     = "ip:release"
	PermTokenCreate   = "token:create"
	PermTokenRead     = "token:read"
	PermTokenDelete   = "token:delete"
)

// ---- v0.8.11 permission constants ----

const (
	PermV2SectionList   = "ipam:section:list"
	PermV2SectionRead   = "ipam:section:read"
	PermV2SectionWrite  = "ipam:section:write"
	PermV2SectionDelete = "ipam:section:delete"

	PermV2SubnetList   = "ipam:subnet:list"
	PermV2SubnetRead   = "ipam:subnet:read"
	PermV2SubnetWrite  = "ipam:subnet:write"
	PermV2SubnetDelete = "ipam:subnet:delete"

	PermV2IPList    = "ipam:ip_address:list"
	PermV2IPRead    = "ipam:ip_address:read"
	PermV2IPAssign  = "ipam:ip_address:assign"
	PermV2IPRelease = "ipam:ip_address:release"

	PermV2VRFList   = "ipam:vrf:list"
	PermV2VRFRead   = "ipam:vrf:read"
	PermV2VRFWrite  = "ipam:vrf:write"
	PermV2VRFDelete = "ipam:vrf:delete"

	PermV2VLANList   = "ipam:vlan:list"
	PermV2VLANRead   = "ipam:vlan:read"
	PermV2VLANWrite  = "ipam:vlan:write"
	PermV2VLANDelete = "ipam:vlan:delete"

	PermV2UserList  = "auth:user:list"
	PermV2UserRead  = "auth:user:read"
	PermV2UserWrite = "auth:user:write"
	PermV2AuditRead = "auth:audit:read"

	PermV2DeviceRead   = "devices:read"
	PermV2DeviceWrite  = "devices:write"
	PermV2DeviceDelete = "devices:delete"
	PermV2DeviceAdmin  = "devices:admin"
)

// AllPermissions is the authoritative list of valid permission strings.
var AllPermissions = []string{
	PermV2SectionList, PermV2SectionRead, PermV2SectionWrite, PermV2SectionDelete,
	PermV2SubnetList, PermV2SubnetRead, PermV2SubnetWrite, PermV2SubnetDelete,
	PermV2IPList, PermV2IPRead, PermV2IPAssign, PermV2IPRelease,
	PermV2VRFList, PermV2VRFRead, PermV2VRFWrite, PermV2VRFDelete,
	PermV2VLANList, PermV2VLANRead, PermV2VLANWrite, PermV2VLANDelete,
	PermV2UserList, PermV2UserRead, PermV2UserWrite, PermV2AuditRead,
	PermV2DeviceRead, PermV2DeviceWrite, PermV2DeviceDelete, PermV2DeviceAdmin,
}

// IsValidPermission returns true if the given string is a known permission.
func IsValidPermission(p string) bool {
	for _, v := range AllPermissions {
		if v == p {
			return true
		}
	}
	return false
}

// ResourceScope identifies a resource that a permission check should be scoped to.
type ResourceScope struct {
	Type string
	ID   int64
}

// CheckPermission returns nil if userID has the given permission (optionally scoped
// to any of the provided resources). Falls back to the legacy role column when the
// user has no assigned roles yet.
func (s *Service) CheckPermission(ctx context.Context, userID int64, permission string, scopes ...ResourceScope) error {
	if userID <= 0 {
		return fmt.Errorf("permission denied")
	}

	count, err := s.repository.CountUserRoles(ctx, userID)
	if err == nil && count > 0 {
		perms, err := s.repository.GetUserPermissions(ctx, userID)
		if err != nil {
			return fmt.Errorf("permission denied")
		}
		if permMatches(perms, permission, scopes) {
			return nil
		}
		return fmt.Errorf("permission denied: %s", permission)
	}

	// Legacy fallback: use the role column
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("permission denied")
	}
	if legacyRoleHasPermission(user.Role, permission) {
		return nil
	}
	return fmt.Errorf("permission denied: %s", permission)
}

// permMatches returns true if any permission in perms satisfies the request.
func permMatches(perms []*models.RolePermission, permission string, scopes []ResourceScope) bool {
	for _, p := range perms {
		if p.Permission != permission {
			continue
		}
		if p.ResourceType == nil {
			return true // global grant
		}
		for _, s := range scopes {
			if *p.ResourceType == s.Type && (p.ResourceID == nil || *p.ResourceID == s.ID) {
				return true
			}
		}
	}
	return false
}

// legacyRoleHasPermission maps old role strings to new-style permission strings.
func legacyRoleHasPermission(role, permission string) bool {
	switch role {
	case "admin":
		return true // admin has everything
	case "user":
		adminOnly := map[string]bool{PermV2UserWrite: true, PermV2AuditRead: true, PermV2DeviceAdmin: true}
		return !adminOnly[permission]
	case "viewer":
		readPerms := map[string]bool{
			PermV2SectionList: true, PermV2SectionRead: true,
			PermV2SubnetList: true, PermV2SubnetRead: true,
			PermV2IPList: true, PermV2IPRead: true,
			PermV2VRFList: true, PermV2VRFRead: true,
			PermV2VLANList: true, PermV2VLANRead: true,
			PermV2UserList: true, PermV2UserRead: true,
			PermV2DeviceRead: true,
		}
		return readPerms[permission]
	}
	return false
}

// ---- Legacy RBAC (kept for backward compatibility with existing tests) ----

func (s *Service) HasPermission(user *models.User, permission string) bool {
	if user == nil {
		return false
	}
	if user.Role == RoleAdmin {
		return true
	}
	for _, p := range getRolePermissions(user.Role) {
		if p == permission {
			return true
		}
	}
	return false
}

func (s *Service) CanAccessResource(user *models.User, action string) bool {
	return s.HasPermission(user, action)
}

func (s *Service) RequirePermission(user *models.User, permission string) error {
	if !s.HasPermission(user, permission) {
		return fmt.Errorf("user does not have permission: %s", permission)
	}
	return nil
}

func getRolePermissions(role string) []string {
	switch role {
	case RoleAdmin:
		return []string{
			PermSectionCreate, PermSectionRead, PermSectionUpdate, PermSectionDelete,
			PermSubnetCreate, PermSubnetRead, PermSubnetUpdate, PermSubnetDelete,
			PermIPCreate, PermIPRead, PermIPUpdate, PermIPDelete, PermIPAssign, PermIPRelease,
			PermTokenCreate, PermTokenRead, PermTokenDelete,
		}
	case RoleUser:
		return []string{
			PermSectionRead, PermSectionCreate, PermSectionUpdate, PermSectionDelete,
			PermSubnetRead, PermSubnetCreate, PermSubnetUpdate, PermSubnetDelete,
			PermIPRead, PermIPCreate, PermIPUpdate, PermIPDelete, PermIPAssign, PermIPRelease,
			PermTokenCreate, PermTokenRead, PermTokenDelete,
		}
	case RoleViewer:
		return []string{PermSectionRead, PermSubnetRead, PermIPRead, PermTokenRead}
	}
	return nil
}

func IsValidRole(role string) bool {
	for _, v := range ValidRoles {
		if v == role {
			return true
		}
	}
	return false
}

// ---- Role management (v0.8.11) ----

func (s *Service) CreateRole(ctx context.Context, name, description string) (*models.Role, error) {
	if name == "" {
		return nil, fmt.Errorf("role name is required")
	}
	if !isValidRoleName(name) {
		return nil, fmt.Errorf("role name may only contain letters, numbers, hyphens, and underscores")
	}
	return s.repository.CreateRole(ctx, name, description, false)
}

func (s *Service) GetRole(ctx context.Context, id int64) (*models.Role, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid role ID")
	}
	return s.repository.GetRoleByID(ctx, id)
}

func (s *Service) ListRoles(ctx context.Context) ([]*models.Role, error) {
	roles, err := s.repository.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	if roles == nil {
		roles = []*models.Role{}
	}
	return roles, nil
}

func (s *Service) UpdateRole(ctx context.Context, id int64, name, description string) (*models.Role, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid role ID")
	}
	if name == "" {
		return nil, fmt.Errorf("role name is required")
	}
	return s.repository.UpdateRole(ctx, id, name, description)
}

func (s *Service) DeleteRole(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid role ID")
	}
	return s.repository.DeleteRole(ctx, id)
}

func (s *Service) AddPermissionToRole(ctx context.Context, roleID int64, permission string, resourceType *string, resourceID *int64) (*models.RolePermission, error) {
	if roleID <= 0 {
		return nil, fmt.Errorf("invalid role ID")
	}
	if !IsValidPermission(permission) {
		return nil, fmt.Errorf("unknown permission: %s", permission)
	}
	return s.repository.AddPermissionToRole(ctx, roleID, permission, resourceType, resourceID)
}

func (s *Service) RemovePermissionFromRole(ctx context.Context, permissionID int64) error {
	if permissionID <= 0 {
		return fmt.Errorf("invalid permission ID")
	}
	return s.repository.RemovePermissionFromRole(ctx, permissionID)
}

func (s *Service) GetUserRoles(ctx context.Context, userID int64) ([]*models.Role, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	roles, err := s.repository.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	if roles == nil {
		roles = []*models.Role{}
	}
	return roles, nil
}

func (s *Service) AssignRoleToUser(ctx context.Context, userID, roleID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	if roleID <= 0 {
		return fmt.Errorf("invalid role ID")
	}
	return s.repository.AssignRoleToUser(ctx, userID, roleID)
}

func (s *Service) RemoveRoleFromUser(ctx context.Context, userID, roleID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	if roleID <= 0 {
		return fmt.Errorf("invalid role ID")
	}
	return s.repository.RemoveRoleFromUser(ctx, userID, roleID)
}

func isValidRoleName(name string) bool {
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}
