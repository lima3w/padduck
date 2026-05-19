package services

import (
	"context"
	"fmt"

	"ipam-next/models"
)

type topologyRepo interface {
	ListTopologyHints(ctx context.Context, status string) ([]*models.TopologyHint, error)
	GetTopologyHint(ctx context.Context, id int64) (*models.TopologyHint, error)
	UpsertTopologyHint(ctx context.Context, sourceType string, sourceID int64, targetType string, targetID int64, hintType string, confidence float64, evidence *string) (*models.TopologyHint, error)
	UpdateTopologyHintStatus(ctx context.Context, id int64, status string) (*models.TopologyHint, error)
}

// TopologyService handles business logic for topology hints.
type TopologyService struct {
	repository topologyRepo
}

// NewTopologyService creates a new TopologyService.
func NewTopologyService(repo topologyRepo) *TopologyService {
	return &TopologyService{repository: repo}
}

// ListHints returns all topology hints, optionally filtered by status.
func (s *TopologyService) ListHints(ctx context.Context, status string) ([]*models.TopologyHint, error) {
	return s.repository.ListTopologyHints(ctx, status)
}

// GetHint returns a single topology hint by ID.
func (s *TopologyService) GetHint(ctx context.Context, id int64) (*models.TopologyHint, error) {
	return s.repository.GetTopologyHint(ctx, id)
}

// UpsertHint creates or updates a topology hint.
func (s *TopologyService) UpsertHint(ctx context.Context, sourceType string, sourceID int64, targetType string, targetID int64, hintType string, confidence float64, evidence *string) (*models.TopologyHint, error) {
	return s.repository.UpsertTopologyHint(ctx, sourceType, sourceID, targetType, targetID, hintType, confidence, evidence)
}

// UpdateHintStatus updates the status of a topology hint. Status must be one of
// "suggested", "confirmed", or "dismissed".
func (s *TopologyService) UpdateHintStatus(ctx context.Context, id int64, status string) (*models.TopologyHint, error) {
	switch status {
	case "suggested", "confirmed", "dismissed":
		// valid
	default:
		return nil, fmt.Errorf("invalid status %q: must be suggested, confirmed, or dismissed", status)
	}
	return s.repository.UpdateTopologyHintStatus(ctx, id, status)
}
