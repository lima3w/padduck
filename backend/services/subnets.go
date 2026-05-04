package services

import (
	"context"
	"fmt"
	"net"

	"ipam-next/models"
)

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

// CreateSubnet creates a new subnet with CIDR validation
func (s *Service) CreateSubnet(ctx context.Context, sectionID int64, networkAddress string, prefixLength int, description string) (*models.Subnet, error) {
	if sectionID <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}

	if err := ValidateCIDR(networkAddress, prefixLength); err != nil {
		return nil, err
	}

	return s.repository.CreateSubnet(ctx, sectionID, networkAddress, prefixLength, description)
}

// GetSubnet retrieves a subnet by ID
func (s *Service) GetSubnet(ctx context.Context, id int64) (*models.Subnet, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	return s.repository.GetSubnetByID(ctx, id)
}

// ListSubnets returns all subnets in a section
func (s *Service) ListSubnets(ctx context.Context, sectionID int64) ([]*models.Subnet, error) {
	if sectionID <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}

	return s.repository.ListSubnetsBySection(ctx, sectionID)
}

// UpdateSubnet updates a subnet's description
func (s *Service) UpdateSubnet(ctx context.Context, id int64, description string) (*models.Subnet, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}

	return s.repository.UpdateSubnet(ctx, id, description)
}

// DeleteSubnet deletes a subnet and its IP addresses (cascade)
func (s *Service) DeleteSubnet(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid subnet ID")
	}

	return s.repository.DeleteSubnet(ctx, id)
}
