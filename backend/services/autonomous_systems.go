package services

import (
	"context"
	"fmt"

	"padduck/models"
)

func (s *Service) CreateAutonomousSystem(ctx context.Context, asn int64, name, description, asType, rir string) (*models.AutonomousSystem, error) {
	if asn <= 0 {
		return nil, fmt.Errorf("ASN must be a positive integer")
	}
	if asType != "internal" && asType != "external" {
		asType = "external"
	}
	return s.repository.CreateAutonomousSystem(ctx, asn, name, description, asType, rir)
}

func (s *Service) GetAutonomousSystem(ctx context.Context, id int64) (*models.AutonomousSystem, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid autonomous system ID")
	}
	return s.repository.GetAutonomousSystemByID(ctx, id)
}

func (s *Service) ListAutonomousSystems(ctx context.Context) ([]*models.AutonomousSystem, error) {
	return s.repository.ListAllAutonomousSystems(ctx)
}

func (s *Service) UpdateAutonomousSystem(ctx context.Context, id, asn int64, name, description, asType, rir string) (*models.AutonomousSystem, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid autonomous system ID")
	}
	if asn <= 0 {
		return nil, fmt.Errorf("ASN must be a positive integer")
	}
	if asType != "internal" && asType != "external" {
		asType = "external"
	}
	return s.repository.UpdateAutonomousSystem(ctx, id, asn, name, description, asType, rir)
}

func (s *Service) DeleteAutonomousSystem(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid autonomous system ID")
	}
	return s.repository.DeleteAutonomousSystem(ctx, id)
}
