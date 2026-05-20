package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"padduck/models"
)

// SplitSubnet atomically: creates child subnets with parent_subnet_id set,
// moves IPs to the correct child, and marks the parent as is_container=true.
func (r *Repository) SplitSubnet(ctx context.Context, parentID int64, childSubnets []*models.Subnet) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Create each child subnet
	for _, child := range childSubnets {
		var childID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO subnets (section_id, network_address, prefix_length, description, parent_subnet_id, is_container)
			VALUES ($1, $2::inet, $3, $4, $5, false)
			RETURNING id`,
			child.SectionID, child.NetworkAddress, child.PrefixLength, child.Description, parentID,
		).Scan(&childID)
		if err != nil {
			return fmt.Errorf("creating child subnet %s/%d: %w", child.NetworkAddress, child.PrefixLength, err)
		}
		child.ID = childID

		// Move IPs that fall within this child's CIDR
		_, err = tx.Exec(ctx, `
			UPDATE ip_addresses
			SET subnet_id = $1, updated_at = CURRENT_TIMESTAMP
			WHERE subnet_id = $2
			  AND address << ($3::text || '/' || $4::text)::inet`,
			childID, parentID, child.NetworkAddress, child.PrefixLength,
		)
		if err != nil {
			return fmt.Errorf("moving IPs to child subnet %s/%d: %w", child.NetworkAddress, child.PrefixLength, err)
		}
	}

	// Mark parent as container
	_, err = tx.Exec(ctx, `
		UPDATE subnets SET is_container = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
		parentID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// MergeSubnets atomically: creates (or updates) a parent subnet, moves all IPs to it,
// and deletes the merged children. Returns the parent subnet.
func (r *Repository) MergeSubnets(ctx context.Context, subnetIDs []int64, parent *models.Subnet) (*models.Subnet, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Create the parent subnet
	var parentID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO subnets (section_id, network_address, prefix_length, description, is_container)
		VALUES ($1, $2::inet, $3, $4, false)
		RETURNING id`,
		parent.SectionID, parent.NetworkAddress, parent.PrefixLength, parent.Description,
	).Scan(&parentID)
	if err != nil {
		return nil, fmt.Errorf("creating merged parent subnet: %w", err)
	}

	// Move all IPs from each child subnet to the parent
	for _, sid := range subnetIDs {
		_, err = tx.Exec(ctx, `
			UPDATE ip_addresses
			SET subnet_id = $1, updated_at = CURRENT_TIMESTAMP
			WHERE subnet_id = $2`,
			parentID, sid,
		)
		if err != nil {
			return nil, fmt.Errorf("moving IPs from subnet %d: %w", sid, err)
		}
	}

	// Delete merged children
	for _, sid := range subnetIDs {
		_, err = tx.Exec(ctx, `DELETE FROM subnets WHERE id = $1`, sid)
		if err != nil {
			return nil, fmt.Errorf("deleting merged subnet %d: %w", sid, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetSubnetByID(ctx, parentID)
}

// ResizeSubnet updates the network_address and prefix_length of a subnet.
func (r *Repository) ResizeSubnet(ctx context.Context, id int64, newNetworkAddr string, newPrefixLen int) (*models.Subnet, error) {
	_, err := r.db.Exec(ctx, `
		UPDATE subnets SET network_address = $1::inet, prefix_length = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3`,
		newNetworkAddr, newPrefixLen, id,
	)
	if err != nil {
		return nil, err
	}
	return r.GetSubnetByID(ctx, id)
}

// ListIPsOutsideCIDR returns IP addresses in a subnet whose address falls outside the given CIDR.
func (r *Repository) ListIPsOutsideCIDR(ctx context.Context, subnetID int64, networkAddr string, prefixLen int) ([]*models.IPAddress, error) {
	cidr := fmt.Sprintf("%s/%d", networkAddr, prefixLen)
	rows, err := r.db.Query(ctx, `
		SELECT `+ipSelectCols+` `+ipFromJoin+`
		WHERE ip.subnet_id = $1
		  AND NOT (ip.address << $2::inet)`,
		subnetID, cidr,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ips := make([]*models.IPAddress, 0)
	for rows.Next() {
		ip, err := scanIP(rows)
		if err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, rows.Err()
}

// ListSiblingSubnets returns subnets with the same section_id as the given subnet, excluding self.
func (r *Repository) ListSiblingSubnets(ctx context.Context, subnetID int64) ([]*models.Subnet, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+subnetSelectCols+` `+subnetFromJoin+`
		WHERE s.section_id = (SELECT section_id FROM subnets WHERE id = $1)
		  AND s.id != $1
		ORDER BY s.network_address`,
		subnetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subnets := make([]*models.Subnet, 0)
	for rows.Next() {
		subnet, err := scanSubnet(rows)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, rows.Err()
}

// GetSectionTopology returns topology node and edge data for all subnets in a section.
func (r *Repository) GetSectionTopology(ctx context.Context, sectionID int64) (*models.SectionTopology, error) {
	rows, err := r.db.Query(ctx, `
		SELECT s.id, s.network_address::text, s.prefix_length, s.description,
		       s.is_container, s.parent_subnet_id, s.vlan_id,
		       COUNT(ip.id) FILTER (WHERE ip.status != 'available') AS used_count,
		       COUNT(ip.id) AS total_count
		FROM subnets s
		LEFT JOIN ip_addresses ip ON ip.subnet_id = s.id
		WHERE s.section_id = $1
		GROUP BY s.id
		ORDER BY s.network_address`,
		sectionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	topology := &models.SectionTopology{
		Nodes: make([]*models.TopologyNode, 0),
		Edges: make([]*models.TopologyEdge, 0),
	}

	for rows.Next() {
		var id int64
		var networkAddr string
		var prefixLen int
		var description string
		var isContainer bool
		var parentSubnetID *int64
		var vlanID *int64
		var usedCount, totalCount int64

		if err := rows.Scan(&id, &networkAddr, &prefixLen, &description,
			&isContainer, &parentSubnetID, &vlanID,
			&usedCount, &totalCount); err != nil {
			return nil, err
		}

		utilisation := 0.0
		if totalCount > 0 {
			utilisation = float64(usedCount) / float64(totalCount) * 100.0
		}

		cidr := fmt.Sprintf("%s/%d", networkAddr, prefixLen)
		node := &models.TopologyNode{
			ID:          id,
			Label:       cidr,
			CIDR:        cidr,
			PrefixLen:   prefixLen,
			IsContainer: isContainer,
			ParentID:    parentSubnetID,
			VLANID:      vlanID,
			Utilisation: utilisation,
		}
		topology.Nodes = append(topology.Nodes, node)

		// Add parent->child edge
		if parentSubnetID != nil {
			topology.Edges = append(topology.Edges, &models.TopologyEdge{
				Source: *parentSubnetID,
				Target: id,
				Type:   "parent_child",
			})
		}

		// Add subnet->vlan edge
		if vlanID != nil {
			topology.Edges = append(topology.Edges, &models.TopologyEdge{
				Source: id,
				Target: *vlanID,
				Type:   "subnet_vlan",
			})
		}
	}

	return topology, rows.Err()
}

// ---------------------------------------------------------------------------
// IPv6 Delegation repository methods
// ---------------------------------------------------------------------------

func scanIPv6Delegation(row interface{ Scan(dest ...any) error }) (*models.IPv6Delegation, error) {
	d := &models.IPv6Delegation{}
	return d, row.Scan(
		&d.ID, &d.ParentSubnetID, &d.DelegatedPrefix,
		&d.DelegatedToDeviceID, &d.DelegatedToDescription,
		&d.ValidLifetimeSec, &d.PreferredLifetimeSec,
		&d.ExpiresAt, &d.CreatedAt,
	)
}

// CreateIPv6Delegation inserts a new IPv6 delegation record.
func (r *Repository) CreateIPv6Delegation(ctx context.Context, d *models.IPv6Delegation) (*models.IPv6Delegation, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO ipv6_delegations
		  (parent_subnet_id, delegated_prefix, delegated_to_device_id, delegated_to_description,
		   valid_lifetime_sec, preferred_lifetime_sec, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		d.ParentSubnetID, d.DelegatedPrefix, d.DelegatedToDeviceID, d.DelegatedToDescription,
		d.ValidLifetimeSec, d.PreferredLifetimeSec, d.ExpiresAt,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetIPv6DelegationByID(ctx, id)
}

// GetIPv6DelegationByID returns a single delegation by ID.
func (r *Repository) GetIPv6DelegationByID(ctx context.Context, id int64) (*models.IPv6Delegation, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, parent_subnet_id, delegated_prefix, delegated_to_device_id, delegated_to_description,
		       valid_lifetime_sec, preferred_lifetime_sec, expires_at, created_at
		FROM ipv6_delegations WHERE id = $1`,
		id,
	)
	return scanIPv6Delegation(row)
}

// ListIPv6DelegationsBySubnet returns all delegations for a given parent subnet.
func (r *Repository) ListIPv6DelegationsBySubnet(ctx context.Context, subnetID int64) ([]*models.IPv6Delegation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, parent_subnet_id, delegated_prefix, delegated_to_device_id, delegated_to_description,
		       valid_lifetime_sec, preferred_lifetime_sec, expires_at, created_at
		FROM ipv6_delegations WHERE parent_subnet_id = $1 ORDER BY id`,
		subnetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.IPv6Delegation, 0)
	for rows.Next() {
		d, err := scanIPv6Delegation(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

// UpdateIPv6Delegation updates mutable fields of a delegation record.
func (r *Repository) UpdateIPv6Delegation(ctx context.Context, id int64, d *models.IPv6Delegation) (*models.IPv6Delegation, error) {
	_, err := r.db.Exec(ctx, `
		UPDATE ipv6_delegations
		SET delegated_prefix = $1,
		    delegated_to_device_id = $2,
		    delegated_to_description = $3,
		    valid_lifetime_sec = $4,
		    preferred_lifetime_sec = $5,
		    expires_at = $6
		WHERE id = $7`,
		d.DelegatedPrefix, d.DelegatedToDeviceID, d.DelegatedToDescription,
		d.ValidLifetimeSec, d.PreferredLifetimeSec, d.ExpiresAt, id,
	)
	if err != nil {
		return nil, err
	}
	return r.GetIPv6DelegationByID(ctx, id)
}

// DeleteIPv6Delegation removes a delegation record.
func (r *Repository) DeleteIPv6Delegation(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM ipv6_delegations WHERE id = $1`, id)
	return err
}

// computeIsExpired returns true if the delegation's ExpiresAt is set and in the past.
func computeIsExpired(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return false
	}
	return time.Now().After(*expiresAt)
}
