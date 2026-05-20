package services

import (
	"context"
	"fmt"

	"padduck/models"
)

func (s *Service) CreateVRF(ctx context.Context, name, rd, description string) (*models.VRF, error) {
	if name == "" {
		return nil, fmt.Errorf("VRF name is required")
	}
	return s.repository.CreateVRF(ctx, name, rd, description)
}

func (s *Service) GetVRF(ctx context.Context, id int64) (*models.VRF, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VRF ID")
	}
	return s.repository.GetVRFByID(ctx, id)
}

func (s *Service) ListVRFs(ctx context.Context) ([]*models.VRF, error) {
	return s.repository.ListAllVRFs(ctx)
}

func (s *Service) UpdateVRF(ctx context.Context, id int64, name, rd, description string) (*models.VRF, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VRF ID")
	}
	if name == "" {
		return nil, fmt.Errorf("VRF name is required")
	}
	return s.repository.UpdateVRF(ctx, id, name, rd, description)
}

func (s *Service) DeleteVRF(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid VRF ID")
	}
	return s.repository.DeleteVRF(ctx, id)
}
