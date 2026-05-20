package services

import (
	"context"
	"fmt"

	"padduck/models"
)

// CreateSection creates a new section
func (s *Service) CreateSection(ctx context.Context, name, description string, createdBy int64) (*models.Section, error) {
	if name == "" {
		return nil, fmt.Errorf("section name is required")
	}

	return s.repository.CreateSection(ctx, name, description, createdBy)
}

// GetSection retrieves a section by ID
func (s *Service) GetSection(ctx context.Context, id int64) (*models.Section, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}

	return s.repository.GetSectionByID(ctx, id)
}

// ListSections returns all sections
func (s *Service) ListSections(ctx context.Context) ([]*models.Section, error) {
	return s.repository.ListAllSections(ctx)
}

// UpdateSection updates an existing section
func (s *Service) UpdateSection(ctx context.Context, id int64, name, description string) (*models.Section, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}
	if name == "" {
		return nil, fmt.Errorf("section name is required")
	}

	return s.repository.UpdateSection(ctx, id, name, description)
}

// DeleteSection deletes a section and its subnets (cascade)
func (s *Service) DeleteSection(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid section ID")
	}

	return s.repository.DeleteSection(ctx, id)
}
