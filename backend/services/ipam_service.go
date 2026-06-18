package services

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"padduck/models"
	"padduck/repository"
	"padduck/utils"
)

// IPAMService owns all IP address management operations: networks, subnets,
// IPs, VRFs, VLANs, tags, search, dashboard, IPv6 delegations, and subnet tools.
type IPAMService struct {
	repo          *repository.Repository
	config        *ConfigService
	dns           *DNSService
	summaryCache  *ttlCache[*models.DashboardSummary]
	activityCache *ttlCache[[]*models.DashboardActivity]
}

func NewIPAMService(repo *repository.Repository, config *ConfigService, dns *DNSService) *IPAMService {
	return &IPAMService{
		repo:          repo,
		config:        config,
		dns:           dns,
		summaryCache:  newTTLCache[*models.DashboardSummary](30 * time.Second),
		activityCache: newTTLCache[[]*models.DashboardActivity](15 * time.Second),
	}
}

// setCustomFieldValues validates required fields and persists the given values.
func (s *IPAMService) setCustomFieldValues(ctx context.Context, entityType string, entityID int64, values map[string]*string) error {
	defs, err := s.repo.ListCustomFieldDefinitions(ctx, entityType)
	if err != nil {
		return err
	}
	for _, def := range defs {
		if def.IsRequired {
			val, ok := values[def.Name]
			if !ok || val == nil || *val == "" {
				return fmt.Errorf("field %q is required", def.Name)
			}
		}
	}
	return s.repo.SetCustomFieldValues(ctx, entityType, entityID, defs, values)
}

// ---- Networks ---------------------------------------------------------------

func (s *IPAMService) CreateNetwork(ctx context.Context, name, description string, createdBy int64) (*models.Network, error) {
	if name == "" {
		return nil, fmt.Errorf("section name is required")
	}
	return s.repo.CreateNetwork(ctx, name, description, createdBy)
}

func (s *IPAMService) GetNetwork(ctx context.Context, id int64) (*models.Network, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}
	return s.repo.GetNetworkByID(ctx, id)
}

func (s *IPAMService) ListNetworks(ctx context.Context) ([]*models.Network, error) {
	return s.repo.ListAllNetworks(ctx)
}

func (s *IPAMService) UpdateNetwork(ctx context.Context, id int64, name, description string) (*models.Network, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}
	if name == "" {
		return nil, fmt.Errorf("section name is required")
	}
	return s.repo.UpdateNetwork(ctx, id, name, description)
}

func (s *IPAMService) DeleteNetwork(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid section ID")
	}
	return s.repo.DeleteNetwork(ctx, id)
}

// ---- Subnets ----------------------------------------------------------------

// SubnetOverlapError is returned when a new subnet overlaps an existing one.
type SubnetOverlapError struct {
	ConflictingCIDR string
}

func (e *SubnetOverlapError) Error() string {
	return fmt.Sprintf("subnet overlaps with existing subnet %s", e.ConflictingCIDR)
}

// ValidateCIDR validates a CIDR notation.
func ValidateCIDR(address string, prefixLength int) error {
	if prefixLength < 0 || prefixLength > 128 {
		return fmt.Errorf("invalid prefix length: %d", prefixLength)
	}
	if net.ParseIP(address) == nil {
		return fmt.Errorf("invalid network address: %s", address)
	}
	return nil
}

func (s *IPAMService) checkOverlap(ctx context.Context, networkID int64, networkAddress string, prefixLength int, excludeID int64) error {
	allowed, _ := s.config.GetCtx(ctx, "allow_subnet_overlaps")
	if allowed == "true" {
		return nil
	}

	newCIDR := fmt.Sprintf("%s/%d", networkAddress, prefixLength)
	_, newNet, err := net.ParseCIDR(newCIDR)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %s", newCIDR)
	}

	existing, err := s.repo.ListSubnetsBySection(ctx, networkID)
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

// OverlapPair represents two overlapping subnets.
type OverlapPair struct {
	SubnetA *models.Subnet `json:"subnet_a"`
	SubnetB *models.Subnet `json:"subnet_b"`
}

// OverlapReport returns all overlapping subnet pairs across all sections.
func (s *IPAMService) OverlapReport(ctx context.Context) ([]*OverlapPair, error) {
	all, err := s.repo.ListAllSubnets(ctx)
	if err != nil {
		return nil, err
	}

	bySec := make(map[int64][]*models.Subnet)
	for _, sub := range all {
		bySec[sub.NetworkID] = append(bySec[sub.NetworkID], sub)
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

// broadcastAddr computes the broadcast / last address of a network.
func broadcastAddr(n *net.IPNet) string {
	if ip4 := n.IP.To4(); ip4 != nil {
		mask := n.Mask
		broadcast := make(net.IP, 4)
		for i := 0; i < 4; i++ {
			broadcast[i] = ip4[i] | ^mask[i]
		}
		return broadcast.String()
	}
	ip6 := n.IP.To16()
	mask := n.Mask
	last := make(net.IP, 16)
	for i := 0; i < 16; i++ {
		last[i] = ip6[i] | ^mask[i]
	}
	return last.String()
}

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

func (s *IPAMService) validateVLANVRFConsistency(ctx context.Context, vlanID int64, subnetVRFID *int64) error {
	vlan, err := s.repo.GetVLANByID(ctx, vlanID)
	if err != nil {
		return fmt.Errorf("VLAN not found")
	}
	if vlan.DomainID == nil {
		return nil
	}
	if vlan.VRFID != nil {
		if subnetVRFID == nil || *vlan.VRFID != *subnetVRFID {
			return fmt.Errorf("subnet VRF does not match VLAN VRF (domain %d requires VRF %d)", *vlan.DomainID, *vlan.VRFID)
		}
	}
	return nil
}

func (s *IPAMService) GetVLANSubnets(ctx context.Context, vlanID int64) ([]*models.Subnet, error) {
	if vlanID <= 0 {
		return nil, fmt.Errorf("invalid VLAN ID")
	}
	return s.repo.GetVLANSubnets(ctx, vlanID)
}

func (s *IPAMService) AssignSubnetToVLAN(ctx context.Context, vlanID, subnetID int64) (*models.Subnet, error) {
	if vlanID <= 0 {
		return nil, fmt.Errorf("invalid VLAN ID")
	}
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}
	if _, err := s.repo.GetVLANByID(ctx, vlanID); err != nil {
		return nil, fmt.Errorf("VLAN not found")
	}
	if err := s.validateVLANVRFConsistency(ctx, vlanID, nil); err != nil {
		return nil, err
	}
	return s.repo.AssignSubnetToVLAN(ctx, subnetID, &vlanID)
}

func (s *IPAMService) RemoveSubnetFromVLAN(ctx context.Context, vlanID, subnetID int64) (*models.Subnet, error) {
	if vlanID <= 0 {
		return nil, fmt.Errorf("invalid VLAN ID")
	}
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}
	subnet, err := s.repo.GetSubnetByID(ctx, subnetID)
	if err != nil {
		return nil, err
	}
	if subnet.VLANID == nil || *subnet.VLANID != vlanID {
		return nil, fmt.Errorf("subnet is not assigned to this VLAN")
	}
	return s.repo.AssignSubnetToVLAN(ctx, subnetID, nil)
}

func (s *IPAMService) CreateSubnet(ctx context.Context, networkID int64, networkAddress string, prefixLength int, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64, customFields ...map[string]*string) (*models.Subnet, error) {
	if networkID <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}

	if err := ValidateCIDR(networkAddress, prefixLength); err != nil {
		return nil, err
	}

	if gateway != nil && *gateway != "" {
		if err := validateGatewayInCIDR(*gateway, networkAddress, prefixLength); err != nil {
			return nil, err
		}
	} else {
		gateway = nil
	}

	if err := s.checkOverlap(ctx, networkID, networkAddress, prefixLength, 0); err != nil {
		return nil, err
	}

	if !autoFirst {
		if v, _ := s.config.GetCtx(ctx, "default_auto_reserve_first"); v == "true" {
			autoFirst = true
		}
	}
	if !autoLast {
		if v, _ := s.config.GetCtx(ctx, "default_auto_reserve_last"); v == "true" {
			autoLast = true
		}
	}

	if vlanID != nil {
		if err := s.validateVLANVRFConsistency(ctx, *vlanID, nil); err != nil {
			return nil, err
		}
	}

	subnet, err := s.repo.CreateSubnetWithVLAN(ctx, networkID, networkAddress, prefixLength, description, gateway, autoFirst, autoLast, locationID, nameserverID, vlanID)
	if err != nil {
		return nil, err
	}

	cidr := fmt.Sprintf("%s/%d", networkAddress, prefixLength)
	_, ipNet, _ := net.ParseCIDR(cidr)
	if autoFirst && ipNet != nil {
		networkIP := ipNet.IP.String()
		if _, err := s.repo.CreateIPAddress(ctx, subnet.ID, networkIP, "", "reserved", nil, nil, nil, nil); err != nil {
			slog.Warn("auto-reserve first IP failed", "subnet_id", subnet.ID, "ip", networkIP, "error", err)
		}
	}
	if autoLast && ipNet != nil {
		bcastIP := broadcastAddr(ipNet)
		if _, err := s.repo.CreateIPAddress(ctx, subnet.ID, bcastIP, "", "reserved", nil, nil, nil, nil); err != nil {
			slog.Warn("auto-reserve last IP failed", "subnet_id", subnet.ID, "ip", bcastIP, "error", err)
		}
	}

	if len(customFields) > 0 && customFields[0] != nil {
		_ = s.setCustomFieldValues(ctx, "subnet", subnet.ID, customFields[0])
		subnet.CustomFields, _ = s.repo.GetCustomFieldValues(ctx, "subnet", subnet.ID)
	}

	return subnet, nil
}

func (s *IPAMService) GetSubnet(ctx context.Context, id int64) (*models.Subnet, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}
	subnet, err := s.repo.GetSubnetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	subnet.CustomFields, _ = s.repo.GetCustomFieldValues(ctx, "subnet", id)
	return subnet, nil
}

func (s *IPAMService) ListSubnets(ctx context.Context, networkID int64) ([]*models.Subnet, error) {
	if networkID <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}
	return s.repo.ListSubnetsBySection(ctx, networkID)
}

func (s *IPAMService) UpdateSubnet(ctx context.Context, id int64, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64, customFields map[string]*string, technitiumScopeName ...string) (*models.Subnet, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	if gateway != nil && *gateway != "" {
		existing, err := s.repo.GetSubnetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if err := validateGatewayInCIDR(*gateway, existing.NetworkAddress, existing.PrefixLength); err != nil {
			return nil, err
		}
	} else {
		gateway = nil
	}

	if vlanID != nil {
		if err := s.validateVLANVRFConsistency(ctx, *vlanID, nil); err != nil {
			return nil, err
		}
	}

	scopeName := ""
	if len(technitiumScopeName) > 0 {
		scopeName = technitiumScopeName[0]
	}
	subnet, err := s.repo.UpdateSubnetWithVLAN(ctx, id, description, gateway, autoFirst, autoLast, locationID, nameserverID, vlanID, scopeName)
	if err != nil {
		return nil, err
	}

	if customFields != nil {
		_ = s.setCustomFieldValues(ctx, "subnet", subnet.ID, customFields)
	}
	subnet.CustomFields, _ = s.repo.GetCustomFieldValues(ctx, "subnet", subnet.ID)
	return subnet, nil
}

func (s *IPAMService) ListSubnetsByLocation(ctx context.Context, locationID int64) ([]*models.Subnet, error) {
	if locationID <= 0 {
		return nil, fmt.Errorf("invalid location ID")
	}
	return s.repo.ListSubnetsByLocation(ctx, locationID)
}

func (s *IPAMService) DeleteSubnet(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid subnet ID")
	}
	return s.repo.DeleteSubnet(ctx, id)
}

// ---- IP Addresses -----------------------------------------------------------

func (s *IPAMService) CreateIPAddress(ctx context.Context, subnetID int64, address, hostname string, status string, tagID *int64, macAddress, ptrRecord, dnsName *string, customFields ...map[string]*string) (*models.IPAddress, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	parsedIP := net.ParseIP(address)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", address)
	}

	validStatuses := map[string]bool{"available": true, "assigned": true, "reserved": true}
	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid IP status: %s", status)
	}

	subnet, err := s.repo.GetSubnetByID(ctx, subnetID)
	if err != nil {
		return nil, fmt.Errorf("subnet not found: %w", err)
	}
	_, ipNet, err := net.ParseCIDR(fmt.Sprintf("%s/%d", subnet.NetworkAddress, subnet.PrefixLength))
	if err != nil {
		return nil, fmt.Errorf("invalid subnet CIDR: %w", err)
	}
	if !ipNet.Contains(parsedIP) {
		return nil, fmt.Errorf("IP address %s is not within subnet %s/%d", address, subnet.NetworkAddress, subnet.PrefixLength)
	}

	if macAddress != nil && *macAddress != "" {
		normalized, err := utils.NormalizeMAC(*macAddress)
		if err != nil {
			return nil, err
		}
		macAddress = &normalized
	}

	ip, err := s.repo.CreateIPAddress(ctx, subnetID, address, hostname, status, tagID, macAddress, ptrRecord, dnsName)
	if err != nil {
		return nil, normalizeCreateIPAddressError(err, address)
	}

	if len(customFields) > 0 && customFields[0] != nil {
		_ = s.setCustomFieldValues(ctx, "ip_address", ip.ID, customFields[0])
		ip.CustomFields, _ = s.repo.GetCustomFieldValues(ctx, "ip_address", ip.ID)
	}

	return ip, nil
}

func normalizeCreateIPAddressError(err error, address string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "ip_addresses_subnet_id_address_key" {
		return fmt.Errorf("IP address %s already exists in this subnet", address)
	}
	return err
}

func (s *IPAMService) UpdateIPAddressMeta(ctx context.Context, id int64, hostname string, tagID *int64, macAddress, ptrRecord, dnsName *string, customFields ...map[string]*string) (*models.IPAddress, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid IP address ID")
	}

	if macAddress != nil && *macAddress != "" {
		normalized, err := utils.NormalizeMAC(*macAddress)
		if err != nil {
			return nil, err
		}
		macAddress = &normalized
	}

	ip, err := s.repo.UpdateIPAddressFull(ctx, id, hostname, tagID, macAddress, ptrRecord, dnsName)
	if err != nil {
		return nil, err
	}

	if len(customFields) > 0 && customFields[0] != nil {
		_ = s.setCustomFieldValues(ctx, "ip_address", ip.ID, customFields[0])
	}
	ip.CustomFields, _ = s.repo.GetCustomFieldValues(ctx, "ip_address", ip.ID)
	return ip, nil
}

func (s *IPAMService) QuickCreateIPAddress(ctx context.Context, addr string) (*models.IPAddress, error) {
	subnet, err := s.repo.FindSubnetForIP(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("no subnet found for %s", addr)
	}
	return s.CreateIPAddress(ctx, subnet.ID, addr, "", "assigned", nil, nil, nil, nil)
}

func (s *IPAMService) GetIPAddress(ctx context.Context, id int64) (*models.IPAddress, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid IP address ID")
	}
	ip, err := s.repo.GetIPAddressByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ip.CustomFields, _ = s.repo.GetCustomFieldValues(ctx, "ip_address", id)
	return ip, nil
}

func (s *IPAMService) ListIPAddresses(ctx context.Context, subnetID int64) ([]*models.IPAddress, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}
	return s.repo.ListIPAddressesBySubnet(ctx, subnetID)
}

func (s *IPAMService) AssignIPAddress(ctx context.Context, id int64, deviceID *int64) (*models.IPAddress, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid IP address ID")
	}
	return s.repo.UpdateIPAddressStatus(ctx, id, "assigned", deviceID)
}

func (s *IPAMService) ReleaseIPAddress(ctx context.Context, id int64) (*models.IPAddress, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid IP address ID")
	}
	ip, err := s.repo.GetIPAddressByID(ctx, id)
	if err != nil {
		return nil, err
	}
	result, err := s.repo.UpdateIPAddressStatus(ctx, id, "available", nil)
	if err == nil {
		go s.dns.RemoveIPFromDNS(ctx, ip)
	}
	return result, err
}

func (s *IPAMService) DeleteIPAddress(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid IP address ID")
	}
	ip, err := s.repo.GetIPAddressByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteIPAddress(ctx, id); err != nil {
		return err
	}
	go s.dns.RemoveIPFromDNS(ctx, ip)
	return nil
}

func (s *IPAMService) BulkDeleteIPAddresses(ctx context.Context, ids []int64) (int, error) {
	ips := make([]*models.IPAddress, 0, len(ids))
	for _, id := range ids {
		ip, err := s.repo.GetIPAddressByID(ctx, id)
		if err == nil {
			ips = append(ips, ip)
		}
	}

	deleted, err := s.repo.BulkDeleteIPAddresses(ctx, ids)
	for _, ip := range ips {
		go s.dns.RemoveIPFromDNS(ctx, ip)
	}
	return len(deleted), err
}

func (s *IPAMService) FindNextAvailableIP(ctx context.Context, subnetID int64) (*models.IPAddress, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	availableIPs, err := s.repo.ListAvailableIPsBySubnet(ctx, subnetID)
	if err != nil {
		return nil, err
	}

	if len(availableIPs) == 0 {
		return nil, fmt.Errorf("no available IP addresses in subnet")
	}

	return availableIPs[0], nil
}

func (s *IPAMService) AllocateIPAddress(ctx context.Context, subnetID int64, deviceID *int64) (*models.IPAddress, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}
	return s.repo.AllocateIPAddress(ctx, subnetID, deviceID)
}

// SubnetUtilization represents utilization statistics for a subnet.
type SubnetUtilization struct {
	Total       int64   `json:"total"`
	Available   int64   `json:"available"`
	Assigned    int64   `json:"assigned"`
	Reserved    int64   `json:"reserved"`
	Utilization float64 `json:"utilization_percent"`
}

func (s *IPAMService) GetSubnetUtilization(ctx context.Context, subnetID int64) (*SubnetUtilization, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	total, available, assigned, reserved, err := s.repo.GetSubnetUtilizationCounts(ctx, subnetID)
	if err != nil {
		return nil, err
	}

	if total == 0 {
		return &SubnetUtilization{}, nil
	}

	return &SubnetUtilization{
		Total:       total,
		Available:   available,
		Assigned:    assigned,
		Reserved:    reserved,
		Utilization: (float64(assigned+reserved) / float64(total)) * 100,
	}, nil
}

func (s *IPAMService) AssignIPAddressWithLease(ctx context.Context, id int64, deviceID *int64, leaseDurationDays int) (*models.IPAddress, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid IP address ID")
	}

	now := time.Now().UTC()
	assignedAtTime := now
	var expiresAtTime *time.Time

	if leaseDurationDays > 0 {
		expiresAt := now.AddDate(0, 0, leaseDurationDays)
		expiresAtTime = &expiresAt
	}

	return s.repo.UpdateIPAddressWithLease(ctx, id, "assigned", deviceID, &assignedAtTime, expiresAtTime)
}

func (s *IPAMService) IsIPLeaseExpired(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid IP address ID")
	}

	ip, err := s.GetIPAddress(ctx, id)
	if err != nil {
		return false, err
	}

	if ip.ExpiresAt == nil {
		return false, nil
	}

	return time.Now().After(*ip.ExpiresAt), nil
}

func (s *IPAMService) ReleaseExpiredLease(ctx context.Context, id int64) (*models.IPAddress, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid IP address ID")
	}

	expired, err := s.IsIPLeaseExpired(ctx, id)
	if err != nil {
		return nil, err
	}

	if !expired {
		return nil, fmt.Errorf("IP lease has not expired")
	}

	return s.ReleaseIPAddress(ctx, id)
}

// ---- VRFs -------------------------------------------------------------------

func (s *IPAMService) CreateVRF(ctx context.Context, name, rd, description string) (*models.VRF, error) {
	if name == "" {
		return nil, fmt.Errorf("VRF name is required")
	}
	return s.repo.CreateVRF(ctx, name, rd, description)
}

func (s *IPAMService) GetVRF(ctx context.Context, id int64) (*models.VRF, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VRF ID")
	}
	return s.repo.GetVRFByID(ctx, id)
}

func (s *IPAMService) ListVRFs(ctx context.Context) ([]*models.VRF, error) {
	return s.repo.ListAllVRFs(ctx)
}

func (s *IPAMService) UpdateVRF(ctx context.Context, id int64, name, rd, description string) (*models.VRF, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VRF ID")
	}
	if name == "" {
		return nil, fmt.Errorf("VRF name is required")
	}
	return s.repo.UpdateVRF(ctx, id, name, rd, description)
}

func (s *IPAMService) DeleteVRF(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid VRF ID")
	}
	return s.repo.DeleteVRF(ctx, id)
}

// ---- VLANs ------------------------------------------------------------------

func (s *IPAMService) CreateVLAN(ctx context.Context, vrfID *int64, domainID *int64, groupID *int64, vlanID int, name, description string) (*models.VLAN, error) {
	if vlanID < 0 || vlanID > 4094 {
		return nil, fmt.Errorf("VLAN ID must be between 0 and 4094")
	}
	if name == "" {
		return nil, fmt.Errorf("VLAN name is required")
	}
	return s.repo.CreateVLAN(ctx, vrfID, domainID, groupID, vlanID, name, description)
}

func (s *IPAMService) GetVLAN(ctx context.Context, id int64) (*models.VLAN, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VLAN ID")
	}
	return s.repo.GetVLANByID(ctx, id)
}

func (s *IPAMService) ListVLANs(ctx context.Context) ([]*models.VLAN, error) {
	return s.repo.ListAllVLANs(ctx)
}

func (s *IPAMService) ListVLANsByVRF(ctx context.Context, vrfID int64) ([]*models.VLAN, error) {
	if vrfID <= 0 {
		return nil, fmt.Errorf("invalid VRF ID")
	}
	return s.repo.ListVLANsByVRF(ctx, vrfID)
}

func (s *IPAMService) UpdateVLAN(ctx context.Context, id int64, domainID *int64, groupID *int64, name, description string) (*models.VLAN, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VLAN ID")
	}
	if name == "" {
		return nil, fmt.Errorf("VLAN name is required")
	}
	return s.repo.UpdateVLAN(ctx, id, domainID, groupID, name, description)
}

func (s *IPAMService) DeleteVLAN(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid VLAN ID")
	}
	return s.repo.DeleteVLAN(ctx, id)
}

func (s *IPAMService) CreateVLANDomain(ctx context.Context, name string, description *string) (*models.VLANDomain, error) {
	if name == "" {
		return nil, fmt.Errorf("VLAN domain name is required")
	}
	return s.repo.CreateVLANDomain(ctx, name, description)
}

func (s *IPAMService) GetVLANDomain(ctx context.Context, id int64) (*models.VLANDomain, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VLAN domain ID")
	}
	return s.repo.GetVLANDomainByID(ctx, id)
}

func (s *IPAMService) ListVLANDomains(ctx context.Context) ([]*models.VLANDomain, error) {
	return s.repo.ListVLANDomains(ctx)
}

func (s *IPAMService) UpdateVLANDomain(ctx context.Context, id int64, name string, description *string) (*models.VLANDomain, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VLAN domain ID")
	}
	if name == "" {
		return nil, fmt.Errorf("VLAN domain name is required")
	}
	return s.repo.UpdateVLANDomain(ctx, id, name, description)
}

func (s *IPAMService) DeleteVLANDomain(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid VLAN domain ID")
	}
	return s.repo.DeleteVLANDomain(ctx, id)
}

func (s *IPAMService) CreateVLANGroup(ctx context.Context, name string, description *string, colour *string) (*models.VLANGroup, error) {
	if name == "" {
		return nil, fmt.Errorf("VLAN group name is required")
	}
	return s.repo.CreateVLANGroup(ctx, name, description, colour)
}

func (s *IPAMService) GetVLANGroup(ctx context.Context, id int64) (*models.VLANGroup, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VLAN group ID")
	}
	return s.repo.GetVLANGroupByID(ctx, id)
}

func (s *IPAMService) ListVLANGroups(ctx context.Context) ([]*models.VLANGroup, error) {
	return s.repo.ListVLANGroups(ctx)
}

func (s *IPAMService) UpdateVLANGroup(ctx context.Context, id int64, name string, description *string, colour *string) (*models.VLANGroup, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VLAN group ID")
	}
	if name == "" {
		return nil, fmt.Errorf("VLAN group name is required")
	}
	return s.repo.UpdateVLANGroup(ctx, id, name, description, colour)
}

func (s *IPAMService) DeleteVLANGroup(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid VLAN group ID")
	}
	return s.repo.DeleteVLANGroup(ctx, id)
}

func (s *IPAMService) GetVLANUsageReport(ctx context.Context) (*models.VLANUsageReport, error) {
	entries, err := s.repo.GetVLANUsageReport(ctx)
	if err != nil {
		return nil, err
	}
	return &models.VLANUsageReport{
		Entries:     entries,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ---- Tags -------------------------------------------------------------------

var (
	ErrSystemTag = errors.New("cannot modify system tag")
	ErrTagInUse  = errors.New("tag is in use")
)

func (s *IPAMService) ListIPTags(ctx context.Context) ([]*models.IPTag, error) {
	return s.repo.ListIPTags(ctx)
}

func (s *IPAMService) GetIPTag(ctx context.Context, id int64) (*models.IPTag, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid tag ID")
	}
	tag, err := s.repo.GetIPTagByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("tag %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return tag, nil
}

func (s *IPAMService) CreateIPTag(ctx context.Context, name, colour string, description *string) (*models.IPTag, error) {
	if name == "" {
		return nil, fmt.Errorf("tag name is required")
	}
	if colour == "" {
		colour = "#6B7280"
	}
	return s.repo.CreateIPTag(ctx, name, colour, description)
}

func (s *IPAMService) UpdateIPTag(ctx context.Context, id int64, name, colour string, description *string) (*models.IPTag, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid tag ID")
	}
	if name == "" {
		return nil, fmt.Errorf("tag name is required")
	}
	if colour == "" {
		colour = "#6B7280"
	}
	return s.repo.UpdateIPTag(ctx, id, name, colour, description)
}

func (s *IPAMService) DeleteIPTag(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid tag ID")
	}
	if err := s.repo.DeleteIPTag(ctx, id); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			return fmt.Errorf("tag %d: %w", id, ErrNotFound)
		}
		if strings.Contains(msg, "system tag") {
			return fmt.Errorf("tag %d: %w", id, ErrSystemTag)
		}
		if strings.Contains(msg, "in use") {
			return fmt.Errorf("tag %d: %w", id, ErrTagInUse)
		}
		return err
	}
	return nil
}

// ---- Search -----------------------------------------------------------------

// GlobalSearchResult holds results from a cross-entity search.
type GlobalSearchResult struct {
	Sections []*models.Network `json:"networks"`
	Subnets  []*models.Subnet  `json:"subnets"`
	Devices  []*models.Device  `json:"devices"`
}

const (
	DefaultLimit  = 50
	MaxLimit      = 500
	DefaultOffset = 0
)

func (s *IPAMService) GlobalSearch(ctx context.Context, query string, limit int64) (*GlobalSearchResult, error) {
	empty := &GlobalSearchResult{
		Sections: make([]*models.Network, 0),
		Subnets:  make([]*models.Subnet, 0),
		Devices:  make([]*models.Device, 0),
	}
	if query == "" {
		return empty, nil
	}
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		result = empty
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		secs, err := s.repo.SearchNetworks(ctx, query, limit, 0)
		if err == nil && len(secs) > 0 {
			mu.Lock()
			result.Sections = secs
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		subs, err := s.repo.GlobalSearchSubnets(ctx, query, limit)
		if err == nil && len(subs) > 0 {
			mu.Lock()
			result.Subnets = subs
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		devs, err := s.repo.SearchDevicesWithCustomFields(ctx, &repository.DeviceSearchFilter{Query: query}, nil)
		if err == nil && len(devs) > 0 {
			if int64(len(devs)) > limit {
				devs = devs[:limit]
			}
			mu.Lock()
			result.Devices = devs
			mu.Unlock()
		}
	}()
	wg.Wait()

	return result, nil
}

func (s *IPAMService) SearchNetworks(ctx context.Context, query string, limit, offset int64) ([]*models.Network, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if offset < 0 {
		offset = DefaultOffset
	}

	return s.repo.SearchNetworks(ctx, query, limit, offset)
}

func (s *IPAMService) SearchSubnets(ctx context.Context, networkID int64, query string, limit, offset int64, cfFilters ...map[string]string) ([]*models.Subnet, error) {
	if networkID <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if offset < 0 {
		offset = DefaultOffset
	}

	var cf map[string]string
	if len(cfFilters) > 0 {
		cf = cfFilters[0]
	}
	return s.repo.SearchSubnetsWithCustomFields(ctx, networkID, query, limit, offset, cf)
}

// IPSearchOptions holds additional search filters for IP address search.
type IPSearchOptions struct {
	TagID          *int64
	MACAddress     string
	PTRRecord      string
	IsAssigned     *bool
	LastSeenAfter  *time.Time
	LastSeenBefore *time.Time
	CustomFields   map[string]string
}

func (s *IPAMService) SearchIPAddresses(ctx context.Context, subnetID int64, query string, status string, limit, offset int64, opts ...IPSearchOptions) ([]*models.IPAddress, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	validStatuses := map[string]bool{"available": true, "assigned": true, "reserved": true}
	if status != "" && !validStatuses[status] {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if offset < 0 {
		offset = DefaultOffset
	}

	var repoFilter repository.IPSearchFilter
	var cfFilters map[string]string
	if len(opts) > 0 {
		o := opts[0]
		repoFilter = repository.IPSearchFilter{
			TagID:          o.TagID,
			MACAddress:     o.MACAddress,
			PTRRecord:      o.PTRRecord,
			IsAssigned:     o.IsAssigned,
			LastSeenAfter:  o.LastSeenAfter,
			LastSeenBefore: o.LastSeenBefore,
		}
		cfFilters = o.CustomFields
	}

	return s.repo.SearchIPAddressesWithCustomFields(ctx, subnetID, query, status, limit, offset, repoFilter, cfFilters)
}

func (s *IPAMService) SearchIPAddressesGlobal(ctx context.Context, query string) ([]*models.IPAddress, error) {
	if query == "" {
		return []*models.IPAddress{}, nil
	}
	return s.repo.SearchIPAddressesGlobal(ctx, query, 20)
}

// ---- Dashboard / Pagination -------------------------------------------------

func (s *IPAMService) GetDashboardSummary(ctx context.Context) (*models.DashboardSummary, error) {
	if summary, ok := s.summaryCache.get("summary"); ok {
		return cloneDashboardSummary(summary), nil
	}

	summary, err := s.repo.GetDashboardSummary(ctx)
	if err != nil {
		return nil, err
	}
	s.summaryCache.set("summary", cloneDashboardSummary(summary))
	return cloneDashboardSummary(summary), nil
}

func (s *IPAMService) GetDashboardRecentActivity(ctx context.Context) ([]*models.DashboardActivity, error) {
	if activities, ok := s.activityCache.get("recent"); ok {
		return cloneDashboardActivities(activities), nil
	}

	activities, err := s.repo.GetDashboardRecentActivity(ctx)
	if err != nil {
		return nil, err
	}
	s.activityCache.set("recent", cloneDashboardActivities(activities))
	return cloneDashboardActivities(activities), nil
}

func cloneDashboardSummary(summary *models.DashboardSummary) *models.DashboardSummary {
	if summary == nil {
		return nil
	}
	out := *summary
	out.TopSubnets = append([]models.SubnetUtilization(nil), summary.TopSubnets...)
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

func (s *IPAMService) GetSubnetTree(ctx context.Context, networkID int64) ([]models.SubnetTreeNode, error) {
	if networkID <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}

	nodes, err := s.repo.GetSubnetTreeBySection(ctx, networkID)
	if err != nil {
		return nil, err
	}

	return buildTree(nodes), nil
}

func buildTree(flat []models.SubnetTreeNode) []models.SubnetTreeNode {
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

	sort.Slice(items, func(i, j int) bool {
		pi, _ := items[i].net.Mask.Size()
		pj, _ := items[j].net.Mask.Size()
		if pi != pj {
			return pi < pj
		}
		return items[i].net.IP.String() < items[j].net.IP.String()
	})

	parentIdx := make([]int, len(items))
	for i := range parentIdx {
		parentIdx[i] = -1
	}

	for i := 1; i < len(items); i++ {
		for j := i - 1; j >= 0; j-- {
			pj, _ := items[j].net.Mask.Size()
			pi, _ := items[i].net.Mask.Size()
			if pj < pi && items[j].net.Contains(items[i].net.IP) {
				parentIdx[i] = j
				break
			}
		}
	}

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

func (s *IPAMService) ListNetworksPaginated(ctx context.Context, page, limit int) ([]*models.Network, int64, error) {
	return s.ListNetworksPaginatedWithOptions(ctx, page, limit, repository.ListOptions{})
}

func (s *IPAMService) ListNetworksPaginatedWithOptions(ctx context.Context, page, limit int, opts repository.ListOptions) ([]*models.Network, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	opts.Limit = limit
	opts.Offset = (page - 1) * limit
	return s.repo.ListNetworksPaginatedWithOptions(ctx, opts)
}

func (s *IPAMService) ListSubnetsPaginated(ctx context.Context, networkID int64, page, limit int) ([]*models.Subnet, int64, error) {
	return s.ListSubnetsPaginatedWithOptions(ctx, networkID, page, limit, repository.ListOptions{})
}

func (s *IPAMService) ListSubnetsPaginatedWithOptions(ctx context.Context, networkID int64, page, limit int, opts repository.ListOptions) ([]*models.Subnet, int64, error) {
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
	return s.repo.ListSubnetsBySectionPaginatedWithOptions(ctx, networkID, opts)
}

func (s *IPAMService) ListIPAddressesPaginated(ctx context.Context, subnetID int64, page, limit int) ([]*models.IPAddress, int64, error) {
	return s.ListIPAddressesPaginatedWithOptions(ctx, subnetID, page, limit, repository.ListOptions{})
}

func (s *IPAMService) ListIPAddressesFullRange(ctx context.Context, subnetID int64, page, limit int) ([]*models.IPAddress, int64, error) {
	if subnetID <= 0 {
		return nil, 0, fmt.Errorf("invalid subnet ID")
	}
	subnet, err := s.repo.GetSubnetByID(ctx, subnetID)
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
	return s.repo.ListIPAddressesFullRange(ctx, subnetID, subnet.NetworkAddress, subnet.PrefixLength, offset, limit)
}

func (s *IPAMService) ListIPAddressesPaginatedWithOptions(ctx context.Context, subnetID int64, page, limit int, opts repository.ListOptions) ([]*models.IPAddress, int64, error) {
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
	return s.repo.ListIPAddressesBySubnetPaginatedWithOptions(ctx, subnetID, opts)
}

func (s *IPAMService) ListVLANsPaginated(ctx context.Context, page, limit int) ([]*models.VLAN, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repo.ListVLANsPaginated(ctx, limit, offset)
}

func (s *IPAMService) ListVRFsPaginated(ctx context.Context, page, limit int) ([]*models.VRF, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repo.ListVRFsPaginated(ctx, limit, offset)
}

// ---- IPv6 Delegations -------------------------------------------------------

func (s *IPAMService) ListDelegations(ctx context.Context, subnetID int64) ([]*models.IPv6Delegation, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}
	delegations, err := s.repo.ListIPv6DelegationsBySubnet(ctx, subnetID)
	if err != nil {
		return nil, err
	}
	for _, d := range delegations {
		d.IsExpired = isExpiredNow(d.ExpiresAt)
	}
	return delegations, nil
}

func (s *IPAMService) CreateDelegation(ctx context.Context, d *models.IPv6Delegation) (*models.IPv6Delegation, error) {
	if d.ParentSubnetID <= 0 {
		return nil, fmt.Errorf("invalid parent subnet ID")
	}
	if d.DelegatedPrefix == "" {
		return nil, fmt.Errorf("delegated prefix is required")
	}
	result, err := s.repo.CreateIPv6Delegation(ctx, d)
	if err != nil {
		return nil, err
	}
	result.IsExpired = isExpiredNow(result.ExpiresAt)
	return result, nil
}

func (s *IPAMService) UpdateDelegation(ctx context.Context, id int64, d *models.IPv6Delegation) (*models.IPv6Delegation, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid delegation ID")
	}
	if d.DelegatedPrefix == "" {
		return nil, fmt.Errorf("delegated prefix is required")
	}
	result, err := s.repo.UpdateIPv6Delegation(ctx, id, d)
	if err != nil {
		return nil, err
	}
	result.IsExpired = isExpiredNow(result.ExpiresAt)
	return result, nil
}

func (s *IPAMService) DeleteDelegation(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid delegation ID")
	}
	return s.repo.DeleteIPv6Delegation(ctx, id)
}

func isExpiredNow(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return false
	}
	return time.Now().After(*expiresAt)
}

// ---- Subnet Tools -----------------------------------------------------------

// SplitSubnet splits a subnet into 2^(newPrefixLen-currentPrefix) child subnets.
func (s *IPAMService) SplitSubnet(ctx context.Context, subnetID int64, newPrefixLen int) ([]*models.Subnet, error) {
	parent, err := s.repo.GetSubnetByID(ctx, subnetID)
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

	parentCIDR := fmt.Sprintf("%s/%d", parent.NetworkAddress, parent.PrefixLength)
	_, parentNet, err := net.ParseCIDR(parentCIDR)
	if err != nil {
		return nil, fmt.Errorf("invalid parent CIDR: %w", err)
	}

	childSubnets := make([]*models.Subnet, 0, childCount)
	blockingAddrs := make([]string, 0, childCount*2)
	ip := cloneIP(parentNet.IP.To4())
	for i := 0; i < childCount; i++ {
		childNet := &net.IPNet{
			IP:   cloneIP(ip),
			Mask: net.CIDRMask(newPrefixLen, 32),
		}
		child := &models.Subnet{
			NetworkID:      parent.NetworkID,
			NetworkAddress: childNet.IP.String(),
			PrefixLength:   newPrefixLen,
			Description:    fmt.Sprintf("Split from %s", parentCIDR),
		}
		childSubnets = append(childSubnets, child)

		blockingAddrs = append(blockingAddrs, childNet.IP.String())
		broadcast := broadcastAddr(childNet)
		blockingAddrs = append(blockingAddrs, broadcast)

		incrementIP(ip, 32-newPrefixLen)
	}

	blockingIPs, err := s.repo.ListIPsAtAddresses(ctx, subnetID, blockingAddrs)
	if err != nil {
		return nil, fmt.Errorf("checking blocking IPs: %w", err)
	}
	if len(blockingIPs) > 0 {
		addrs := make([]string, len(blockingIPs))
		for i, ip := range blockingIPs {
			addrs[i] = ip.Address
		}
		sort.Strings(addrs)
		return nil, &SplitBlockedError{BlockingIPs: addrs}
	}

	if err := s.repo.SplitSubnet(ctx, subnetID, childSubnets); err != nil {
		return nil, fmt.Errorf("split transaction failed: %w", err)
	}

	return childSubnets, nil
}

func cloneIP(ip net.IP) net.IP {
	c := make(net.IP, len(ip))
	copy(c, ip)
	return c
}

func incrementIP(ip net.IP, hostBits int) {
	inc := uint32(1) << uint(hostBits)
	b := ip.To4()
	val := binary.BigEndian.Uint32(b)
	val += inc
	binary.BigEndian.PutUint32(b, val)
}

// SplitBlockedError is returned when split cannot proceed due to IPs on network/broadcast addresses.
type SplitBlockedError struct {
	BlockingIPs []string
}

func (e *SplitBlockedError) Error() string {
	return fmt.Sprintf("the following IPs fall on network or broadcast addresses of the new subnets and must be removed first: %s", strings.Join(e.BlockingIPs, ", "))
}

// MergeSubnets merges multiple subnets into a common supernet.
func (s *IPAMService) MergeSubnets(ctx context.Context, subnetIDs []int64) (*models.Subnet, error) {
	if len(subnetIDs) < 2 {
		return nil, fmt.Errorf("at least 2 subnets required for merge")
	}

	subnets := make([]*models.Subnet, 0, len(subnetIDs))
	for _, id := range subnetIDs {
		sub, err := s.repo.GetSubnetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("subnet %d not found: %w", id, err)
		}
		subnets = append(subnets, sub)
	}

	networkID := subnets[0].NetworkID
	prefixLen := subnets[0].PrefixLength
	for _, sub := range subnets[1:] {
		if sub.NetworkID != networkID {
			return nil, fmt.Errorf("all subnets must be in the same section")
		}
		if sub.PrefixLength != prefixLen {
			return nil, fmt.Errorf("all subnets must have the same prefix length")
		}
	}

	n := len(subnetIDs)
	if n == 0 || (n&(n-1)) != 0 {
		return nil, fmt.Errorf("number of subnets to merge must be a power of 2 (got %d)", n)
	}

	newPrefixLen := prefixLen - int(math.Log2(float64(n)))
	if newPrefixLen < 0 {
		return nil, fmt.Errorf("resulting prefix length would be negative")
	}

	var minIP net.IP
	for _, sub := range subnets {
		ip := net.ParseIP(sub.NetworkAddress).To4()
		if minIP == nil || ipLess(ip, minIP) {
			minIP = ip
		}
	}

	supernet := &net.IPNet{
		IP:   minIP.Mask(net.CIDRMask(newPrefixLen, 32)),
		Mask: net.CIDRMask(newPrefixLen, 32),
	}

	for _, sub := range subnets {
		subIP := net.ParseIP(sub.NetworkAddress).To4()
		if !supernet.Contains(subIP) {
			return nil, fmt.Errorf("subnet %s/%d is not contiguous with the others for merging", sub.NetworkAddress, sub.PrefixLength)
		}
	}

	parent := &models.Subnet{
		NetworkID:      networkID,
		NetworkAddress: supernet.IP.String(),
		PrefixLength:   newPrefixLen,
		Description:    fmt.Sprintf("Merged from %d subnets", len(subnetIDs)),
	}

	result, err := s.repo.MergeSubnets(ctx, subnetIDs, parent)
	if err != nil {
		return nil, fmt.Errorf("merge transaction failed: %w", err)
	}

	return result, nil
}

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
func (s *IPAMService) ResizeSubnet(ctx context.Context, subnetID int64, newPrefix string) (*models.Subnet, error) {
	existing, err := s.repo.GetSubnetByID(ctx, subnetID)
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

	if existing.ParentSubnetID != nil {
		parent, err := s.repo.GetSubnetByID(ctx, *existing.ParentSubnetID)
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
		outsideIPs, err := s.repo.ListIPsOutsideCIDR(ctx, subnetID, newNetworkAddr, newPrefixLen)
		if err != nil {
			return nil, fmt.Errorf("checking IPs outside new range: %w", err)
		}
		if len(outsideIPs) > 0 {
			return nil, &SubnetResizeConflictError{ConflictingIPs: outsideIPs}
		}
	} else {
		siblings, err := s.repo.ListSiblingSubnets(ctx, subnetID)
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
				return nil, &SubnetResizeConflictError{ConflictingSubnets: siblings}
			}
		}
	}

	result, err := s.repo.ResizeSubnet(ctx, subnetID, newNetworkAddr, newPrefixLen)
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
