package repository

import (
	"context"

	"padduck/models"
)

// VLAN operations

func (r *Repository) CreateVLAN(ctx context.Context, vrfID *int64, domainID *int64, groupID *int64, vlanID int, name, description string) (*models.VLAN, error) {
	query := `INSERT INTO vlans (vrf_id, domain_id, group_id, vlan_id, name, description)
	          VALUES ($1, $2, $3, $4, $5, $6)
	          RETURNING id, vrf_id, domain_id, group_id, vlan_id, name, description, created_at, updated_at`
	vlan := &models.VLAN{}
	err := r.db.QueryRow(ctx, query, vrfID, domainID, groupID, vlanID, name, description).Scan(
		&vlan.ID, &vlan.VRFID, &vlan.DomainID, &vlan.GroupID, &vlan.VlanID, &vlan.Name, &vlan.Description, &vlan.CreatedAt, &vlan.UpdatedAt,
	)
	return vlan, err
}

func (r *Repository) GetVLANByID(ctx context.Context, id int64) (*models.VLAN, error) {
	query := `SELECT id, vrf_id, domain_id, group_id, vlan_id, name, description, created_at, updated_at FROM vlans WHERE id = $1`
	vlan := &models.VLAN{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&vlan.ID, &vlan.VRFID, &vlan.DomainID, &vlan.GroupID, &vlan.VlanID, &vlan.Name, &vlan.Description, &vlan.CreatedAt, &vlan.UpdatedAt,
	)
	return vlan, err
}

func (r *Repository) ListAllVLANs(ctx context.Context) ([]*models.VLAN, error) {
	query := `SELECT id, vrf_id, domain_id, group_id, vlan_id, name, description, created_at, updated_at FROM vlans ORDER BY vlan_id ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vlans := make([]*models.VLAN, 0)
	for rows.Next() {
		vlan := &models.VLAN{}
		err := rows.Scan(&vlan.ID, &vlan.VRFID, &vlan.DomainID, &vlan.GroupID, &vlan.VlanID, &vlan.Name, &vlan.Description, &vlan.CreatedAt, &vlan.UpdatedAt)
		if err != nil {
			return nil, err
		}
		vlans = append(vlans, vlan)
	}
	return vlans, rows.Err()
}

func (r *Repository) ListVLANsByVRF(ctx context.Context, vrfID int64) ([]*models.VLAN, error) {
	query := `SELECT id, vrf_id, domain_id, group_id, vlan_id, name, description, created_at, updated_at FROM vlans WHERE vrf_id = $1 ORDER BY vlan_id ASC`
	rows, err := r.db.Query(ctx, query, vrfID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vlans := make([]*models.VLAN, 0)
	for rows.Next() {
		vlan := &models.VLAN{}
		err := rows.Scan(&vlan.ID, &vlan.VRFID, &vlan.DomainID, &vlan.GroupID, &vlan.VlanID, &vlan.Name, &vlan.Description, &vlan.CreatedAt, &vlan.UpdatedAt)
		if err != nil {
			return nil, err
		}
		vlans = append(vlans, vlan)
	}
	return vlans, rows.Err()
}

// ListVLANsPaginated returns a page of VLANs with a total count.
func (r *Repository) ListVLANsPaginated(ctx context.Context, limit, offset int) ([]*models.VLAN, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM vlans`).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, vrf_id, domain_id, group_id, vlan_id, name, description, created_at, updated_at FROM vlans ORDER BY vlan_id ASC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	vlans := make([]*models.VLAN, 0)
	for rows.Next() {
		vlan := &models.VLAN{}
		if err := rows.Scan(&vlan.ID, &vlan.VRFID, &vlan.DomainID, &vlan.GroupID, &vlan.VlanID, &vlan.Name, &vlan.Description, &vlan.CreatedAt, &vlan.UpdatedAt); err != nil {
			return nil, 0, err
		}
		vlans = append(vlans, vlan)
	}
	return vlans, total, rows.Err()
}

func (r *Repository) UpdateVLAN(ctx context.Context, id int64, domainID *int64, groupID *int64, name, description string) (*models.VLAN, error) {
	query := `UPDATE vlans SET name = $1, description = $2, domain_id = $3, group_id = $4, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $5
	          RETURNING id, vrf_id, domain_id, group_id, vlan_id, name, description, created_at, updated_at`
	vlan := &models.VLAN{}
	err := r.db.QueryRow(ctx, query, name, description, domainID, groupID, id).Scan(
		&vlan.ID, &vlan.VRFID, &vlan.DomainID, &vlan.GroupID, &vlan.VlanID, &vlan.Name, &vlan.Description, &vlan.CreatedAt, &vlan.UpdatedAt,
	)
	return vlan, err
}

func (r *Repository) DeleteVLAN(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM vlans WHERE id = $1`, id)
	return err
}

// GetVLANSubnets returns all subnets assigned to a VLAN.
func (r *Repository) GetVLANSubnets(ctx context.Context, vlanID int64) ([]*models.Subnet, error) {
	query := `SELECT ` + subnetSelectCols + ` ` + subnetFromJoin + ` WHERE s.vlan_id = $1 ORDER BY s.network_address`
	rows, err := r.db.Query(ctx, query, vlanID)
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

// VLANDomain operations

func (r *Repository) CreateVLANDomain(ctx context.Context, name string, description *string) (*models.VLANDomain, error) {
	query := `INSERT INTO vlan_domains (name, description)
	          VALUES ($1, $2)
	          RETURNING id, name, description, created_at, updated_at`
	d := &models.VLANDomain{}
	err := r.db.QueryRow(ctx, query, name, description).Scan(
		&d.ID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt,
	)
	return d, err
}

func (r *Repository) GetVLANDomainByID(ctx context.Context, id int64) (*models.VLANDomain, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM vlan_domains WHERE id = $1`
	d := &models.VLANDomain{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt,
	)
	return d, err
}

func (r *Repository) ListVLANDomains(ctx context.Context) ([]*models.VLANDomain, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM vlan_domains ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	domains := make([]*models.VLANDomain, 0)
	for rows.Next() {
		d := &models.VLANDomain{}
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

func (r *Repository) UpdateVLANDomain(ctx context.Context, id int64, name string, description *string) (*models.VLANDomain, error) {
	query := `UPDATE vlan_domains SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $3
	          RETURNING id, name, description, created_at, updated_at`
	d := &models.VLANDomain{}
	err := r.db.QueryRow(ctx, query, name, description, id).Scan(
		&d.ID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt,
	)
	return d, err
}

func (r *Repository) DeleteVLANDomain(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM vlan_domains WHERE id = $1`, id)
	return err
}

// VLANGroup operations

func (r *Repository) CreateVLANGroup(ctx context.Context, name string, description *string, colour *string) (*models.VLANGroup, error) {
	query := `INSERT INTO vlan_groups (name, description, colour)
	          VALUES ($1, $2, $3)
	          RETURNING id, name, description, colour, created_at, updated_at`
	g := &models.VLANGroup{}
	err := r.db.QueryRow(ctx, query, name, description, colour).Scan(
		&g.ID, &g.Name, &g.Description, &g.Colour, &g.CreatedAt, &g.UpdatedAt,
	)
	return g, err
}

func (r *Repository) GetVLANGroupByID(ctx context.Context, id int64) (*models.VLANGroup, error) {
	query := `SELECT id, name, description, colour, created_at, updated_at FROM vlan_groups WHERE id = $1`
	g := &models.VLANGroup{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&g.ID, &g.Name, &g.Description, &g.Colour, &g.CreatedAt, &g.UpdatedAt,
	)
	return g, err
}

func (r *Repository) ListVLANGroups(ctx context.Context) ([]*models.VLANGroup, error) {
	query := `SELECT id, name, description, colour, created_at, updated_at FROM vlan_groups ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]*models.VLANGroup, 0)
	for rows.Next() {
		g := &models.VLANGroup{}
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.Colour, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *Repository) UpdateVLANGroup(ctx context.Context, id int64, name string, description *string, colour *string) (*models.VLANGroup, error) {
	query := `UPDATE vlan_groups SET name = $1, description = $2, colour = $3, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $4
	          RETURNING id, name, description, colour, created_at, updated_at`
	g := &models.VLANGroup{}
	err := r.db.QueryRow(ctx, query, name, description, colour, id).Scan(
		&g.ID, &g.Name, &g.Description, &g.Colour, &g.CreatedAt, &g.UpdatedAt,
	)
	return g, err
}

func (r *Repository) DeleteVLANGroup(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM vlan_groups WHERE id = $1`, id)
	return err
}

// GetVLANUsageReport returns per-VLAN metrics by joining vlans → subnets → ip_addresses.
func (r *Repository) GetVLANUsageReport(ctx context.Context) ([]*models.VLANUsageEntry, error) {
	query := `
SELECT
    v.id                                          AS vlan_id,
    v.name                                        AS vlan_name,
    v.vlan_id                                     AS vlan_tag,
    COUNT(DISTINCT s.id)                          AS subnet_count,
    COUNT(ip.id)                                  AS ip_count,
    COALESCE(SUM(CASE
        WHEN s.prefix_length IS NOT NULL
        THEN POWER(2, 32 - s.prefix_length)::BIGINT
        ELSE 0 END), 0)                           AS total_ips,
    CASE WHEN COALESCE(SUM(CASE
        WHEN s.prefix_length IS NOT NULL
        THEN POWER(2, 32 - s.prefix_length)::BIGINT
        ELSE 0 END), 0) = 0
        THEN 0.0
        ELSE ROUND(
            COUNT(ip.id)::NUMERIC /
            SUM(POWER(2, 32 - s.prefix_length)::BIGINT)::NUMERIC * 100, 2
        )
    END                                           AS utilisation_pct
FROM vlans v
LEFT JOIN subnets s ON s.vlan_id = v.id
LEFT JOIN ip_addresses ip ON ip.subnet_id = s.id
GROUP BY v.id, v.name, v.vlan_id
ORDER BY v.vlan_id ASC
`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]*models.VLANUsageEntry, 0)
	for rows.Next() {
		e := &models.VLANUsageEntry{}
		if err := rows.Scan(&e.VLANID, &e.VLANName, &e.VLANTag, &e.SubnetCount, &e.IPCount, &e.TotalIPs, &e.UtilisationPct); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
