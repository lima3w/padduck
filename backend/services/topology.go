package services

import (
	"context"
	"fmt"

	"padduck/models"
)

type topologyRepo interface {
	ListTopologyHints(ctx context.Context, status string) ([]*models.TopologyHint, error)
	GetTopologyHint(ctx context.Context, id int64) (*models.TopologyHint, error)
	UpsertTopologyHint(ctx context.Context, sourceType string, sourceID int64, targetType string, targetID int64, hintType string, confidence float64, evidence *string) (*models.TopologyHint, error)
	UpdateTopologyHintStatus(ctx context.Context, id int64, status string) (*models.TopologyHint, error)

	UpsertTopologyNode(ctx context.Context, orgID *int64, resourceType string, resourceID int64) (*models.GraphNode, error)
	UpsertTopologyEdge(ctx context.Context, orgID *int64, srcType string, srcID int64, tgtType string, tgtID int64, relationship string, confidence float64, status string, evidence *string) (*models.GraphEdge, error)
	GetTopologyNeighbors(ctx context.Context, orgID *int64, rootType string, rootID int64, maxDepth int) ([]*models.GraphNode, []*models.GraphEdge, error)
	GetTopologyPath(ctx context.Context, fromType string, fromID int64, toType string, toID int64) ([]*models.GraphNode, []*models.GraphEdge, error)
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

// AddNode idempotently registers a resource as a node in the topology graph.
func (s *TopologyService) AddNode(ctx context.Context, orgID *int64, resourceType string, resourceID int64) (*models.GraphNode, error) {
	return s.repository.UpsertTopologyNode(ctx, orgID, resourceType, resourceID)
}

// AddEdge creates or updates a directed edge between two resource nodes.
// Both nodes are auto-upserted so callers don't need to pre-register them.
func (s *TopologyService) AddEdge(ctx context.Context, orgID *int64, srcType string, srcID int64, tgtType string, tgtID int64, relationship string) (*models.GraphEdge, error) {
	if _, err := s.repository.UpsertTopologyNode(ctx, orgID, srcType, srcID); err != nil {
		return nil, err
	}
	if _, err := s.repository.UpsertTopologyNode(ctx, orgID, tgtType, tgtID); err != nil {
		return nil, err
	}
	return s.repository.UpsertTopologyEdge(ctx, orgID, srcType, srcID, tgtType, tgtID, relationship, 1.0, "confirmed", nil)
}

// GetNeighbors returns a Cytoscape-compatible graph of all nodes reachable from the
// root resource within maxDepth hops.
func (s *TopologyService) GetNeighbors(ctx context.Context, orgID *int64, rootType string, rootID int64, maxDepth int) (*models.CytoscapeGraph, error) {
	nodes, edges, err := s.repository.GetTopologyNeighbors(ctx, orgID, rootType, rootID, maxDepth)
	if err != nil {
		return nil, err
	}
	return toCytoscapeGraph(nodes, edges), nil
}

// GetPath returns a Cytoscape-compatible graph of the shortest path between two resources.
func (s *TopologyService) GetPath(ctx context.Context, fromType string, fromID int64, toType string, toID int64) (*models.CytoscapeGraph, error) {
	nodes, edges, err := s.repository.GetTopologyPath(ctx, fromType, fromID, toType, toID)
	if err != nil {
		return nil, err
	}
	return toCytoscapeGraph(nodes, edges), nil
}

// toCytoscapeGraph converts raw graph nodes and edges into the Cytoscape.js element format.
func toCytoscapeGraph(nodes []*models.GraphNode, edges []*models.GraphEdge) *models.CytoscapeGraph {
	g := &models.CytoscapeGraph{
		Nodes: make([]models.CytoscapeNode, 0, len(nodes)),
		Edges: make([]models.CytoscapeEdge, 0, len(edges)),
	}
	for _, n := range nodes {
		g.Nodes = append(g.Nodes, models.CytoscapeNode{Data: map[string]any{
			"id":            fmt.Sprintf("%s_%d", n.ResourceType, n.ResourceID),
			"resource_type": n.ResourceType,
			"resource_id":   n.ResourceID,
		}})
	}
	for _, e := range edges {
		g.Edges = append(g.Edges, models.CytoscapeEdge{Data: map[string]any{
			"id":           fmt.Sprintf("edge_%d", e.ID),
			"source":       nodeKey(nodes, e.SourceNodeID),
			"target":       nodeKey(nodes, e.TargetNodeID),
			"relationship": e.Relationship,
			"status":       e.Status,
		}})
	}
	return g
}

// nodeKey returns "resourceType_resourceID" for a given node DB id, falling back to the raw id.
func nodeKey(nodes []*models.GraphNode, nodeID int64) string {
	for _, n := range nodes {
		if n.ID == nodeID {
			return fmt.Sprintf("%s_%d", n.ResourceType, n.ResourceID)
		}
	}
	return fmt.Sprintf("unknown_%d", nodeID)
}
