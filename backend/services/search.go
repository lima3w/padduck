package services

import (
	"context"
	"fmt"
	"time"

	"ipam-next/models"
	"ipam-next/repository"
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

// IPSearchOptions holds additional search filters for IP address search
type IPSearchOptions struct {
	TagID          *int64
	MACAddress     string
	PTRRecord      string
	IsAssigned     *bool
	LastSeenAfter  *time.Time
	LastSeenBefore *time.Time
}

// SearchIPAddresses searches for IP addresses in a subnet
func (s *Service) SearchIPAddresses(ctx context.Context, subnetID int64, query string, status string, limit, offset int64, opts ...IPSearchOptions) ([]*models.IPAddress, error) {
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
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

	var repoFilter repository.IPSearchFilter
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
	}

	return s.repository.SearchIPAddresses(ctx, subnetID, query, status, limit, offset, repoFilter)
}
