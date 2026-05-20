package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"padduck/models"
	"padduck/repository"
)

// GlobalSearchResult holds results from a cross-entity search.
type GlobalSearchResult struct {
	Sections []*models.Section `json:"sections"`
	Subnets  []*models.Subnet  `json:"subnets"`
	Devices  []*models.Device  `json:"devices"`
}

// GlobalSearch searches sections, subnets, and devices concurrently.
func (s *Service) GlobalSearch(ctx context.Context, query string, limit int64) (*GlobalSearchResult, error) {
	empty := &GlobalSearchResult{
		Sections: make([]*models.Section, 0),
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
		secs, err := s.repository.SearchSections(ctx, query, limit, 0)
		if err == nil && len(secs) > 0 {
			mu.Lock()
			result.Sections = secs
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		subs, err := s.repository.GlobalSearchSubnets(ctx, query, limit)
		if err == nil && len(subs) > 0 {
			mu.Lock()
			result.Subnets = subs
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		devs, err := s.repository.SearchDevicesWithCustomFields(ctx, &repository.DeviceSearchFilter{Query: query}, nil)
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
func (s *Service) SearchSubnets(ctx context.Context, sectionID int64, query string, limit, offset int64, cfFilters ...map[string]string) ([]*models.Subnet, error) {
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

	var cf map[string]string
	if len(cfFilters) > 0 {
		cf = cfFilters[0]
	}
	return s.repository.SearchSubnetsWithCustomFields(ctx, sectionID, query, limit, offset, cf)
}

// IPSearchOptions holds additional search filters for IP address search
type IPSearchOptions struct {
	TagID          *int64
	MACAddress     string
	PTRRecord      string
	IsAssigned     *bool
	LastSeenAfter  *time.Time
	LastSeenBefore *time.Time
	CustomFields   map[string]string
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

	return s.repository.SearchIPAddressesWithCustomFields(ctx, subnetID, query, status, limit, offset, repoFilter, cfFilters)
}
