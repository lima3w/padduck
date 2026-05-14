package services

import (
	"context"
	"fmt"

	"ipam-next/models"
	"ipam-next/repository"
)

// NameserverCreateRequest holds input for creating a nameserver.
type NameserverCreateRequest = repository.NameserverParams

// NameserverUpdateRequest holds input for updating a nameserver.
type NameserverUpdateRequest = repository.NameserverParams

// CreateNameserver creates a new nameserver entry.
func (s *Service) CreateNameserver(ctx context.Context, req *NameserverCreateRequest) (*models.Nameserver, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("nameserver name is required")
	}
	if req.Server1 == "" {
		return nil, fmt.Errorf("server1 is required")
	}
	return s.repository.CreateNameserver(ctx, req)
}

// GetNameserver retrieves a nameserver by ID.
func (s *Service) GetNameserver(ctx context.Context, id int64) (*models.Nameserver, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid nameserver ID")
	}
	return s.repository.GetNameserverByID(ctx, id)
}

// ListNameservers returns all nameservers.
func (s *Service) ListNameservers(ctx context.Context) ([]*models.Nameserver, error) {
	ns, err := s.repository.ListNameservers(ctx)
	if err != nil {
		return nil, err
	}
	if ns == nil {
		ns = []*models.Nameserver{}
	}
	return ns, nil
}

// UpdateNameserver updates an existing nameserver.
func (s *Service) UpdateNameserver(ctx context.Context, id int64, req *NameserverUpdateRequest) (*models.Nameserver, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid nameserver ID")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("nameserver name is required")
	}
	if req.Server1 == "" {
		return nil, fmt.Errorf("server1 is required")
	}
	return s.repository.UpdateNameserver(ctx, id, req)
}

// DeleteNameserver deletes a nameserver by ID.
func (s *Service) DeleteNameserver(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid nameserver ID")
	}
	return s.repository.DeleteNameserver(ctx, id)
}
