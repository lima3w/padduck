package services

import (
	"context"
	"fmt"

	"ipam-next/models"
)

const (
	DefaultLimit  = 50
	MaxLimit      = 500
	DefaultOffset = 0
)

// SearchSections searches for sections by name or description
func (s *Service) SearchSections(ctx context.Context, query string, limit, offset int64) ([]*models.Section, error) {
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

	return s.repository.SearchSections(ctx, query, limit, offset)
}

// SearchSubnets searches for subnets in a section by network address or description
func (s *Service) SearchSubnets(ctx context.Context, sectionID int64, query string, limit, offset int64) ([]*models.Subnet, error) {
	if sectionID <= 0 {
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

	return s.repository.SearchSubnets(ctx, sectionID, query, limit, offset)
}

// SearchIPAddresses searches for IP addresses in a subnet
func (s *Service) SearchIPAddresses(ctx context.Context, subnetID int64, query string, status string, limit, offset int64) ([]*models.IPAddress, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Validate status if provided
	validStatuses := map[string]bool{
		"available": true,
		"assigned":  true,
		"reserved":  true,
	}
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

	return s.repository.SearchIPAddresses(ctx, subnetID, query, status, limit, offset)
}
