package services

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"padduck/models"
)

// CreateIPAddress creates a new IP address record
func (s *Service) CreateIPAddress(ctx context.Context, subnetID int64, address, hostname string, status string, tagID *int64, macAddress, ptrRecord, dnsName *string, customFields ...map[string]*string) (*models.IPAddress, error) {
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

	subnet, err := s.repository.GetSubnetByID(ctx, subnetID)
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

	ip, err := s.repository.CreateIPAddress(ctx, subnetID, address, hostname, status, nil, tagID, macAddress, ptrRecord, dnsName)
	if err != nil {
		return nil, normalizeCreateIPAddressError(err, address)
	}

	if len(customFields) > 0 && customFields[0] != nil {
		_ = s.SetCustomFieldValues(ctx, "ip_address", ip.ID, customFields[0])
		ip.CustomFields, _ = s.repository.GetCustomFieldValues(ctx, "ip_address", ip.ID)
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

// UpdateIPAddressMeta updates hostname, tag, mac, ptr_record, and dns_name fields of an IP address
func (s *Service) UpdateIPAddressMeta(ctx context.Context, id int64, hostname string, tagID *int64, macAddress, ptrRecord, dnsName *string, customFields ...map[string]*string) (*models.IPAddress, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid IP address ID")
	}
	ip, err := s.repository.UpdateIPAddressFull(ctx, id, hostname, tagID, macAddress, ptrRecord, dnsName)
	if err != nil {
		return nil, err
	}

	if len(customFields) > 0 && customFields[0] != nil {
		_ = s.SetCustomFieldValues(ctx, "ip_address", ip.ID, customFields[0])
	}
	ip.CustomFields, _ = s.repository.GetCustomFieldValues(ctx, "ip_address", ip.ID)
	return ip, nil
}

// GetIPAddress retrieves an IP address by ID
func (s *Service) GetIPAddress(ctx context.Context, id int64) (*models.IPAddress, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid IP address ID")
	}

	ip, err := s.repository.GetIPAddressByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ip.CustomFields, _ = s.repository.GetCustomFieldValues(ctx, "ip_address", id)
	return ip, nil
}

// ListIPAddresses returns all IP addresses in a subnet
func (s *Service) ListIPAddresses(ctx context.Context, subnetID int64) ([]*models.IPAddress, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	return s.repository.ListIPAddressesBySubnet(ctx, subnetID)
}

// AssignIPAddress marks an IP as assigned to a user/device
func (s *Service) AssignIPAddress(ctx context.Context, id int64, assignedTo string) (*models.IPAddress, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid IP address ID")
	}

	if assignedTo == "" {
		return nil, fmt.Errorf("assigned_to cannot be empty")
	}

	return s.repository.UpdateIPAddressStatus(ctx, id, "assigned", &assignedTo)
}

// ReleaseIPAddress marks an IP as available again
func (s *Service) ReleaseIPAddress(ctx context.Context, id int64) (*models.IPAddress, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid IP address ID")
	}
	ip, err := s.repository.GetIPAddressByID(ctx, id)
	if err != nil {
		return nil, err
	}
	result, err := s.repository.UpdateIPAddressStatus(ctx, id, "available", nil)
	if err == nil {
		go s.DNS.RemoveIPFromDNS(ctx, ip)
	}
	return result, err
}

// DeleteIPAddress deletes an IP address record
func (s *Service) DeleteIPAddress(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid IP address ID")
	}
	ip, err := s.repository.GetIPAddressByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repository.DeleteIPAddress(ctx, id); err != nil {
		return err
	}
	go s.DNS.RemoveIPFromDNS(ctx, ip)
	return nil
}

// FindNextAvailableIP returns the next available IP in a subnet
func (s *Service) FindNextAvailableIP(ctx context.Context, subnetID int64) (*models.IPAddress, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	availableIPs, err := s.repository.ListAvailableIPsBySubnet(ctx, subnetID)
	if err != nil {
		return nil, err
	}

	if len(availableIPs) == 0 {
		return nil, fmt.Errorf("no available IP addresses in subnet")
	}

	return availableIPs[0], nil
}

// AllocateIPAddress atomically finds and assigns the next available IP
func (s *Service) AllocateIPAddress(ctx context.Context, subnetID int64, assignedTo string) (*models.IPAddress, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	if assignedTo == "" {
		return nil, fmt.Errorf("assigned_to cannot be empty")
	}

	return s.repository.AllocateIPAddress(ctx, subnetID, assignedTo)
}

// SubnetUtilization represents utilization statistics for a subnet
type SubnetUtilization struct {
	Total       int64   `json:"total"`
	Available   int64   `json:"available"`
	Assigned    int64   `json:"assigned"`
	Reserved    int64   `json:"reserved"`
	Utilization float64 `json:"utilization_percent"`
}

// GetSubnetUtilization calculates utilization statistics for a subnet
func (s *Service) GetSubnetUtilization(ctx context.Context, subnetID int64) (*SubnetUtilization, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	total, available, assigned, reserved, err := s.repository.GetSubnetUtilizationCounts(ctx, subnetID)
	if err != nil {
		return nil, err
	}

	if total == 0 {
		return &SubnetUtilization{
			Total:       0,
			Available:   0,
			Assigned:    0,
			Reserved:    0,
			Utilization: 0,
		}, nil
	}

	utilization := (float64(assigned+reserved) / float64(total)) * 100

	return &SubnetUtilization{
		Total:       total,
		Available:   available,
		Assigned:    assigned,
		Reserved:    reserved,
		Utilization: utilization,
	}, nil
}

// AssignIPAddressWithLease marks an IP as assigned with optional lease expiration
func (s *Service) AssignIPAddressWithLease(ctx context.Context, id int64, assignedTo string, leaseDurationDays int) (*models.IPAddress, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid IP address ID")
	}

	if assignedTo == "" {
		return nil, fmt.Errorf("assigned_to cannot be empty")
	}

	now := time.Now().UTC()
	assignedAtTime := now
	var expiresAtTime *time.Time

	if leaseDurationDays > 0 {
		expiresAt := now.AddDate(0, 0, leaseDurationDays)
		expiresAtTime = &expiresAt
	}

	return s.repository.UpdateIPAddressWithLease(ctx, id, "assigned", &assignedTo, &assignedAtTime, expiresAtTime)
}

// IsIPLeaseExpired checks if an IP's lease has expired
func (s *Service) IsIPLeaseExpired(ctx context.Context, id int64) (bool, error) {
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

// ReleaseExpiredLease releases an IP if its lease has expired
func (s *Service) ReleaseExpiredLease(ctx context.Context, id int64) (*models.IPAddress, error) {
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
