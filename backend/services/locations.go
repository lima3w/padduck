package services

import (
	"context"
	"fmt"
	"strings"

	"ipam-next/models"
	"ipam-next/repository"
)

// LocationCreateRequest holds input for creating a location.
type LocationCreateRequest = repository.LocationParams

// LocationUpdateRequest holds input for updating a location.
type LocationUpdateRequest = repository.LocationParams

// CreateLocation creates a new location.
func (s *Service) CreateLocation(ctx context.Context, req *LocationCreateRequest) (*models.Location, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("location name is required")
	}
	if req.Type == "" {
		req.Type = "other"
	}
	if req.Status == "" {
		req.Status = "active"
	}
	return s.repository.CreateLocation(ctx, req)
}

// GetLocation retrieves a location by ID.
func (s *Service) GetLocation(ctx context.Context, id int64) (*models.Location, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid location ID")
	}
	loc, err := s.repository.GetLocationByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("location %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return loc, nil
}

// ListLocations returns all locations.
func (s *Service) ListLocations(ctx context.Context) ([]*models.Location, error) {
	return s.repository.ListLocations(ctx)
}

// UpdateLocation updates an existing location.
func (s *Service) UpdateLocation(ctx context.Context, id int64, req *LocationUpdateRequest) (*models.Location, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid location ID")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("location name is required")
	}
	if req.Type == "" {
		req.Type = "other"
	}
	if req.Status == "" {
		req.Status = "active"
	}
	loc, err := s.repository.UpdateLocation(ctx, id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("location %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return loc, nil
}

// DeleteLocation deletes a location by ID.
func (s *Service) DeleteLocation(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid location ID")
	}
	if err := s.repository.DeleteLocation(ctx, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("location %d: %w", id, ErrNotFound)
		}
		return err
	}
	return nil
}

// GetLocationTree returns all locations assembled into a nested tree.
func (s *Service) GetLocationTree(ctx context.Context) ([]*models.LocationTreeNode, error) {
	locs, err := s.repository.GetLocationTree(ctx)
	if err != nil {
		return nil, err
	}
	return buildLocationTree(locs), nil
}

// buildLocationTree assembles a flat location list (breadth-first) into a nested tree.
func buildLocationTree(locs []*models.Location) []*models.LocationTreeNode {
	nodeMap := make(map[int64]*models.LocationTreeNode, len(locs))
	for _, l := range locs {
		nodeMap[l.ID] = &models.LocationTreeNode{Location: *l, Children: []*models.LocationTreeNode{}}
	}

	roots := make([]*models.LocationTreeNode, 0)
	for _, l := range locs {
		node := nodeMap[l.ID]
		if l.ParentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := nodeMap[*l.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}
	return roots
}
