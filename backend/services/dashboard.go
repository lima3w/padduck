package services

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"padduck/models"
	"padduck/repository"
)

// GetDashboardSummary returns aggregate IPAM statistics.
func (s *Service) GetDashboardSummary(ctx context.Context) (*models.DashboardSummary, error) {
	if s.dashboardSummaryCache == nil {
		s.dashboardSummaryCache = newTTLCache[*models.DashboardSummary](30 * time.Second)
	}
	if summary, ok := s.dashboardSummaryCache.get("summary"); ok {
		return cloneDashboardSummary(summary), nil
	}

	summary, err := s.repository.GetDashboardSummary(ctx)
	if err != nil {
		return nil, err
	}
	s.dashboardSummaryCache.set("summary", cloneDashboardSummary(summary))
	return cloneDashboardSummary(summary), nil
}

// GetDashboardRecentActivity returns the most recent IPAM-relevant audit log entries.
func (s *Service) GetDashboardRecentActivity(ctx context.Context) ([]*models.DashboardActivity, error) {
	if s.dashboardActivityCache == nil {
		s.dashboardActivityCache = newTTLCache[[]*models.DashboardActivity](15 * time.Second)
	}
	if activities, ok := s.dashboardActivityCache.get("recent"); ok {
		return cloneDashboardActivities(activities), nil
	}

	activities, err := s.repository.GetDashboardRecentActivity(ctx)
	if err != nil {
		return nil, err
	}
	s.dashboardActivityCache.set("recent", cloneDashboardActivities(activities))
	return cloneDashboardActivities(activities), nil
}

func cloneDashboardSummary(summary *models.DashboardSummary) *models.DashboardSummary {
	if summary == nil {
		return nil
	}
	out := *summary
	out.TopSubnets = append([]models.SubnetUtilisation(nil), summary.TopSubnets...)
	return &out
}

func cloneDashboardActivities(activities []*models.DashboardActivity) []*models.DashboardActivity {
	out := make([]*models.DashboardActivity, 0, len(activities))
	for _, activity := range activities {
		if activity == nil {
			out = append(out, nil)
			continue
		}
		clone := *activity
		out = append(out, &clone)
	}
	return out
}

// GetSubnetTree returns subnets for a section as a nested tree, ordered by network address.
// Since the DB schema has no parent_subnet_id, we build the hierarchy by containment:
// subnet A is parent of subnet B if B's network is within A's network and there is no
// smaller subnet C also containing B.
func (s *Service) GetSubnetTree(ctx context.Context, networkID int64) ([]models.SubnetTreeNode, error) {
	if networkID <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}

	nodes, err := s.repository.GetSubnetTreeBySection(ctx, networkID)
	if err != nil {
		return nil, err
	}

	return buildTree(nodes), nil
}

// buildTree arranges flat nodes into a parent-child tree based on CIDR containment.
func buildTree(flat []models.SubnetTreeNode) []models.SubnetTreeNode {
	// Parse each node's network
	type parsed struct {
		node models.SubnetTreeNode
		net  *net.IPNet
	}

	items := make([]parsed, 0, len(flat))
	for _, n := range flat {
		_, ipNet, err := net.ParseCIDR(n.CIDR)
		if err == nil {
			items = append(items, parsed{node: n, net: ipNet})
		}
	}

	// Sort by prefix length ascending (largest subnets first), then by network address
	sort.Slice(items, func(i, j int) bool {
		pi, _ := items[i].net.Mask.Size()
		pj, _ := items[j].net.Mask.Size()
		if pi != pj {
			return pi < pj
		}
		return items[i].net.IP.String() < items[j].net.IP.String()
	})

	// For each node, find its most-specific parent
	parentIdx := make([]int, len(items))
	for i := range parentIdx {
		parentIdx[i] = -1
	}

	for i := 1; i < len(items); i++ {
		// Iterate backwards: since items are sorted by ascending prefix length,
		// the first container we find going backwards is the most specific parent.
		for j := i - 1; j >= 0; j-- {
			pj, _ := items[j].net.Mask.Size()
			pi, _ := items[i].net.Mask.Size()
			if pj < pi && items[j].net.Contains(items[i].net.IP) {
				parentIdx[i] = j
				break // found most-specific parent, stop early
			}
		}
	}

	// Build tree using indices
	nodeChildren := make([][]int, len(items))
	for i := range nodeChildren {
		nodeChildren[i] = make([]int, 0)
	}
	roots := make([]int, 0)
	for i, p := range parentIdx {
		if p == -1 {
			roots = append(roots, i)
		} else {
			nodeChildren[p] = append(nodeChildren[p], i)
		}
	}

	var build func(idx int) models.SubnetTreeNode
	build = func(idx int) models.SubnetTreeNode {
		n := items[idx].node
		n.Children = make([]models.SubnetTreeNode, 0, len(nodeChildren[idx]))
		for _, childIdx := range nodeChildren[idx] {
			n.Children = append(n.Children, build(childIdx))
		}
		return n
	}

	result := make([]models.SubnetTreeNode, 0, len(roots))
	for _, r := range roots {
		result = append(result, build(r))
	}
	return result
}

// ListNetworksPaginated returns a paginated list of sections.
func (s *Service) ListNetworksPaginated(ctx context.Context, page, limit int) ([]*models.Network, int64, error) {
	return s.ListNetworksPaginatedWithOptions(ctx, page, limit, repository.ListOptions{})
}

func (s *Service) ListNetworksPaginatedWithOptions(ctx context.Context, page, limit int, opts repository.ListOptions) ([]*models.Network, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	opts.Limit = limit
	opts.Offset = (page - 1) * limit
	return s.repository.ListNetworksPaginatedWithOptions(ctx, opts)
}

// ListSubnetsPaginated returns a paginated list of subnets for a section.
func (s *Service) ListSubnetsPaginated(ctx context.Context, networkID int64, page, limit int) ([]*models.Subnet, int64, error) {
	return s.ListSubnetsPaginatedWithOptions(ctx, networkID, page, limit, repository.ListOptions{})
}

func (s *Service) ListSubnetsPaginatedWithOptions(ctx context.Context, networkID int64, page, limit int, opts repository.ListOptions) ([]*models.Subnet, int64, error) {
	if networkID <= 0 {
		return nil, 0, fmt.Errorf("invalid section ID")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	opts.Limit = limit
	opts.Offset = (page - 1) * limit
	return s.repository.ListSubnetsBySectionPaginatedWithOptions(ctx, networkID, opts)
}

// ListIPAddressesPaginated returns a paginated list of IP addresses for a subnet.
func (s *Service) ListIPAddressesPaginated(ctx context.Context, subnetID int64, page, limit int) ([]*models.IPAddress, int64, error) {
	return s.ListIPAddressesPaginatedWithOptions(ctx, subnetID, page, limit, repository.ListOptions{})
}

// ListIPAddressesFullRange returns every address in the subnet's IPv4 CIDR range merged with
// recorded IP rows. Virtual entries (status "available", Virtual=true) are synthesised for
// addresses that have no database row.
func (s *Service) ListIPAddressesFullRange(ctx context.Context, subnetID int64, page, limit int) ([]*models.IPAddress, int64, error) {
	if subnetID <= 0 {
		return nil, 0, fmt.Errorf("invalid subnet ID")
	}
	subnet, err := s.repository.GetSubnetByID(ctx, subnetID)
	if err != nil {
		return nil, 0, err
	}
	if strings.Contains(subnet.NetworkAddress, ":") {
		return nil, 0, fmt.Errorf("full range view is not supported for IPv6 subnets")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repository.ListIPAddressesFullRange(ctx, subnetID, subnet.NetworkAddress, subnet.PrefixLength, offset, limit)
}

func (s *Service) ListIPAddressesPaginatedWithOptions(ctx context.Context, subnetID int64, page, limit int, opts repository.ListOptions) ([]*models.IPAddress, int64, error) {
	if subnetID <= 0 {
		return nil, 0, fmt.Errorf("invalid subnet ID")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	opts.Limit = limit
	opts.Offset = (page - 1) * limit
	return s.repository.ListIPAddressesBySubnetPaginatedWithOptions(ctx, subnetID, opts)
}

// ListUsersPaginated returns a paginated list of users.
func (s *Service) ListUsersPaginated(ctx context.Context, page, limit int) ([]*models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repository.ListUsersPaginated(ctx, limit, offset)
}

// ListVLANsPaginated returns a paginated list of VLANs.
func (s *Service) ListVLANsPaginated(ctx context.Context, page, limit int) ([]*models.VLAN, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repository.ListVLANsPaginated(ctx, limit, offset)
}

// ListVRFsPaginated returns a paginated list of VRFs.
func (s *Service) ListVRFsPaginated(ctx context.Context, page, limit int) ([]*models.VRF, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repository.ListVRFsPaginated(ctx, limit, offset)
}

// ListLocationsPaginated returns a paginated list of locations.
func (s *Service) ListLocationsPaginated(ctx context.Context, page, limit int) ([]*models.Location, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repository.ListLocationsPaginated(ctx, limit, offset)
}

// ListCustomersPaginated returns a paginated list of customers.
func (s *Service) ListCustomersPaginated(ctx context.Context, page, limit int) ([]*models.Customer, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repository.ListCustomersPaginated(ctx, limit, offset)
}

// ListAutonomousSystemsPaginated returns a paginated list of autonomous systems.
func (s *Service) ListAutonomousSystemsPaginated(ctx context.Context, page, limit int) ([]*models.AutonomousSystem, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repository.ListAutonomousSystemsPaginated(ctx, limit, offset)
}
