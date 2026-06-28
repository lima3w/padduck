package repository

import (
	"context"
	"fmt"

	"padduck/models"
)

// UpsertTopologyNode inserts or returns the existing node for (resource_type, resource_id).
func (r *Repository) UpsertTopologyNode(ctx context.Context, orgID *int64, resourceType string, resourceID int64) (*models.GraphNode, error) {
	n := &models.GraphNode{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO topology_nodes (organization_id, resource_type, resource_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (resource_type, resource_id) DO UPDATE
		  SET resource_type = EXCLUDED.resource_type
		RETURNING id, organization_id, resource_type, resource_id, created_at`,
		orgID, resourceType, resourceID,
	).Scan(&n.ID, &n.OrganizationID, &n.ResourceType, &n.ResourceID, &n.CreatedAt)
	return n, err
}

// getTopologyNodeID looks up the node ID for a (resource_type, resource_id) pair.
func (r *Repository) getTopologyNodeID(ctx context.Context, resourceType string, resourceID int64) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`SELECT id FROM topology_nodes WHERE resource_type = $1 AND resource_id = $2`,
		resourceType, resourceID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("topology node not found for %s/%d: %w", resourceType, resourceID, err)
	}
	return id, nil
}

// UpsertTopologyEdge creates or updates the edge between two nodes identified by resource type/ID pairs.
// Both nodes must already exist (call UpsertTopologyNode first).
func (r *Repository) UpsertTopologyEdge(ctx context.Context, orgID *int64, srcType string, srcID int64, tgtType string, tgtID int64, relationship string, confidence float64, status string, evidence *string) (*models.GraphEdge, error) {
	sourceNodeID, err := r.getTopologyNodeID(ctx, srcType, srcID)
	if err != nil {
		return nil, err
	}
	targetNodeID, err := r.getTopologyNodeID(ctx, tgtType, tgtID)
	if err != nil {
		return nil, err
	}

	e := &models.GraphEdge{}
	err = r.db.QueryRow(ctx, `
		INSERT INTO topology_edges
		  (organization_id, source_node_id, target_node_id, relationship, confidence, status, evidence)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (source_node_id, target_node_id, relationship) DO UPDATE
		  SET confidence = EXCLUDED.confidence,
		      status     = EXCLUDED.status,
		      evidence   = EXCLUDED.evidence
		RETURNING id, organization_id, source_node_id, target_node_id, relationship, confidence, status, evidence, created_at`,
		orgID, sourceNodeID, targetNodeID, relationship, confidence, status, evidence,
	).Scan(&e.ID, &e.OrganizationID, &e.SourceNodeID, &e.TargetNodeID,
		&e.Relationship, &e.Confidence, &e.Status, &e.Evidence, &e.CreatedAt)
	return e, err
}

// GetTopologyNeighbors performs a BFS from a root node up to maxDepth hops and returns
// all reachable nodes and the edges traversed, scoped to the given org when non-nil.
func (r *Repository) GetTopologyNeighbors(ctx context.Context, orgID *int64, rootType string, rootID int64, maxDepth int) ([]*models.GraphNode, []*models.GraphEdge, error) {
	rootNodeID, err := r.getTopologyNodeID(ctx, rootType, rootID)
	if err != nil {
		return nil, nil, err
	}

	if maxDepth <= 0 {
		maxDepth = 3
	}

	// BFS using a recursive CTE. We traverse edges in both directions so the graph
	// is treated as undirected for reachability purposes.
	orgFilter := ``
	args := []any{rootNodeID, maxDepth}
	if orgID != nil {
		orgFilter = `WHERE n.organization_id = $3`
		args = append(args, *orgID)
	}

	query := fmt.Sprintf(`
		WITH RECURSIVE bfs AS (
		  SELECT id, 0 AS depth FROM topology_nodes WHERE id = $1
		  UNION
		  SELECT tn.id, bfs.depth + 1
		  FROM bfs
		  JOIN topology_edges te ON te.source_node_id = bfs.id OR te.target_node_id = bfs.id
		  JOIN topology_nodes tn ON tn.id = CASE WHEN te.source_node_id = bfs.id THEN te.target_node_id ELSE te.source_node_id END
		  WHERE bfs.depth < $2
		)
		SELECT DISTINCT n.id, n.organization_id, n.resource_type, n.resource_id, n.created_at
		FROM bfs
		JOIN topology_nodes n ON n.id = bfs.id
		%s
		ORDER BY n.resource_type, n.resource_id`, orgFilter)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	nodeIDs := map[int64]bool{}
	var nodes []*models.GraphNode
	for rows.Next() {
		n := &models.GraphNode{}
		if err := rows.Scan(&n.ID, &n.OrganizationID, &n.ResourceType, &n.ResourceID, &n.CreatedAt); err != nil {
			return nil, nil, err
		}
		nodes = append(nodes, n)
		nodeIDs[n.ID] = true
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	edges, err := r.edgesBetween(ctx, nodeIDs)
	return nodes, edges, err
}

// GetTopologyPath finds the shortest path between two resource nodes via BFS and returns
// the nodes and edges along the path.
func (r *Repository) GetTopologyPath(ctx context.Context, fromType string, fromID int64, toType string, toID int64) ([]*models.GraphNode, []*models.GraphEdge, error) {
	fromNodeID, err := r.getTopologyNodeID(ctx, fromType, fromID)
	if err != nil {
		return nil, nil, err
	}
	toNodeID, err := r.getTopologyNodeID(ctx, toType, toID)
	if err != nil {
		return nil, nil, err
	}

	if fromNodeID == toNodeID {
		n := &models.GraphNode{}
		if err := r.db.QueryRow(ctx,
			`SELECT id, organization_id, resource_type, resource_id, created_at FROM topology_nodes WHERE id = $1`, fromNodeID,
		).Scan(&n.ID, &n.OrganizationID, &n.ResourceType, &n.ResourceID, &n.CreatedAt); err != nil {
			return nil, nil, err
		}
		return []*models.GraphNode{n}, nil, nil
	}

	// BFS in Go to find the shortest path.
	type step struct {
		nodeID int64
		path   []int64
	}

	visited := map[int64]bool{fromNodeID: true}
	queue := []step{{fromNodeID, []int64{fromNodeID}}}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		// Fetch all edges touching cur.nodeID.
		rows, err := r.db.Query(ctx, `
			SELECT source_node_id, target_node_id FROM topology_edges
			WHERE source_node_id = $1 OR target_node_id = $1`, cur.nodeID)
		if err != nil {
			return nil, nil, err
		}

		var neighbors []int64
		for rows.Next() {
			var src, tgt int64
			if err := rows.Scan(&src, &tgt); err != nil {
				rows.Close()
				return nil, nil, err
			}
			next := tgt
			if tgt == cur.nodeID {
				next = src
			}
			neighbors = append(neighbors, next)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, nil, err
		}

		for _, next := range neighbors {
			if visited[next] {
				continue
			}
			newPath := append(append([]int64{}, cur.path...), next)
			if next == toNodeID {
				return r.pathNodesAndEdges(ctx, newPath)
			}
			visited[next] = true
			queue = append(queue, step{next, newPath})
		}
	}

	return nil, nil, fmt.Errorf("no path found between %s/%d and %s/%d", fromType, fromID, toType, toID)
}

// edgesBetween returns all edges where both endpoints are in the given node ID set.
func (r *Repository) edgesBetween(ctx context.Context, nodeIDs map[int64]bool) ([]*models.GraphEdge, error) {
	if len(nodeIDs) == 0 {
		return nil, nil
	}
	ids := make([]int64, 0, len(nodeIDs))
	for id := range nodeIDs {
		ids = append(ids, id)
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, organization_id, source_node_id, target_node_id, relationship, confidence, status, evidence, created_at
		FROM topology_edges
		WHERE source_node_id = ANY($1) AND target_node_id = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []*models.GraphEdge
	for rows.Next() {
		e := &models.GraphEdge{}
		if err := rows.Scan(&e.ID, &e.OrganizationID, &e.SourceNodeID, &e.TargetNodeID,
			&e.Relationship, &e.Confidence, &e.Status, &e.Evidence, &e.CreatedAt); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

// pathNodesAndEdges loads the full node and edge records for a BFS-discovered path.
func (r *Repository) pathNodesAndEdges(ctx context.Context, nodeIDPath []int64) ([]*models.GraphNode, []*models.GraphEdge, error) {
	nodeIDs := map[int64]bool{}
	for _, id := range nodeIDPath {
		nodeIDs[id] = true
	}

	// Load nodes in path order.
	nodes := make([]*models.GraphNode, 0, len(nodeIDPath))
	for _, id := range nodeIDPath {
		n := &models.GraphNode{}
		if err := r.db.QueryRow(ctx,
			`SELECT id, organization_id, resource_type, resource_id, created_at FROM topology_nodes WHERE id = $1`, id,
		).Scan(&n.ID, &n.OrganizationID, &n.ResourceType, &n.ResourceID, &n.CreatedAt); err != nil {
			return nil, nil, err
		}
		nodes = append(nodes, n)
	}

	edges, err := r.edgesBetween(ctx, nodeIDs)
	return nodes, edges, err
}
