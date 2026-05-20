package services

import (
	"context"
	"fmt"
	"net"
	"strings"

	"ipam-next/models"
	"ipam-next/repository"
)

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

func (s *Service) ListNATRules(ctx context.Context) ([]*models.NATRule, error) {
	return s.repository.ListNATRules(ctx)
}

func (s *Service) GetNATRule(ctx context.Context, id int64) (*models.NATRule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid NAT rule ID")
	}
	return s.repository.GetNATRuleByID(ctx, id)
}

func (s *Service) CreateNATRule(ctx context.Context, req *repository.NATRuleParams) (*models.NATRule, error) {
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
	return s.repository.CreateNATRule(ctx, req)
}

func (s *Service) UpdateNATRule(ctx context.Context, id int64, req *repository.NATRuleParams) (*models.NATRule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid NAT rule ID")
	}
	return s.CreateNATRuleWithID(ctx, id, req)
}

func (s *Service) CreateNATRuleWithID(ctx context.Context, id int64, req *repository.NATRuleParams) (*models.NATRule, error) {
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
	return s.repository.UpdateNATRule(ctx, id, req)
}

func (s *Service) DeleteNATRule(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid NAT rule ID")
	}
	return s.repository.DeleteNATRule(ctx, id)
}

func (s *Service) ListDHCPServers(ctx context.Context) ([]*models.DHCPServer, error) {
	return s.repository.ListDHCPServers(ctx)
}

func (s *Service) CreateDHCPServer(ctx context.Context, req *repository.DHCPServerParams) (*models.DHCPServer, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("DHCP server name is required")
	}
	if net.ParseIP(req.Address) == nil {
		return nil, fmt.Errorf("DHCP server address must be an IP address")
	}
	req.Status = defaultString(req.Status, "active")
	return s.repository.CreateDHCPServer(ctx, req)
}

func (s *Service) UpdateDHCPServer(ctx context.Context, id int64, req *repository.DHCPServerParams) (*models.DHCPServer, error) {
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
	return s.repository.UpdateDHCPServer(ctx, id, req)
}

func (s *Service) DeleteDHCPServer(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid DHCP server ID")
	}
	return s.repository.DeleteDHCPServer(ctx, id)
}

func (s *Service) ListDHCPLeases(ctx context.Context, serverID int64) ([]*models.DHCPLease, error) {
	return s.repository.ListDHCPLeases(ctx, serverID)
}

func (s *Service) CreateDHCPLease(ctx context.Context, req *repository.DHCPLeaseParams) (*models.DHCPLease, error) {
	if err := validateDHCPLease(req); err != nil {
		return nil, err
	}
	req.State = defaultString(req.State, "active")
	return s.repository.CreateDHCPLease(ctx, req)
}

func (s *Service) UpdateDHCPLease(ctx context.Context, id int64, req *repository.DHCPLeaseParams) (*models.DHCPLease, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid DHCP lease ID")
	}
	if err := validateDHCPLease(req); err != nil {
		return nil, err
	}
	req.State = defaultString(req.State, "active")
	return s.repository.UpdateDHCPLease(ctx, id, req)
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

func (s *Service) DeleteDHCPLease(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid DHCP lease ID")
	}
	return s.repository.DeleteDHCPLease(ctx, id)
}

func (s *Service) ListCircuitProviders(ctx context.Context) ([]*models.CircuitProvider, error) {
	return s.repository.ListCircuitProviders(ctx)
}

func (s *Service) CreateCircuitProvider(ctx context.Context, req *repository.CircuitProviderParams) (*models.CircuitProvider, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("provider name is required")
	}
	return s.repository.CreateCircuitProvider(ctx, req)
}

func (s *Service) UpdateCircuitProvider(ctx context.Context, id int64, req *repository.CircuitProviderParams) (*models.CircuitProvider, error) {
	if id <= 0 || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("valid provider ID and name are required")
	}
	return s.repository.UpdateCircuitProvider(ctx, id, req)
}

func (s *Service) DeleteCircuitProvider(ctx context.Context, id int64) error {
	return s.repository.DeleteCircuitProvider(ctx, id)
}

func (s *Service) ListPhysicalCircuits(ctx context.Context) ([]*models.PhysicalCircuit, error) {
	return s.repository.ListPhysicalCircuits(ctx)
}

func (s *Service) CreatePhysicalCircuit(ctx context.Context, req *repository.PhysicalCircuitParams) (*models.PhysicalCircuit, error) {
	if req.ProviderID <= 0 || strings.TrimSpace(req.CircuitID) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("provider, circuit ID, and name are required")
	}
	req.Type = defaultString(req.Type, "ethernet")
	req.Status = defaultString(req.Status, "active")
	return s.repository.CreatePhysicalCircuit(ctx, req)
}

func (s *Service) UpdatePhysicalCircuit(ctx context.Context, id int64, req *repository.PhysicalCircuitParams) (*models.PhysicalCircuit, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid physical circuit ID")
	}
	return s.CreatePhysicalCircuitWithID(ctx, id, req)
}

func (s *Service) CreatePhysicalCircuitWithID(ctx context.Context, id int64, req *repository.PhysicalCircuitParams) (*models.PhysicalCircuit, error) {
	if req.ProviderID <= 0 || strings.TrimSpace(req.CircuitID) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("provider, circuit ID, and name are required")
	}
	req.Type = defaultString(req.Type, "ethernet")
	req.Status = defaultString(req.Status, "active")
	return s.repository.UpdatePhysicalCircuit(ctx, id, req)
}

func (s *Service) DeletePhysicalCircuit(ctx context.Context, id int64) error {
	return s.repository.DeletePhysicalCircuit(ctx, id)
}

func (s *Service) ListLogicalCircuits(ctx context.Context) ([]*models.LogicalCircuit, error) {
	return s.repository.ListLogicalCircuits(ctx)
}

func (s *Service) CreateLogicalCircuit(ctx context.Context, req *repository.LogicalCircuitParams) (*models.LogicalCircuit, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("logical circuit name is required")
	}
	req.Type = defaultString(req.Type, "l2vpn")
	req.Status = defaultString(req.Status, "active")
	return s.repository.CreateLogicalCircuit(ctx, req)
}

func (s *Service) UpdateLogicalCircuit(ctx context.Context, id int64, req *repository.LogicalCircuitParams) (*models.LogicalCircuit, error) {
	if id <= 0 || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("valid logical circuit ID and name are required")
	}
	req.Type = defaultString(req.Type, "l2vpn")
	req.Status = defaultString(req.Status, "active")
	return s.repository.UpdateLogicalCircuit(ctx, id, req)
}

func (s *Service) DeleteLogicalCircuit(ctx context.Context, id int64) error {
	return s.repository.DeleteLogicalCircuit(ctx, id)
}

func (s *Service) ListCustomerAssociations(ctx context.Context, customerID int64) ([]*models.CustomerAssociation, error) {
	return s.repository.ListCustomerAssociations(ctx, customerID)
}

func (s *Service) CreateCustomerAssociation(ctx context.Context, req *repository.CustomerAssociationParams) (*models.CustomerAssociation, error) {
	if req.CustomerID <= 0 || req.ObjectID <= 0 || strings.TrimSpace(req.ObjectType) == "" {
		return nil, fmt.Errorf("customer, object type, and object ID are required")
	}
	req.Relationship = defaultString(req.Relationship, "owner")
	return s.repository.CreateCustomerAssociation(ctx, req)
}

func (s *Service) DeleteCustomerAssociation(ctx context.Context, id int64) error {
	return s.repository.DeleteCustomerAssociation(ctx, id)
}
