package services

import (
	"context"
	"fmt"
	"math"
	"net"

	"ipam-next/models"
)

// SplitSubnet splits a subnet into 2^(newPrefixLen-currentPrefix) child subnets.
// Validates: newPrefixLen > current prefix, child count <= 256.
// In DB transaction: creates children with parent_subnet_id, moves IPs, marks parent is_container.
// Emits audit log for each child (action "subnet.split").
func (s *Service) SplitSubnet(ctx context.Context, subnetID int64, newPrefixLen int) ([]*models.Subnet, error) {
	parent, err := s.repository.GetSubnetByID(ctx, subnetID)
	if err != nil {
		return nil, fmt.Errorf("subnet not found: %w", err)
	}

	if newPrefixLen <= parent.PrefixLength {
		return nil, fmt.Errorf("new prefix length %d must be greater than current prefix length %d", newPrefixLen, parent.PrefixLength)
	}
	if newPrefixLen > 32 {
		return nil, fmt.Errorf("new prefix length %d exceeds maximum of 32", newPrefixLen)
	}

	diff := newPrefixLen - parent.PrefixLength
	childCount := int(math.Pow(2, float64(diff)))
	if childCount > 256 {
		return nil, fmt.Errorf("split would produce %d subnets, maximum is 256", childCount)
	}

	// Enumerate child CIDRs
	parentCIDR := fmt.Sprintf("%s/%d", parent.NetworkAddress, parent.PrefixLength)
	_, parentNet, err := net.ParseCIDR(parentCIDR)
	if err != nil {
		return nil, fmt.Errorf("invalid parent CIDR: %w", err)
	}

	childSubnets := make([]*models.Subnet, 0, childCount)
	ip := cloneIP(parentNet.IP.To4())
	for i := 0; i < childCount; i++ {
		childNet := &net.IPNet{
			IP:   cloneIP(ip),
			Mask: net.CIDRMask(newPrefixLen, 32),
		}
		child := &models.Subnet{
			SectionID:      parent.SectionID,
			NetworkAddress: childNet.IP.String(),
			PrefixLength:   newPrefixLen,
			Description:    fmt.Sprintf("Split from %s", parentCIDR),
		}
		childSubnets = append(childSubnets, child)
		incrementIP(ip, 32-newPrefixLen)
	}

	if err := s.repository.SplitSubnet(ctx, subnetID, childSubnets); err != nil {
		return nil, fmt.Errorf("split transaction failed: %w", err)
	}

	return childSubnets, nil
}

// cloneIP returns a copy of the given IP.
func cloneIP(ip net.IP) net.IP {
	c := make(net.IP, len(ip))
	copy(c, ip)
	return c
}

// incrementIP adds 2^hostBits to the IP address in place.
func incrementIP(ip net.IP, hostBits int) {
	inc := uint32(1) << uint(hostBits)
	b := ip.To4()
	val := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	val += inc
	b[0] = byte(val >> 24)
	b[1] = byte(val >> 16)
	b[2] = byte(val >> 8)
	b[3] = byte(val)
}

// MergeSubnets merges multiple subnets into a common supernet.
// Validates: all subnets same prefix length, same section, and they form a contiguous block.
// In DB transaction: creates parent subnet, moves all IPs, deletes merged children.
// Emits audit log (action "subnet.merge").
func (s *Service) MergeSubnets(ctx context.Context, subnetIDs []int64) (*models.Subnet, error) {
	if len(subnetIDs) < 2 {
		return nil, fmt.Errorf("at least 2 subnets required for merge")
	}

	// Load all subnets
	subnets := make([]*models.Subnet, 0, len(subnetIDs))
	for _, id := range subnetIDs {
		sub, err := s.repository.GetSubnetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("subnet %d not found: %w", id, err)
		}
		subnets = append(subnets, sub)
	}

	// Validate: same section
	sectionID := subnets[0].SectionID
	prefixLen := subnets[0].PrefixLength
	for _, sub := range subnets[1:] {
		if sub.SectionID != sectionID {
			return nil, fmt.Errorf("all subnets must be in the same section")
		}
		if sub.PrefixLength != prefixLen {
			return nil, fmt.Errorf("all subnets must have the same prefix length")
		}
	}

	// Validate: count must be a power of 2
	n := len(subnetIDs)
	if n == 0 || (n&(n-1)) != 0 {
		return nil, fmt.Errorf("number of subnets to merge must be a power of 2 (got %d)", n)
	}

	// Compute new prefix length
	newPrefixLen := prefixLen - int(math.Log2(float64(n)))
	if newPrefixLen < 0 {
		return nil, fmt.Errorf("resulting prefix length would be negative")
	}

	// Find the supernet: use the smallest network address
	var minIP net.IP
	for _, sub := range subnets {
		ip := net.ParseIP(sub.NetworkAddress).To4()
		if minIP == nil || ipLess(ip, minIP) {
			minIP = ip
		}
	}

	// Compute the supernet
	supernet := &net.IPNet{
		IP:   minIP.Mask(net.CIDRMask(newPrefixLen, 32)),
		Mask: net.CIDRMask(newPrefixLen, 32),
	}

	// Validate: all subnets fall within the supernet
	for _, sub := range subnets {
		subIP := net.ParseIP(sub.NetworkAddress).To4()
		if !supernet.Contains(subIP) {
			return nil, fmt.Errorf("subnet %s/%d is not contiguous with the others for merging", sub.NetworkAddress, sub.PrefixLength)
		}
	}

	parent := &models.Subnet{
		SectionID:      sectionID,
		NetworkAddress: supernet.IP.String(),
		PrefixLength:   newPrefixLen,
		Description:    fmt.Sprintf("Merged from %d subnets", len(subnetIDs)),
	}

	result, err := s.repository.MergeSubnets(ctx, subnetIDs, parent)
	if err != nil {
		return nil, fmt.Errorf("merge transaction failed: %w", err)
	}

	return result, nil
}

// ipLess returns true if a < b (byte-wise comparison).
func ipLess(a, b net.IP) bool {
	for i := range a {
		if a[i] < b[i] {
			return true
		}
		if a[i] > b[i] {
			return false
		}
	}
	return false
}

// ResizeSubnet changes a subnet's CIDR to newPrefix.
// Expand: checks for overlap with sibling subnets.
// Shrink: checks if any IPs are outside the new range (returns error listing them).
// Also checks new CIDR fits within parent subnet if parent exists.
// Emits audit log (action "subnet.resize").
func (s *Service) ResizeSubnet(ctx context.Context, subnetID int64, newPrefix string) (*models.Subnet, error) {
	existing, err := s.repository.GetSubnetByID(ctx, subnetID)
	if err != nil {
		return nil, fmt.Errorf("subnet not found: %w", err)
	}

	newIP, newNet, err := net.ParseCIDR(newPrefix)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR %q: %w", newPrefix, err)
	}
	_ = newIP

	newNetworkAddr := newNet.IP.String()
	ones, _ := newNet.Mask.Size()
	newPrefixLen := ones

	// Check new CIDR fits within parent subnet if parent exists
	if existing.ParentSubnetID != nil {
		parent, err := s.repository.GetSubnetByID(ctx, *existing.ParentSubnetID)
		if err == nil {
			parentCIDR := fmt.Sprintf("%s/%d", parent.NetworkAddress, parent.PrefixLength)
			_, parentNet, perr := net.ParseCIDR(parentCIDR)
			if perr == nil {
				if !parentNet.Contains(newNet.IP) {
					return nil, fmt.Errorf("new CIDR %s does not fit within parent subnet %s", newPrefix, parentCIDR)
				}
			}
		}
	}

	isShrink := newPrefixLen > existing.PrefixLength

	if isShrink {
		// Check if any IPs are outside the new range
		outsideIPs, err := s.repository.ListIPsOutsideCIDR(ctx, subnetID, newNetworkAddr, newPrefixLen)
		if err != nil {
			return nil, fmt.Errorf("checking IPs outside new range: %w", err)
		}
		if len(outsideIPs) > 0 {
			return nil, &SubnetResizeConflictError{
				ConflictingIPs: outsideIPs,
			}
		}
	} else {
		// Expand: check for overlap with sibling subnets
		siblings, err := s.repository.ListSiblingSubnets(ctx, subnetID)
		if err != nil {
			return nil, fmt.Errorf("checking sibling subnets: %w", err)
		}
		for _, sib := range siblings {
			sibCIDR := fmt.Sprintf("%s/%d", sib.NetworkAddress, sib.PrefixLength)
			_, sibNet, err := net.ParseCIDR(sibCIDR)
			if err != nil {
				continue
			}
			if newNet.Contains(sibNet.IP) || sibNet.Contains(newNet.IP) {
				return nil, &SubnetResizeConflictError{
					ConflictingSubnets: siblings,
				}
			}
		}
	}

	result, err := s.repository.ResizeSubnet(ctx, subnetID, newNetworkAddr, newPrefixLen)
	if err != nil {
		return nil, fmt.Errorf("resize failed: %w", err)
	}

	return result, nil
}

// SubnetResizeConflictError is returned when resize cannot proceed due to conflicts.
type SubnetResizeConflictError struct {
	ConflictingIPs     []*models.IPAddress
	ConflictingSubnets []*models.Subnet
}

func (e *SubnetResizeConflictError) Error() string {
	if len(e.ConflictingIPs) > 0 {
		return fmt.Sprintf("resize would leave %d IP address(es) outside the new subnet range", len(e.ConflictingIPs))
	}
	if len(e.ConflictingSubnets) > 0 {
		return fmt.Sprintf("resize would overlap with %d existing subnet(s)", len(e.ConflictingSubnets))
	}
	return "resize conflict"
}
