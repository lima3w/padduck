package services

import (
	"context"
	"fmt"
	"net"
	"time"

	"ipam-next/models"
)

// CreateIPAddress creates a new IP address record
func (s *Service) CreateIPAddress(ctx context.Context, subnetID int64, address, hostname string, status string, tagID *int64, macAddress, ptrRecord *string) (*models.IPAddress, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	if net.ParseIP(address) == nil {
		return nil, fmt.Errorf("invalid IP address: %s", address)
	}

	validStatuses := map[string]bool{"available": true, "assigned": true, "reserved": true}
	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid IP status: %s", status)
	}

	return s.repository.CreateIPAddress(ctx, subnetID, address, hostname, status, nil, tagID, macAddress, ptrRecord)
}

// UpdateIPAddressMeta updates tag, mac, and ptr_record fields of an IP address
func (s *Service) UpdateIPAddressMeta(ctx context.Context, id int64, tagID *int64, macAddress, ptrRecord *string) (*models.IPAddress, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid IP address ID")
	}
	return s.repository.UpdateIPAddressFull(ctx, id, tagID, macAddress, ptrRecord)
}

// GetIPAddress retrieves an IP address by ID
func (s *Service) GetIPAddress(ctx context.Context, id int64) (*models.IPAddress, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid IP address ID")
	}

	return s.repository.GetIPAddressByID(ctx, id)
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

	return s.repository.UpdateIPAddressStatus(ctx, id, "available", nil)
}

// DeleteIPAddress deletes an IP address record
func (s *Service) DeleteIPAddress(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid IP address ID")
	}

	return s.repository.DeleteIPAddress(ctx, id)
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
	Total      int64   `json:"total"`
	Available  int64   `json:"available"`
	Assigned   int64   `json:"assigned"`
	Reserved   int64   `json:"reserved"`
	Utilization float64 `json:"utilization_percent"`
}

// GetSubnetUtilization calculates utilization statistics for a subnet
func (s *Service) GetSubnetUtilization(ctx context.Context, subnetID int64) (*SubnetUtilization, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	total, err := s.repository.CountTotalIPsBySubnet(ctx, subnetID)
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

	available, err := s.repository.CountIPsByStatus(ctx, subnetID, "available")
	if err != nil {
		return nil, err
	}

	assigned, err := s.repository.CountIPsByStatus(ctx, subnetID, "assigned")
	if err != nil {
		return nil, err
	}

	reserved, err := s.repository.CountIPsByStatus(ctx, subnetID, "reserved")
	if err != nil {
		return nil, err
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

	now := time.Now()
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
