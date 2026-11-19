package services

import (
	"context"
	"errors"
	"fmt"

	"padduck/models"
)

// ErrGrantDenied is returned when a caller may not create, revoke or list a
// grant. Handlers map it to 403.
var ErrGrantDenied = errors.New("not allowed to manage grants for this scope")

// Scope types a grant can be limited to. "section" is the legacy name the
// permission checks use for a network and is treated as "network".
const (
	ScopeNetwork          = "network"
	ScopeSubnet           = "subnet"
	ScopeVLAN             = "vlan"
	ScopeVRF              = "vrf"
	ScopeLocation         = "location"
	ScopeCustomer         = "customer"
	ScopeAutonomousSystem = "autonomous_system"
)

var grantScopeTypes = map[string]bool{
	ScopeNetwork: true, ScopeSubnet: true, ScopeVLAN: true, ScopeVRF: true,
	ScopeLocation: true, ScopeCustomer: true, ScopeAutonomousSystem: true,
}

// GrantScopeTypes lists the scope types accepted when creating a grant.
func GrantScopeTypes() []string {
	return []string{ScopeNetwork, ScopeSubnet, ScopeVLAN, ScopeVRF, ScopeLocation, ScopeCustomer, ScopeAutonomousSystem}
}

func normalizeScopeType(t string) string {
	if t == "section" {
		return ScopeNetwork
	}
	return t
}

// expandScopes returns the given scopes plus every resource that contains them,
// so a grant on a container covers what is inside it:
//
//	subnet   -> its parent subnets, its network, its VLAN, its location and the location's ancestors
//	location -> its ancestors
//
// Scope types are normalized. A lookup failure leaves the scope unexpanded,
// which can only narrow access.
func (s *IdentityService) expandScopes(ctx context.Context, scopes []ResourceScope) []ResourceScope {
	out := make([]ResourceScope, 0, len(scopes)*3)
	seen := map[ResourceScope]bool{}
	add := func(t string, id int64) {
		k := ResourceScope{Type: t, ID: id}
		if !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	addLocation := func(id int64) {
		ids, err := s.repo.GetLocationAncestors(ctx, id)
		if err != nil || len(ids) == 0 {
			add(ScopeLocation, id)
			return
		}
		for _, a := range ids {
			add(ScopeLocation, a)
		}
	}
	for _, sc := range scopes {
		t := normalizeScopeType(sc.Type)
		add(t, sc.ID)
		switch t {
		case ScopeSubnet:
			id := sc.ID
			for depth := 0; depth < 16; depth++ { // bounded walk up the parent chain
				sub, err := s.repo.GetSubnetByID(ctx, id)
				if err != nil {
					break
				}
				add(ScopeNetwork, sub.NetworkID)
				if sub.VLANID != nil {
					add(ScopeVLAN, *sub.VLANID)
				}
				if sub.LocationID != nil {
					addLocation(*sub.LocationID)
				}
				if sub.ParentSubnetID == nil {
					break
				}
				id = *sub.ParentSubnetID
				add(ScopeSubnet, id)
			}
		case ScopeLocation:
			addLocation(sc.ID)
		}
		if sc.LocationScope != nil {
			addLocation(*sc.LocationScope)
		}
	}
	return out
}

func grantCovers(g *models.RoleGrant, expanded []ResourceScope) bool {
	if g.ScopeType == nil {
		return true
	}
	if g.ScopeID == nil {
		return false
	}
	t := normalizeScopeType(*g.ScopeType)
	for _, sc := range expanded {
		if sc.Type == t && sc.ID == *g.ScopeID {
			return true
		}
	}
	return false
}

// hasGrant reports whether a direct grant gives the user permission over the
// scopes: any global grant, or a scoped grant on one of the scopes or a
// resource containing them. With no scopes only global grants count.
func (s *IdentityService) hasGrant(ctx context.Context, userID int64, permission string, scopes []ResourceScope) bool {
	grants, err := s.repo.ListUserPermissionGrants(ctx, userID, permission)
	if err != nil || len(grants) == 0 {
		return false
	}
	var expanded []ResourceScope
	for _, g := range grants {
		if g.ScopeType == nil {
			return true
		}
		if expanded == nil {
			expanded = s.expandScopes(ctx, scopes)
		}
		if grantCovers(g, expanded) {
			return true
		}
	}
	return false
}

// canManageScope reports whether the user may administer grants over a scope:
// they need PermV2OrgWrite globally (nil scope) or over that scope.
func (s *IdentityService) canManageScope(ctx context.Context, userID int64, scopeType *string, scopeID *int64) bool {
	if scopeType == nil || scopeID == nil {
		return s.CheckPermission(ctx, userID, PermV2OrgWrite) == nil
	}
	return s.CheckPermission(ctx, userID, PermV2OrgWrite, ResourceScope{Type: normalizeScopeType(*scopeType), ID: *scopeID}) == nil
}

// CreateGrantRequest carries the parameters for granting a direct permission.
type CreateGrantRequest struct {
	UserID     int64   `json:"user_id"`
	Permission string  `json:"permission"`
	ScopeType  *string `json:"scope_type"`
	ScopeID    *int64  `json:"scope_id"`
}

// CreateGrant grants a permission to a user, optionally limited to one
// resource. The grantor must be allowed to administer that scope (org-write
// over it) and must hold the permission over it themselves, so a delegated
// admin cannot hand out more than they have or reach outside their scope.
func (s *IdentityService) CreateGrant(ctx context.Context, grantorID int64, req CreateGrantRequest) (*models.RoleGrant, error) {
	if !IsValidPermission(req.Permission) {
		return nil, fmt.Errorf("unknown permission: %s", req.Permission)
	}
	if req.Permission == PermV2PlatformAdmin {
		return nil, fmt.Errorf("the platform admin permission cannot be granted this way")
	}
	if (req.ScopeType == nil) != (req.ScopeID == nil) {
		return nil, fmt.Errorf("scope_type and scope_id must be given together")
	}
	var scopeType *string
	if req.ScopeType != nil {
		t := normalizeScopeType(*req.ScopeType)
		if !grantScopeTypes[t] {
			return nil, fmt.Errorf("unknown scope type: %s", *req.ScopeType)
		}
		if err := s.scopeResourceExists(ctx, t, *req.ScopeID); err != nil {
			return nil, err
		}
		scopeType = &t
	}
	if !s.canManageScope(ctx, grantorID, scopeType, req.ScopeID) {
		return nil, ErrGrantDenied
	}
	var scopes []ResourceScope
	if scopeType != nil {
		scopes = []ResourceScope{{Type: *scopeType, ID: *req.ScopeID}}
	}
	if err := s.CheckPermission(ctx, grantorID, req.Permission, scopes...); err != nil {
		return nil, fmt.Errorf("%w: you do not hold %s over this scope", ErrGrantDenied, req.Permission)
	}
	target, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if target.OrganizationID == nil {
		return nil, fmt.Errorf("target user has no organization")
	}
	return s.repo.CreateRoleGrant(ctx, *target.OrganizationID, req.UserID, req.Permission, scopeType, req.ScopeID, &grantorID)
}

// scopeResourceExists checks the scope target exists in the caller's organization.
func (s *IdentityService) scopeResourceExists(ctx context.Context, scopeType string, id int64) error {
	var err error
	switch scopeType {
	case ScopeNetwork:
		_, err = s.repo.GetNetworkByID(ctx, id)
	case ScopeSubnet:
		_, err = s.repo.GetSubnetByID(ctx, id)
	case ScopeVLAN:
		_, err = s.repo.GetVLANByID(ctx, id)
	case ScopeVRF:
		_, err = s.repo.GetVRFByID(ctx, id)
	case ScopeLocation:
		_, err = s.repo.GetLocationByID(ctx, id)
	case ScopeCustomer:
		_, err = s.repo.GetCustomerByID(ctx, id)
	case ScopeAutonomousSystem:
		_, err = s.repo.GetAutonomousSystemByID(ctx, id)
	}
	if err != nil {
		return fmt.Errorf("%s %d not found", scopeType, id)
	}
	return nil
}

// RevokeGrant removes a grant. The caller must administer the grant's scope.
func (s *IdentityService) RevokeGrant(ctx context.Context, callerID, grantID int64) (*models.RoleGrant, error) {
	g, err := s.repo.GetRoleGrant(ctx, grantID)
	if err != nil {
		return nil, fmt.Errorf("grant not found")
	}
	if !s.canManageScope(ctx, callerID, g.ScopeType, g.ScopeID) {
		return nil, ErrGrantDenied
	}
	if err := s.repo.DeleteRoleGrant(ctx, grantID); err != nil {
		return nil, err
	}
	return g, nil
}

// ListUserGrants returns a user's direct grants that the caller may see: all
// of them for the user themselves or for an administrator of the whole org,
// otherwise only those within a scope the caller administers.
func (s *IdentityService) ListUserGrants(ctx context.Context, callerID, userID int64) ([]*models.RoleGrant, error) {
	grants, err := s.repo.ListUserGrants(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]*models.RoleGrant, 0, len(grants))
	for _, g := range grants {
		if callerID == userID || s.canManageScope(ctx, callerID, g.ScopeType, g.ScopeID) {
			out = append(out, g)
		}
	}
	return out, nil
}
