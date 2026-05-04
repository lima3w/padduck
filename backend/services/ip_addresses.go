package services

import (
	"context"
	"fmt"
	"net"

	"ipam-next/models"
)

// CreateIPAddress creates a new IP address record
func (s *Service) CreateIPAddress(ctx context.Context, subnetID int64, address, hostname string, status string) (*models.IPAddress, error) {
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

	return s.repository.CreateIPAddress(ctx, subnetID, address, hostname, status, nil)
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
