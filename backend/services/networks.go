package services

import (
	"context"
	"fmt"

	"padduck/models"
)

// CreateNetwork creates a new section
func (s *Service) CreateNetwork(ctx context.Context, name, description string, createdBy int64) (*models.Network, error) {
	if name == "" {
		return nil, fmt.Errorf("section name is required")
	}

	return s.repository.CreateNetwork(ctx, name, description, createdBy)
}

// GetNetwork retrieves a section by ID
func (s *Service) GetNetwork(ctx context.Context, id int64) (*models.Network, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}

	return s.repository.GetNetworkByID(ctx, id)
}

// ListNetworks returns all sections
func (s *Service) ListNetworks(ctx context.Context) ([]*models.Network, error) {
	return s.repository.ListAllNetworks(ctx)
}

// UpdateNetwork updates an existing section
func (s *Service) UpdateNetwork(ctx context.Context, id int64, name, description string) (*models.Network, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}
	if name == "" {
		return nil, fmt.Errorf("section name is required")
	}

	return s.repository.UpdateNetwork(ctx, id, name, description)
}

// DeleteNetwork deletes a section and its subnets (cascade)
func (s *Service) DeleteNetwork(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid section ID")
	}

	return s.repository.DeleteNetwork(ctx, id)
}
