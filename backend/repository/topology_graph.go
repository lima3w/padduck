package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
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
		  JOIN topology_edges te ON (te.source_node_id = bfs.id OR te.target_node_id = bfs.id) AND te.status <> 'dismissed'
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
			WHERE (source_node_id = $1 OR target_node_id = $1) AND status <> 'dismissed'`, cur.nodeID)
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
		WHERE source_node_id = ANY($1) AND target_node_id = ANY($1) AND status <> 'dismissed'`, ids)
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

// autoEdgesCTE selects every relationship that can be derived from inventory
// columns. Derived edges carry evidence 'auto' and are rebuilt on each sync.
const autoEdgesCTE = `
	derived(src_type, src_id, tgt_type, tgt_id, rel) AS (
		SELECT 'subnet', parent_subnet_id, 'subnet', id, 'contains' FROM subnets WHERE parent_subnet_id IS NOT NULL
		UNION
		SELECT 'subnet', id, 'vlan', vlan_id, 'member_of' FROM subnets WHERE vlan_id IS NOT NULL
		UNION
		SELECT 'subnet', id, 'location', location_id, 'member_of' FROM subnets WHERE location_id IS NOT NULL
		UNION
		SELECT 'device', device_id, 'subnet', subnet_id, 'member_of' FROM ip_addresses
			WHERE device_id IS NOT NULL AND subnet_id IS NOT NULL
		UNION
		SELECT 'device', id, 'location', location_id, 'member_of' FROM devices WHERE location_id IS NOT NULL
		UNION
		SELECT 'location', parent_id, 'location', id, 'contains' FROM locations WHERE parent_id IS NOT NULL
	)`

// SyncTopologyGraph reconciles the graph with inventory: it registers a node
// for every subnet, device, VLAN and location, removes nodes whose resource is
// gone, and rebuilds the edges derived from inventory relationships. Edges
// that were suggested or created by hand are never touched.
func (r *Repository) SyncTopologyGraph(ctx context.Context) error {
	for _, t := range []struct{ typ, table string }{
		{"subnet", "subnets"}, {"device", "devices"}, {"vlan", "vlans"}, {"location", "locations"},
	} {
		if _, err := r.db.Exec(ctx, fmt.Sprintf(
			`DELETE FROM topology_nodes WHERE resource_type=$1 AND resource_id NOT IN (SELECT id FROM %s)`, t.table), t.typ); err != nil {
			return err
		}
		if _, err := r.db.Exec(ctx, fmt.Sprintf(
			`INSERT INTO topology_nodes (resource_type, resource_id) SELECT $1, id FROM %s ON CONFLICT DO NOTHING`, t.table), t.typ); err != nil {
			return err
		}
	}
	if _, err := r.db.Exec(ctx, `
		WITH`+autoEdgesCTE+`
		DELETE FROM topology_edges e USING topology_nodes s, topology_nodes t
		WHERE e.source_node_id=s.id AND e.target_node_id=t.id AND e.evidence='auto'
		  AND NOT EXISTS (SELECT 1 FROM derived d WHERE d.src_type=s.resource_type AND d.src_id=s.resource_id
		                  AND d.tgt_type=t.resource_type AND d.tgt_id=t.resource_id AND d.rel=e.relationship)`); err != nil {
		return err
	}
	_, err := r.db.Exec(ctx, `
		WITH`+autoEdgesCTE+`
		INSERT INTO topology_edges (source_node_id, target_node_id, relationship, confidence, status, evidence)
		SELECT s.id, t.id, d.rel, 1.0, 'confirmed', 'auto'
		FROM derived d
		JOIN topology_nodes s ON s.resource_type=d.src_type AND s.resource_id=d.src_id
		JOIN topology_nodes t ON t.resource_type=d.tgt_type AND t.resource_id=d.tgt_id
		ON CONFLICT (source_node_id, target_node_id, relationship) DO NOTHING`)
	return err
}

// SuggestTopologyEdge records a discovered relationship as a suggested edge.
// An existing edge (confirmed, or dismissed by an admin) is left as it is.
// Any other open suggestion from the same source with the same evidence prefix
// and relationship is dropped so a moved uplink does not leave a stale one.
func (r *Repository) SuggestTopologyEdge(ctx context.Context, srcType string, srcID int64, tgtType string, tgtID int64, relationship string, confidence float64, evidence, evidencePrefix string) error {
	srcNode, err := r.UpsertTopologyNode(ctx, nil, srcType, srcID)
	if err != nil {
		return err
	}
	tgtNode, err := r.UpsertTopologyNode(ctx, nil, tgtType, tgtID)
	if err != nil {
		return err
	}
	if _, err := r.db.Exec(ctx, `
		DELETE FROM topology_edges
		WHERE source_node_id=$1 AND relationship=$2 AND status='suggested'
		  AND evidence LIKE $3 AND target_node_id<>$4`,
		srcNode.ID, relationship, evidencePrefix+"%", tgtNode.ID); err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO topology_edges (source_node_id, target_node_id, relationship, confidence, status, evidence)
		VALUES ($1,$2,$3,$4,'suggested',$5)
		ON CONFLICT (source_node_id, target_node_id, relationship) DO NOTHING`,
		srcNode.ID, tgtNode.ID, relationship, confidence, evidence)
	return err
}

// ListTopologyEdges returns edges (with endpoint resource refs) for one status,
// or all non-dismissed edges when status is empty.
func (r *Repository) ListTopologyEdges(ctx context.Context, status string) ([]*models.GraphEdgeDetail, error) {
	q := `
		SELECT e.id, e.relationship, e.confidence, e.status, e.evidence,
		       s.resource_type, s.resource_id, t.resource_type, t.resource_id
		FROM topology_edges e
		JOIN topology_nodes s ON s.id=e.source_node_id
		JOIN topology_nodes t ON t.id=e.target_node_id`
	args := []any{}
	if status != "" {
		q += ` WHERE e.status=$1`
		args = append(args, status)
	}
	rows, err := r.db.Query(ctx, q+` ORDER BY e.id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*models.GraphEdgeDetail{}
	for rows.Next() {
		e := &models.GraphEdgeDetail{}
		if err := rows.Scan(&e.ID, &e.Relationship, &e.Confidence, &e.Status, &e.Evidence,
			&e.SourceType, &e.SourceID, &e.TargetType, &e.TargetID); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// SetTopologyEdgeStatus confirms or dismisses an edge.
func (r *Repository) SetTopologyEdgeStatus(ctx context.Context, id int64, status string) error {
	tag, err := r.db.Exec(ctx, `UPDATE topology_edges SET status=$2 WHERE id=$1`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// GetTopologyFull returns every node and every non-dismissed edge, capped at limit nodes.
func (r *Repository) GetTopologyFull(ctx context.Context, limit int) ([]*models.GraphNode, []*models.GraphEdge, error) {
	rows, err := r.db.Query(ctx, `SELECT id, organization_id, resource_type, resource_id, created_at FROM topology_nodes ORDER BY id LIMIT $1`, limit)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	ids := map[int64]bool{}
	var nodes []*models.GraphNode
	for rows.Next() {
		n := &models.GraphNode{}
		if err := rows.Scan(&n.ID, &n.OrganizationID, &n.ResourceType, &n.ResourceID, &n.CreatedAt); err != nil {
			return nil, nil, err
		}
		nodes = append(nodes, n)
		ids[n.ID] = true
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	edges, err := r.edgesBetween(ctx, ids)
	return nodes, edges, err
}

// TopologyNodeLabels returns a display label and detail line for each node,
// keyed "type_id".
func (r *Repository) TopologyNodeLabels(ctx context.Context, nodes []*models.GraphNode) (map[string][2]string, error) {
	out := map[string][2]string{}
	queries := map[string]string{
		"subnet":   `SELECT id, host(network_address)||'/'||prefix_length, COALESCE(description,'') FROM subnets WHERE id=ANY($1)`,
		"device":   `SELECT id, hostname, COALESCE(vendor,'')||' '||COALESCE(model,'') FROM devices WHERE id=ANY($1)`,
		"vlan":     `SELECT id, 'VLAN '||vlan_id, name FROM vlans WHERE id=ANY($1)`,
		"location": `SELECT id, name, COALESCE(type,'') FROM locations WHERE id=ANY($1)`,
	}
	byType := map[string][]int64{}
	for _, n := range nodes {
		byType[n.ResourceType] = append(byType[n.ResourceType], n.ResourceID)
	}
	for typ, ids := range byType {
		q, ok := queries[typ]
		if !ok {
			continue
		}
		rows, err := r.db.Query(ctx, q, ids)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id int64
			var label, detail string
			if err := rows.Scan(&id, &label, &detail); err != nil {
				rows.Close()
				return nil, err
			}
			out[fmt.Sprintf("%s_%d", typ, id)] = [2]string{label, strings.TrimSpace(detail)}
		}
		rows.Close()
	}
	return out, nil
}
