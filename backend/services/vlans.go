package services

import (
	"context"
	"fmt"
	"time"

	"ipam-next/models"
)

func (s *Service) CreateVLAN(ctx context.Context, vrfID *int64, domainID *int64, groupID *int64, vlanID int, name, description string) (*models.VLAN, error) {
	if vlanID < 1 || vlanID > 4094 {
		return nil, fmt.Errorf("VLAN ID must be between 1 and 4094")
	}
	if name == "" {
		return nil, fmt.Errorf("VLAN name is required")
	}
	return s.repository.CreateVLAN(ctx, vrfID, domainID, groupID, vlanID, name, description)
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

func (s *Service) UpdateVLAN(ctx context.Context, id int64, domainID *int64, groupID *int64, name, description string) (*models.VLAN, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VLAN ID")
	}
	if name == "" {
		return nil, fmt.Errorf("VLAN name is required")
	}
	return s.repository.UpdateVLAN(ctx, id, domainID, groupID, name, description)
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

// VLAN Group methods

func (s *Service) CreateVLANGroup(ctx context.Context, name string, description *string, colour *string) (*models.VLANGroup, error) {
	if name == "" {
		return nil, fmt.Errorf("VLAN group name is required")
	}
	return s.repository.CreateVLANGroup(ctx, name, description, colour)
}

func (s *Service) GetVLANGroup(ctx context.Context, id int64) (*models.VLANGroup, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VLAN group ID")
	}
	return s.repository.GetVLANGroupByID(ctx, id)
}

func (s *Service) ListVLANGroups(ctx context.Context) ([]*models.VLANGroup, error) {
	return s.repository.ListVLANGroups(ctx)
}

func (s *Service) UpdateVLANGroup(ctx context.Context, id int64, name string, description *string, colour *string) (*models.VLANGroup, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid VLAN group ID")
	}
	if name == "" {
		return nil, fmt.Errorf("VLAN group name is required")
	}
	return s.repository.UpdateVLANGroup(ctx, id, name, description, colour)
}

func (s *Service) DeleteVLANGroup(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid VLAN group ID")
	}
	return s.repository.DeleteVLANGroup(ctx, id)
}

// GetVLANUsageReport returns per-VLAN metrics: subnet count, IP count, total IPs, utilisation.
func (s *Service) GetVLANUsageReport(ctx context.Context) (*models.VLANUsageReport, error) {
	entries, err := s.repository.GetVLANUsageReport(ctx)
	if err != nil {
		return nil, err
	}
	return &models.VLANUsageReport{
		Entries:     entries,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
