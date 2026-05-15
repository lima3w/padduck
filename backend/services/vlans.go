package services

import (
	"context"
	"fmt"

	"ipam-next/models"
)

func (s *Service) CreateVLAN(ctx context.Context, vrfID *int64, domainID *int64, vlanID int, name, description string) (*models.VLAN, error) {
	if vlanID < 1 || vlanID > 4094 {
		return nil, fmt.Errorf("VLAN ID must be between 1 and 4094")
	}
	if name == "" {
		return nil, fmt.Errorf("VLAN name is required")
	}
	return s.repository.CreateVLAN(ctx, vrfID, domainID, vlanID, name, description)
}

func (s *Service) GetVLAN(ctx context.Context, id int64) (*models.VLAN, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VLAN ID")
	}
	return s.repository.GetVLANByID(ctx, id)
}

func (s *Service) ListVLANs(ctx context.Context) ([]*models.VLAN, error) {
	return s.repository.ListAllVLANs(ctx)
}

func (s *Service) ListVLANsByVRF(ctx context.Context, vrfID int64) ([]*models.VLAN, error) {
	if vrfID <= 0 {
		return nil, fmt.Errorf("invalid VRF ID")
	}
	return s.repository.ListVLANsByVRF(ctx, vrfID)
}

func (s *Service) UpdateVLAN(ctx context.Context, id int64, domainID *int64, name, description string) (*models.VLAN, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VLAN ID")
	}
	if name == "" {
		return nil, fmt.Errorf("VLAN name is required")
	}
	return s.repository.UpdateVLAN(ctx, id, domainID, name, description)
}

func (s *Service) DeleteVLAN(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid VLAN ID")
	}
	return s.repository.DeleteVLAN(ctx, id)
}

// VLAN Domain methods

func (s *Service) CreateVLANDomain(ctx context.Context, name string, description *string) (*models.VLANDomain, error) {
	if name == "" {
		return nil, fmt.Errorf("VLAN domain name is required")
	}
	return s.repository.CreateVLANDomain(ctx, name, description)
}

func (s *Service) GetVLANDomain(ctx context.Context, id int64) (*models.VLANDomain, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VLAN domain ID")
	}
	return s.repository.GetVLANDomainByID(ctx, id)
}

func (s *Service) ListVLANDomains(ctx context.Context) ([]*models.VLANDomain, error) {
	return s.repository.ListVLANDomains(ctx)
}

func (s *Service) UpdateVLANDomain(ctx context.Context, id int64, name string, description *string) (*models.VLANDomain, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VLAN domain ID")
	}
	if name == "" {
		return nil, fmt.Errorf("VLAN domain name is required")
	}
	return s.repository.UpdateVLANDomain(ctx, id, name, description)
}

func (s *Service) DeleteVLANDomain(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid VLAN domain ID")
	}
	return s.repository.DeleteVLANDomain(ctx, id)
}
