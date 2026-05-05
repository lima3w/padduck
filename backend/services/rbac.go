package services

import (
	"fmt"

	"ipam-next/models"
)

// Role constants
const (
	RoleAdmin  = "admin"
	RoleUser   = "user"
	RoleViewer = "viewer"
)

// ValidRoles lists all valid role values
var ValidRoles = []string{RoleAdmin, RoleUser, RoleViewer}

// Permission constants
const (
	// Section permissions
	PermSectionCreate = "section:create"
	PermSectionRead   = "section:read"
	PermSectionUpdate = "section:update"
	PermSectionDelete = "section:delete"

	// Subnet permissions
	PermSubnetCreate = "subnet:create"
	PermSubnetRead   = "subnet:read"
	PermSubnetUpdate = "subnet:update"
	PermSubnetDelete = "subnet:delete"

	// IP address permissions
	PermIPCreate   = "ip:create"
	PermIPRead     = "ip:read"
	PermIPUpdate   = "ip:update"
	PermIPDelete   = "ip:delete"
	PermIPAssign   = "ip:assign"
	PermIPRelease  = "ip:release"

	// Auth permissions
	PermTokenCreate = "token:create"
	PermTokenRead   = "token:read"
	PermTokenDelete = "token:delete"
)

// HasPermission checks if a user has a specific permission
func (s *Service) HasPermission(user *models.User, permission string) bool {
	if user == nil {
		return false
	}

	// Admin can do everything
	if user.Role == RoleAdmin {
		return true
	}

	// Map roles to their permissions
	permissions := getRolePermissions(user.Role)

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// CanAccessResource checks if a user can access a specific resource
func (s *Service) CanAccessResource(user *models.User, action string) bool {
	return s.HasPermission(user, action)
}

// RequirePermission returns an error if the user lacks the permission
func (s *Service) RequirePermission(user *models.User, permission string) error {
	if !s.HasPermission(user, permission) {
		return fmt.Errorf("user does not have permission: %s", permission)
	}
	return nil
}

// getRolePermissions returns the permissions for a given role
func getRolePermissions(role string) []string {
	switch role {
	case RoleAdmin:
		return []string{
			// Admin has all permissions
			PermSectionCreate, PermSectionRead, PermSectionUpdate, PermSectionDelete,
			PermSubnetCreate, PermSubnetRead, PermSubnetUpdate, PermSubnetDelete,
			PermIPCreate, PermIPRead, PermIPUpdate, PermIPDelete, PermIPAssign, PermIPRelease,
			PermTokenCreate, PermTokenRead, PermTokenDelete,
		}

	case RoleUser:
		return []string{
			// User can read and write to most resources
			PermSectionRead, PermSectionCreate, PermSectionUpdate, PermSectionDelete,
			PermSubnetRead, PermSubnetCreate, PermSubnetUpdate, PermSubnetDelete,
			PermIPRead, PermIPCreate, PermIPUpdate, PermIPDelete, PermIPAssign, PermIPRelease,
			PermTokenCreate, PermTokenRead, PermTokenDelete,
		}

	case RoleViewer:
		return []string{
			// Viewer can only read
			PermSectionRead,
			PermSubnetRead,
			PermIPRead,
			PermTokenRead,
		}

	default:
		return []string{}
	}
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	for _, valid := range ValidRoles {
		if valid == role {
			return true
		}
	}
	return false
}
