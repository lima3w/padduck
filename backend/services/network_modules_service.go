package services

import (
	"context"
	"fmt"
	"net"
	"strings"

	"padduck/models"
	"padduck/repository"
)

// NetworkModulesService owns NAT rules, firewall zones, DHCP servers/leases,
// circuit providers/circuits, and BGP autonomous systems.
// Exposed via OpsManager.NetworkModules.
type NetworkModulesService struct {
	repo *repository.Repository
}

func NewNetworkModulesService(repo *repository.Repository) *NetworkModulesService {
	return &NetworkModulesService{repo: repo}
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func validateCIDRLike(value, field string) error {
	if ip := net.ParseIP(value); ip != nil {
		return nil
	}
	if _, _, err := net.ParseCIDR(value); err == nil {
		return nil
	}
	return fmt.Errorf("%s must be an IP address or CIDR", field)
}

// NAT rules

func (s *NetworkModulesService) ListNATRules(ctx context.Context) ([]*models.NATRule, error) {
	return s.repo.ListNATRules(ctx)
}

func (s *NetworkModulesService) GetNATRule(ctx context.Context, id int64) (*models.NATRule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid NAT rule ID")
	}
	return s.repo.GetNATRuleByID(ctx, id)
}

func (s *NetworkModulesService) CreateNATRule(ctx context.Context, req *repository.NATRuleParams) (*models.NATRule, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("NAT rule name is required")
	}
	if err := validateCIDRLike(req.InternalCIDR, "internal CIDR"); err != nil {
		return nil, err
	}
	if err := validateCIDRLike(req.ExternalCIDR, "external CIDR"); err != nil {
		return nil, err
	}
	req.Type = defaultString(req.Type, "static")
	req.Protocol = defaultString(req.Protocol, "any")
	req.Status = defaultString(req.Status, "active")
	return s.repo.CreateNATRule(ctx, req)
}

func (s *NetworkModulesService) UpdateNATRule(ctx context.Context, id int64, req *repository.NATRuleParams) (*models.NATRule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid NAT rule ID")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("NAT rule name is required")
	}
	if err := validateCIDRLike(req.InternalCIDR, "internal CIDR"); err != nil {
		return nil, err
	}
	if err := validateCIDRLike(req.ExternalCIDR, "external CIDR"); err != nil {
		return nil, err
	}
	req.Type = defaultString(req.Type, "static")
	req.Protocol = defaultString(req.Protocol, "any")
	req.Status = defaultString(req.Status, "active")
	return s.repo.UpdateNATRule(ctx, id, req)
}

func (s *NetworkModulesService) DeleteNATRule(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid NAT rule ID")
	}
	return s.repo.DeleteNATRule(ctx, id)
}

// Firewall zones

func (s *NetworkModulesService) ListFirewallZones(ctx context.Context) ([]*models.FirewallZone, error) {
	return s.repo.ListFirewallZones(ctx)
}

func (s *NetworkModulesService) GetFirewallZone(ctx context.Context, id int64) (*models.FirewallZone, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid firewall zone ID")
	}
	return s.repo.GetFirewallZoneByID(ctx, id)
}

func (s *NetworkModulesService) CreateFirewallZone(ctx context.Context, req *repository.FirewallZoneParams) (*models.FirewallZone, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("firewall zone name is required")
	}
	req.Color = defaultString(req.Color, "#2563eb")
	req.Status = defaultString(req.Status, "active")
	return s.repo.CreateFirewallZone(ctx, req)
}

func (s *NetworkModulesService) UpdateFirewallZone(ctx context.Context, id int64, req *repository.FirewallZoneParams) (*models.FirewallZone, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid firewall zone ID")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("firewall zone name is required")
	}
	req.Color = defaultString(req.Color, "#2563eb")
	req.Status = defaultString(req.Status, "active")
	return s.repo.UpdateFirewallZone(ctx, id, req)
}

func (s *NetworkModulesService) DeleteFirewallZone(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid firewall zone ID")
	}
	return s.repo.DeleteFirewallZone(ctx, id)
}

// Firewall zone mappings

func (s *NetworkModulesService) ListFirewallZoneMappings(ctx context.Context, zoneID int64) ([]*models.FirewallZoneMapping, error) {
	return s.repo.ListFirewallZoneMappings(ctx, zoneID)
}

func (s *NetworkModulesService) CreateFirewallZoneMapping(ctx context.Context, req *repository.FirewallZoneMappingParams) (*models.FirewallZoneMapping, error) {
	if err := validateFirewallZoneMapping(req); err != nil {
		return nil, err
	}
	req.Direction = defaultString(req.Direction, "both")
	req.Status = defaultString(req.Status, "active")
	return s.repo.CreateFirewallZoneMapping(ctx, req)
}

func (s *NetworkModulesService) UpdateFirewallZoneMapping(ctx context.Context, id int64, req *repository.FirewallZoneMappingParams) (*models.FirewallZoneMapping, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid firewall mapping ID")
	}
	if err := validateFirewallZoneMapping(req); err != nil {
		return nil, err
	}
	req.Direction = defaultString(req.Direction, "both")
	req.Status = defaultString(req.Status, "active")
	return s.repo.UpdateFirewallZoneMapping(ctx, id, req)
}

func validateFirewallZoneMapping(req *repository.FirewallZoneMappingParams) error {
	if req.ZoneID <= 0 {
		return fmt.Errorf("firewall zone is required")
	}
	if strings.TrimSpace(req.ObjectType) == "" {
		return fmt.Errorf("object type is required")
	}
	if req.ObjectType == "cidr" && strings.TrimSpace(req.CIDR) == "" {
		return fmt.Errorf("CIDR is required for CIDR mappings")
	}
	if req.ObjectID == nil && strings.TrimSpace(req.CIDR) == "" {
		return fmt.Errorf("object ID or CIDR is required")
	}
	if strings.TrimSpace(req.CIDR) != "" {
		if _, _, err := net.ParseCIDR(req.CIDR); err != nil {
			return fmt.Errorf("CIDR must be valid")
		}
	}
	return nil
}

func (s *NetworkModulesService) DeleteFirewallZoneMapping(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid firewall mapping ID")
	}
	return s.repo.DeleteFirewallZoneMapping(ctx, id)
}

// DHCP servers

func (s *NetworkModulesService) ListDHCPServers(ctx context.Context) ([]*models.DHCPServer, error) {
	return s.repo.ListDHCPServers(ctx)
}

func (s *NetworkModulesService) CreateDHCPServer(ctx context.Context, req *repository.DHCPServerParams) (*models.DHCPServer, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("DHCP server name is required")
	}
	if net.ParseIP(req.Address) == nil {
		return nil, fmt.Errorf("DHCP server address must be an IP address")
	}
	req.Status = defaultString(req.Status, "active")
	return s.repo.CreateDHCPServer(ctx, req)
}

func (s *NetworkModulesService) UpdateDHCPServer(ctx context.Context, id int64, req *repository.DHCPServerParams) (*models.DHCPServer, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid DHCP server ID")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("DHCP server name is required")
	}
	if net.ParseIP(req.Address) == nil {
		return nil, fmt.Errorf("DHCP server address must be an IP address")
	}
	req.Status = defaultString(req.Status, "active")
	return s.repo.UpdateDHCPServer(ctx, id, req)
}

func (s *NetworkModulesService) DeleteDHCPServer(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid DHCP server ID")
	}
	return s.repo.DeleteDHCPServer(ctx, id)
}

// DHCP leases

func (s *NetworkModulesService) ListDHCPLeases(ctx context.Context, serverID int64) ([]*models.DHCPLease, error) {
	return s.repo.ListDHCPLeases(ctx, serverID)
}

func (s *NetworkModulesService) CreateDHCPLease(ctx context.Context, req *repository.DHCPLeaseParams) (*models.DHCPLease, error) {
	if err := validateDHCPLease(req); err != nil {
		return nil, err
	}
	req.State = defaultString(req.State, "active")
	return s.repo.CreateDHCPLease(ctx, req)
}

func (s *NetworkModulesService) UpdateDHCPLease(ctx context.Context, id int64, req *repository.DHCPLeaseParams) (*models.DHCPLease, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid DHCP lease ID")
	}
	if err := validateDHCPLease(req); err != nil {
		return nil, err
	}
	req.State = defaultString(req.State, "active")
	return s.repo.UpdateDHCPLease(ctx, id, req)
}

func validateDHCPLease(req *repository.DHCPLeaseParams) error {
	if req.ServerID <= 0 {
		return fmt.Errorf("DHCP server is required")
	}
	if net.ParseIP(req.IPAddress) == nil {
		return fmt.Errorf("lease IP address must be valid")
	}
	if strings.TrimSpace(req.MACAddress) == "" {
		return fmt.Errorf("MAC address is required")
	}
	return nil
}

func (s *NetworkModulesService) DeleteDHCPLease(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid DHCP lease ID")
	}
	return s.repo.DeleteDHCPLease(ctx, id)
}

// Circuit providers

func (s *NetworkModulesService) ListCircuitProviders(ctx context.Context) ([]*models.CircuitProvider, error) {
	return s.repo.ListCircuitProviders(ctx)
}

func (s *NetworkModulesService) CreateCircuitProvider(ctx context.Context, req *repository.CircuitProviderParams) (*models.CircuitProvider, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("provider name is required")
	}
	return s.repo.CreateCircuitProvider(ctx, req)
}

func (s *NetworkModulesService) UpdateCircuitProvider(ctx context.Context, id int64, req *repository.CircuitProviderParams) (*models.CircuitProvider, error) {
	if id <= 0 || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("valid provider ID and name are required")
	}
	return s.repo.UpdateCircuitProvider(ctx, id, req)
}

func (s *NetworkModulesService) DeleteCircuitProvider(ctx context.Context, id int64) error {
	return s.repo.DeleteCircuitProvider(ctx, id)
}

// Physical circuits

func (s *NetworkModulesService) ListPhysicalCircuits(ctx context.Context) ([]*models.PhysicalCircuit, error) {
	return s.repo.ListPhysicalCircuits(ctx)
}

func (s *NetworkModulesService) CreatePhysicalCircuit(ctx context.Context, req *repository.PhysicalCircuitParams) (*models.PhysicalCircuit, error) {
	if req.ProviderID <= 0 || strings.TrimSpace(req.CircuitID) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("provider, circuit ID, and name are required")
	}
	req.Type = defaultString(req.Type, "ethernet")
	req.Status = defaultString(req.Status, "active")
	return s.repo.CreatePhysicalCircuit(ctx, req)
}

func (s *NetworkModulesService) UpdatePhysicalCircuit(ctx context.Context, id int64, req *repository.PhysicalCircuitParams) (*models.PhysicalCircuit, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid physical circuit ID")
	}
	if req.ProviderID <= 0 || strings.TrimSpace(req.CircuitID) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("provider, circuit ID, and name are required")
	}
	req.Type = defaultString(req.Type, "ethernet")
	req.Status = defaultString(req.Status, "active")
	return s.repo.UpdatePhysicalCircuit(ctx, id, req)
}

func (s *NetworkModulesService) DeletePhysicalCircuit(ctx context.Context, id int64) error {
	return s.repo.DeletePhysicalCircuit(ctx, id)
}

// Logical circuits

func (s *NetworkModulesService) ListLogicalCircuits(ctx context.Context) ([]*models.LogicalCircuit, error) {
	return s.repo.ListLogicalCircuits(ctx)
}

func (s *NetworkModulesService) CreateLogicalCircuit(ctx context.Context, req *repository.LogicalCircuitParams) (*models.LogicalCircuit, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("logical circuit name is required")
	}
	req.Type = defaultString(req.Type, "l2vpn")
	req.Status = defaultString(req.Status, "active")
	return s.repo.CreateLogicalCircuit(ctx, req)
}

func (s *NetworkModulesService) UpdateLogicalCircuit(ctx context.Context, id int64, req *repository.LogicalCircuitParams) (*models.LogicalCircuit, error) {
	if id <= 0 || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("valid logical circuit ID and name are required")
	}
	req.Type = defaultString(req.Type, "l2vpn")
	req.Status = defaultString(req.Status, "active")
	return s.repo.UpdateLogicalCircuit(ctx, id, req)
}

func (s *NetworkModulesService) DeleteLogicalCircuit(ctx context.Context, id int64) error {
	return s.repo.DeleteLogicalCircuit(ctx, id)
}

// BGP autonomous systems

func (s *NetworkModulesService) ListAutonomousSystems(ctx context.Context) ([]*models.AutonomousSystem, error) {
	return s.repo.ListAllAutonomousSystems(ctx)
}

func (s *NetworkModulesService) ListAutonomousSystemsPaginated(ctx context.Context, page, limit int) ([]*models.AutonomousSystem, int64, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	return s.repo.ListAutonomousSystemsPaginated(ctx, limit, offset)
}

func (s *NetworkModulesService) GetAutonomousSystem(ctx context.Context, id int64) (*models.AutonomousSystem, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid autonomous system ID")
	}
	return s.repo.GetAutonomousSystemByID(ctx, id)
}

func (s *NetworkModulesService) CreateAutonomousSystem(ctx context.Context, asn int64, name, description, asType, rir string) (*models.AutonomousSystem, error) {
	if asn <= 0 {
		return nil, fmt.Errorf("ASN must be a positive integer")
	}
	if asType != "internal" && asType != "external" {
		asType = "external"
	}
	return s.repo.CreateAutonomousSystem(ctx, asn, name, description, asType, rir)
}

func (s *NetworkModulesService) UpdateAutonomousSystem(ctx context.Context, id, asn int64, name, description, asType, rir string) (*models.AutonomousSystem, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid autonomous system ID")
	}
	if asn <= 0 {
		return nil, fmt.Errorf("ASN must be a positive integer")
	}
	if asType != "internal" && asType != "external" {
		asType = "external"
	}
	return s.repo.UpdateAutonomousSystem(ctx, id, asn, name, description, asType, rir)
}

func (s *NetworkModulesService) DeleteAutonomousSystem(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid autonomous system ID")
	}
	return s.repo.DeleteAutonomousSystem(ctx, id)
}
