package services

import (
	"context"
	"fmt"
	"net"
	"sort"

	"ipam-next/models"
)

// GetDashboardSummary returns aggregate IPAM statistics.
func (s *Service) GetDashboardSummary(ctx context.Context) (*models.DashboardSummary, error) {
	return s.repository.GetDashboardSummary(ctx)
}

// GetDashboardRecentActivity returns the most recent IPAM-relevant audit log entries.
func (s *Service) GetDashboardRecentActivity(ctx context.Context) ([]*models.DashboardActivity, error) {
	return s.repository.GetDashboardRecentActivity(ctx)
}

// GetSubnetTree returns subnets for a section as a nested tree, ordered by network address.
// Since the DB schema has no parent_subnet_id, we build the hierarchy by containment:
// subnet A is parent of subnet B if B's network is within A's network and there is no
// smaller subnet C also containing B.
func (s *Service) GetSubnetTree(ctx context.Context, sectionID int64) ([]models.SubnetTreeNode, error) {
	if sectionID <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}

	nodes, err := s.repository.GetSubnetTreeBySection(ctx, sectionID)
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
		bestParent := -1
		bestPfx := -1
		for j := 0; j < i; j++ {
			pj, _ := items[j].net.Mask.Size()
			pi, _ := items[i].net.Mask.Size()
			if pj < pi && items[j].net.Contains(items[i].net.IP) {
				if pj > bestPfx {
					bestPfx = pj
					bestParent = j
				}
			}
		}
		parentIdx[i] = bestParent
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

// ListSectionsPaginated returns a paginated list of sections.
func (s *Service) ListSectionsPaginated(ctx context.Context, page, limit int) ([]*models.Section, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repository.ListSectionsPaginated(ctx, limit, offset)
}

// ListSubnetsPaginated returns a paginated list of subnets for a section.
func (s *Service) ListSubnetsPaginated(ctx context.Context, sectionID int64, page, limit int) ([]*models.Subnet, int64, error) {
	if sectionID <= 0 {
		return nil, 0, fmt.Errorf("invalid section ID")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repository.ListSubnetsBySectionPaginated(ctx, sectionID, limit, offset)
}

// ListIPAddressesPaginated returns a paginated list of IP addresses for a subnet.
func (s *Service) ListIPAddressesPaginated(ctx context.Context, subnetID int64, page, limit int) ([]*models.IPAddress, int64, error) {
	if subnetID <= 0 {
		return nil, 0, fmt.Errorf("invalid subnet ID")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repository.ListIPAddressesBySubnetPaginated(ctx, subnetID, limit, offset)
}
