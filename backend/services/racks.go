package services

import (
	"context"
	"fmt"
	"strings"

	"padduck/models"
	"padduck/repository"
)

// RackCreateRequest holds input for creating a rack.
type RackCreateRequest = repository.RackParams

// RackUpdateRequest holds input for updating a rack.
type RackUpdateRequest = repository.RackParams

// CreateRack creates a new rack.
func (s *Service) CreateRack(ctx context.Context, req *RackCreateRequest) (*models.Rack, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("rack name is required")
	}
	if req.SizeU <= 0 {
		req.SizeU = 42
	}
	return s.repository.CreateRack(ctx, req)
}

// GetRack retrieves a rack by ID.
func (s *Service) GetRack(ctx context.Context, id int64) (*models.Rack, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid rack ID")
	}
	rack, err := s.repository.GetRackByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("rack %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return rack, nil
}

// ListRacks returns all racks, optionally filtered by location.
func (s *Service) ListRacks(ctx context.Context, locationID *int64) ([]*models.Rack, error) {
	return s.repository.ListRacks(ctx, locationID)
}

// UpdateRack updates an existing rack.
func (s *Service) UpdateRack(ctx context.Context, id int64, req *RackUpdateRequest) (*models.Rack, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid rack ID")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("rack name is required")
	}
	if req.SizeU <= 0 {
		req.SizeU = 42
	}
	rack, err := s.repository.UpdateRack(ctx, id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("rack %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return rack, nil
}

// DeleteRack deletes a rack by ID.
func (s *Service) DeleteRack(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid rack ID")
	}
	if err := s.repository.DeleteRack(ctx, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("rack %d: %w", id, ErrNotFound)
		}
		return err
	}
	return nil
}

// ListDevicesInRack returns all devices assigned to a rack, ordered by rack_unit_start.
func (s *Service) ListDevicesInRack(ctx context.Context, rackID int64) ([]*models.Device, error) {
	if rackID <= 0 {
		return nil, fmt.Errorf("invalid rack ID")
	}
	return s.repository.ListDevicesInRack(ctx, rackID)
}
