package services

import (
	"context"
	"fmt"
	"strings"

	"padduck/models"
	"padduck/repository"
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
	ns, err := s.repository.GetNameserverByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("nameserver %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return ns, nil
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
	ns, err := s.repository.UpdateNameserver(ctx, id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("nameserver %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return ns, nil
}

// DeleteNameserver deletes a nameserver by ID.
func (s *Service) DeleteNameserver(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid nameserver ID")
	}
	if err := s.repository.DeleteNameserver(ctx, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("nameserver %d: %w", id, ErrNotFound)
		}
		return err
	}
	return nil
}
