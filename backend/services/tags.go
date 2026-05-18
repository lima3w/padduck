package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ipam-next/models"
)

var (
	ErrSystemTag = errors.New("cannot modify system tag")
	ErrTagInUse  = errors.New("tag is in use")
)

// ListIPTags returns all IP tags
func (s *Service) ListIPTags(ctx context.Context) ([]*models.IPTag, error) {
	return s.repository.ListIPTags(ctx)
}

// GetIPTag retrieves an IP tag by ID
func (s *Service) GetIPTag(ctx context.Context, id int64) (*models.IPTag, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid tag ID")
	}
	tag, err := s.repository.GetIPTagByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("tag %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return tag, nil
}

// CreateIPTag creates a new custom IP tag (admin only)
func (s *Service) CreateIPTag(ctx context.Context, name, colour string, description *string) (*models.IPTag, error) {
	if name == "" {
		return nil, fmt.Errorf("tag name is required")
	}
	if colour == "" {
		colour = "#6B7280"
	}
	return s.repository.CreateIPTag(ctx, name, colour, description)
}

// UpdateIPTag updates an existing IP tag
func (s *Service) UpdateIPTag(ctx context.Context, id int64, name, colour string, description *string) (*models.IPTag, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid tag ID")
	}
	if name == "" {
		return nil, fmt.Errorf("tag name is required")
	}
	if colour == "" {
		colour = "#6B7280"
	}
	return s.repository.UpdateIPTag(ctx, id, name, colour, description)
}

// DeleteIPTag deletes a non-system tag that is not in use
func (s *Service) DeleteIPTag(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid tag ID")
	}
	if err := s.repository.DeleteIPTag(ctx, id); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			return fmt.Errorf("tag %d: %w", id, ErrNotFound)
		}
		if strings.Contains(msg, "system tag") {
			return fmt.Errorf("tag %d: %w", id, ErrSystemTag)
		}
		if strings.Contains(msg, "in use") {
			return fmt.Errorf("tag %d: %w", id, ErrTagInUse)
		}
		return err
	}
	return nil
}
