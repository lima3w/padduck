package services

import (
	"context"
	"fmt"
	"time"

	"ipam-next/models"
)

// ListDelegations returns all IPv6 delegations for a given subnet.
func (s *Service) ListDelegations(ctx context.Context, subnetID int64) ([]*models.IPv6Delegation, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}
	delegations, err := s.repository.ListIPv6DelegationsBySubnet(ctx, subnetID)
	if err != nil {
		return nil, err
	}
	for _, d := range delegations {
		d.IsExpired = isExpiredNow(d.ExpiresAt)
	}
	return delegations, nil
}

// CreateDelegation creates a new IPv6 delegation for a subnet.
func (s *Service) CreateDelegation(ctx context.Context, d *models.IPv6Delegation) (*models.IPv6Delegation, error) {
	if d.ParentSubnetID <= 0 {
		return nil, fmt.Errorf("invalid parent subnet ID")
	}
	if d.DelegatedPrefix == "" {
		return nil, fmt.Errorf("delegated prefix is required")
	}
	result, err := s.repository.CreateIPv6Delegation(ctx, d)
	if err != nil {
		return nil, err
	}
	result.IsExpired = isExpiredNow(result.ExpiresAt)
	return result, nil
}

// UpdateDelegation updates an existing IPv6 delegation.
func (s *Service) UpdateDelegation(ctx context.Context, id int64, d *models.IPv6Delegation) (*models.IPv6Delegation, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid delegation ID")
	}
	if d.DelegatedPrefix == "" {
		return nil, fmt.Errorf("delegated prefix is required")
	}
	result, err := s.repository.UpdateIPv6Delegation(ctx, id, d)
	if err != nil {
		return nil, err
	}
	result.IsExpired = isExpiredNow(result.ExpiresAt)
	return result, nil
}

// DeleteDelegation deletes an IPv6 delegation by ID.
func (s *Service) DeleteDelegation(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid delegation ID")
	}
	return s.repository.DeleteIPv6Delegation(ctx, id)
}

// isExpiredNow returns true if expiresAt is set and is before the current time.
func isExpiredNow(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return false
	}
	return time.Now().After(*expiresAt)
}
