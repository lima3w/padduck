package services

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"ipam-next/models"
)

// SubnetOverlapError is returned when a new subnet overlaps an existing one
type SubnetOverlapError struct {
	ConflictingCIDR string
}

func (e *SubnetOverlapError) Error() string {
	return fmt.Sprintf("subnet overlaps with existing subnet %s", e.ConflictingCIDR)
}

// ValidateCIDR validates a CIDR notation
func ValidateCIDR(address string, prefixLength int) error {
	if prefixLength < 0 || prefixLength > 32 {
		return fmt.Errorf("invalid prefix length: %d", prefixLength)
	}

	if net.ParseIP(address) == nil {
		return fmt.Errorf("invalid network address: %s", address)
	}

	return nil
}

// checkOverlap checks whether the given CIDR overlaps any existing subnet in the section (excluding excludeID)
func (s *Service) checkOverlap(ctx context.Context, sectionID int64, networkAddress string, prefixLength int, excludeID int64) error {
	allowed, _ := s.Config.GetCtx(ctx, "allow_subnet_overlaps")
	if allowed == "true" {
		return nil
	}

	newCIDR := fmt.Sprintf("%s/%d", networkAddress, prefixLength)
	_, newNet, err := net.ParseCIDR(newCIDR)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %s", newCIDR)
	}

	existing, err := s.repository.ListSubnetsBySection(ctx, sectionID)
	if err != nil {
		return err
	}

	for _, sub := range existing {
		if sub.ID == excludeID {
			continue
		}
		existingCIDR := fmt.Sprintf("%s/%d", sub.NetworkAddress, sub.PrefixLength)
		_, existingNet, err := net.ParseCIDR(existingCIDR)
		if err != nil {
			continue
		}
		if newNet.Contains(existingNet.IP) || existingNet.Contains(newNet.IP) {
			return &SubnetOverlapError{ConflictingCIDR: existingCIDR}
		}
	}
	return nil
}

// OverlapPair represents two overlapping subnets
type OverlapPair struct {
	SubnetA *models.Subnet `json:"subnet_a"`
	SubnetB *models.Subnet `json:"subnet_b"`
}

// OverlapReport returns all overlapping subnet pairs across all sections
func (s *Service) OverlapReport(ctx context.Context) ([]*OverlapPair, error) {
	all, err := s.repository.ListAllSubnets(ctx)
	if err != nil {
		return nil, err
	}

	// Group by section_id in Go to avoid one DB query per section
	bySec := make(map[int64][]*models.Subnet)
	for _, sub := range all {
		bySec[sub.SectionID] = append(bySec[sub.SectionID], sub)
	}

	var pairs []*OverlapPair
	for _, subnets := range bySec {
		for i, a := range subnets {
			_, netA, err := net.ParseCIDR(fmt.Sprintf("%s/%d", a.NetworkAddress, a.PrefixLength))
			if err != nil {
				continue
			}
			for j := i + 1; j < len(subnets); j++ {
				b := subnets[j]
				_, netB, err := net.ParseCIDR(fmt.Sprintf("%s/%d", b.NetworkAddress, b.PrefixLength))
				if err != nil {
					continue
				}
				if netA.Contains(netB.IP) || netB.Contains(netA.IP) {
					pairs = append(pairs, &OverlapPair{SubnetA: a, SubnetB: b})
				}
			}
		}
	}
	return pairs, nil
}

// broadcastAddr computes the broadcast address of a network
func broadcastAddr(n *net.IPNet) string {
	ip := n.IP.To4()
	mask := n.Mask
	broadcast := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		broadcast[i] = ip[i] | ^mask[i]
	}
	return broadcast.String()
}

// validateGatewayInCIDR checks that the gateway IP is within the subnet
func validateGatewayInCIDR(gateway, networkAddress string, prefixLength int) error {
	gwIP := net.ParseIP(gateway)
	if gwIP == nil {
		return fmt.Errorf("gateway is not a valid IP address")
	}
	cidr := fmt.Sprintf("%s/%d", networkAddress, prefixLength)
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	if !ipNet.Contains(gwIP) {
		return fmt.Errorf("gateway %s is not within subnet %s", gateway, cidr)
	}
	return nil
}

// validateVLANVRFConsistency checks that if a subnet's VLAN belongs to a domain,
// all other subnets in that domain use the same VRF (or no VRF). Returns an error if mismatch.
func (s *Service) validateVLANVRFConsistency(ctx context.Context, vlanID int64, subnetVRFID *int64) error {
	vlan, err := s.repository.GetVLANByID(ctx, vlanID)
	if err != nil {
		return fmt.Errorf("VLAN not found")
	}
	// Only enforce if the VLAN belongs to a domain
	if vlan.DomainID == nil {
		return nil
	}
	// If the VLAN requires a specific VRF, the subnet must specify that same VRF.
	if vlan.VRFID != nil {
		if subnetVRFID == nil || *vlan.VRFID != *subnetVRFID {
			return fmt.Errorf("subnet VRF does not match VLAN VRF (domain %d requires VRF %d)", *vlan.DomainID, *vlan.VRFID)
		}
	}
	return nil
}

// GetVLANSubnets returns all subnets assigned to a VLAN.
func (s *Service) GetVLANSubnets(ctx context.Context, vlanID int64) ([]*models.Subnet, error) {
	if vlanID <= 0 {
		return nil, fmt.Errorf("invalid VLAN ID")
	}
	return s.repository.GetVLANSubnets(ctx, vlanID)
}

func (s *Service) AssignSubnetToVLAN(ctx context.Context, vlanID, subnetID int64) (*models.Subnet, error) {
	if vlanID <= 0 {
		return nil, fmt.Errorf("invalid VLAN ID")
	}
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}
	if _, err := s.repository.GetVLANByID(ctx, vlanID); err != nil {
		return nil, fmt.Errorf("VLAN not found")
	}
	if err := s.validateVLANVRFConsistency(ctx, vlanID, nil); err != nil {
		return nil, err
	}
	return s.repository.AssignSubnetToVLAN(ctx, subnetID, &vlanID)
}

func (s *Service) RemoveSubnetFromVLAN(ctx context.Context, vlanID, subnetID int64) (*models.Subnet, error) {
	if vlanID <= 0 {
		return nil, fmt.Errorf("invalid VLAN ID")
	}
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}
	subnet, err := s.repository.GetSubnetByID(ctx, subnetID)
	if err != nil {
		return nil, err
	}
	if subnet.VLANID == nil || *subnet.VLANID != vlanID {
		return nil, fmt.Errorf("subnet is not assigned to this VLAN")
	}
	return s.repository.AssignSubnetToVLAN(ctx, subnetID, nil)
}

// CreateSubnet creates a new subnet with CIDR validation and optional gateway/auto-reserve settings
func (s *Service) CreateSubnet(ctx context.Context, sectionID int64, networkAddress string, prefixLength int, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64, customFields ...map[string]*string) (*models.Subnet, error) {
	if sectionID <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}

	if err := ValidateCIDR(networkAddress, prefixLength); err != nil {
		return nil, err
	}

	// Validate gateway if provided
	if gateway != nil && *gateway != "" {
		if err := validateGatewayInCIDR(*gateway, networkAddress, prefixLength); err != nil {
			return nil, err
		}
	} else {
		gateway = nil
	}

	if err := s.checkOverlap(ctx, sectionID, networkAddress, prefixLength, 0); err != nil {
		return nil, err
	}

	// Apply global defaults
	if !autoFirst {
		if v, _ := s.Config.GetCtx(ctx, "default_auto_reserve_first"); v == "true" {
			autoFirst = true
		}
	}
	if !autoLast {
		if v, _ := s.Config.GetCtx(ctx, "default_auto_reserve_last"); v == "true" {
			autoLast = true
		}
	}

	// VRF consistency check for VLAN-domain subnets
	if vlanID != nil {
		if err := s.validateVLANVRFConsistency(ctx, *vlanID, nil); err != nil {
			return nil, err
		}
	}

	subnet, err := s.repository.CreateSubnetWithVLAN(ctx, sectionID, networkAddress, prefixLength, description, gateway, autoFirst, autoLast, locationID, nameserverID, vlanID)
	if err != nil {
		return nil, err
	}

	// Auto-reserve first/last IP
	cidr := fmt.Sprintf("%s/%d", networkAddress, prefixLength)
	_, ipNet, _ := net.ParseCIDR(cidr)
	if autoFirst && ipNet != nil {
		networkIP := ipNet.IP.String()
		if _, err := s.repository.CreateIPAddress(ctx, subnet.ID, networkIP, "", "reserved", nil, nil, nil, nil); err != nil {
			slog.Warn("auto-reserve first IP failed", "subnet_id", subnet.ID, "ip", networkIP, "error", err)
		}
	}
	if autoLast && ipNet != nil {
		bcastIP := broadcastAddr(ipNet)
		if _, err := s.repository.CreateIPAddress(ctx, subnet.ID, bcastIP, "", "reserved", nil, nil, nil, nil); err != nil {
			slog.Warn("auto-reserve last IP failed", "subnet_id", subnet.ID, "ip", bcastIP, "error", err)
		}
	}

	if len(customFields) > 0 && customFields[0] != nil {
		_ = s.SetCustomFieldValues(ctx, "subnet", subnet.ID, customFields[0])
		subnet.CustomFields, _ = s.repository.GetCustomFieldValues(ctx, "subnet", subnet.ID)
	}

	return subnet, nil
}

// GetSubnet retrieves a subnet by ID
func (s *Service) GetSubnet(ctx context.Context, id int64) (*models.Subnet, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	subnet, err := s.repository.GetSubnetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	subnet.CustomFields, _ = s.repository.GetCustomFieldValues(ctx, "subnet", id)
	return subnet, nil
}

// ListSubnets returns all subnets in a section
func (s *Service) ListSubnets(ctx context.Context, sectionID int64) ([]*models.Subnet, error) {
	if sectionID <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}

	return s.repository.ListSubnetsBySection(ctx, sectionID)
}

// UpdateSubnet updates a subnet's description, gateway, auto-reserve settings, location, nameserver, and VLAN.
func (s *Service) UpdateSubnet(ctx context.Context, id int64, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64, customFields ...map[string]*string) (*models.Subnet, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	if gateway != nil && *gateway != "" {
		existing, err := s.repository.GetSubnetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if err := validateGatewayInCIDR(*gateway, existing.NetworkAddress, existing.PrefixLength); err != nil {
			return nil, err
		}
	} else {
		gateway = nil
	}

	// VRF consistency check for VLAN-domain subnets
	if vlanID != nil {
		if err := s.validateVLANVRFConsistency(ctx, *vlanID, nil); err != nil {
			return nil, err
		}
	}

	subnet, err := s.repository.UpdateSubnetWithVLAN(ctx, id, description, gateway, autoFirst, autoLast, locationID, nameserverID, vlanID)
	if err != nil {
		return nil, err
	}

	if len(customFields) > 0 && customFields[0] != nil {
		_ = s.SetCustomFieldValues(ctx, "subnet", subnet.ID, customFields[0])
	}
	subnet.CustomFields, _ = s.repository.GetCustomFieldValues(ctx, "subnet", subnet.ID)
	return subnet, nil
}

// ListSubnetsByLocation returns all subnets assigned to the given location.
func (s *Service) ListSubnetsByLocation(ctx context.Context, locationID int64) ([]*models.Subnet, error) {
	if locationID <= 0 {
		return nil, fmt.Errorf("invalid location ID")
	}
	return s.repository.ListSubnetsByLocation(ctx, locationID)
}

// DeleteSubnet deletes a subnet and its IP addresses (cascade)
func (s *Service) DeleteSubnet(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid subnet ID")
	}

	return s.repository.DeleteSubnet(ctx, id)
}
